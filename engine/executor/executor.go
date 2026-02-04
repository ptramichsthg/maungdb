package executor

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/febrd/maungdb/engine/auth"
	"github.com/febrd/maungdb/engine/parser"
	"github.com/febrd/maungdb/engine/schema"
	"github.com/febrd/maungdb/engine/storage"
)

type ExecutionResult struct {
	Columns []string
	Rows    [][]string
	Message string
}

func Execute(cmd *parser.Command) (*ExecutionResult, error) {
	switch cmd.Type {
	case parser.CmdCreate:
		return execCreate(cmd)
	case parser.CmdInsert:
		return execInsert(cmd)
	case parser.CmdSelect:
		return execSelect(cmd)
	case parser.CmdUpdate:
		return execUpdate(cmd)
	case parser.CmdDelete:
		return execDelete(cmd)
	default:
		return nil, errors.New("command teu didukung")
	}
}

// ==========================================
// CREATE & INSERT
// ==========================================

func execCreate(cmd *parser.Command) (*ExecutionResult, error) {
	user, _ := auth.CurrentUser()
	fields := splitColumns(cmd.Data)

	perms := map[string][]string{
		"read":  {"user", "admin", "supermaung"},
		"write": {"admin", "supermaung"},
	}

	if err := schema.Create(user.Database, cmd.Table, fields, perms); err != nil {
		return nil, err
	}

	return &ExecutionResult{Message: fmt.Sprintf("✅ Tabel '%s' parantos didamel!", cmd.Table)}, nil
}

// Helper untuk memisahkan kolom CREATE TABLE, menangani koma dalam kurung ENUM(A,B)
func splitColumns(input string) []string {
	var fields []string
	var currentField strings.Builder
	parenCount := 0

	for _, char := range input {
		switch char {
		case '(':
			parenCount++
			currentField.WriteRune(char)
		case ')':
			parenCount--
			currentField.WriteRune(char)
		case ',':
			if parenCount == 0 {
				fields = append(fields, strings.TrimSpace(currentField.String()))
				currentField.Reset()
			} else {
				currentField.WriteRune(char)
			}
		default:
			currentField.WriteRune(char)
		}
	}

	if currentField.Len() > 0 {
		fields = append(fields, strings.TrimSpace(currentField.String()))
	}

	return fields
}

func execInsert(cmd *parser.Command) (*ExecutionResult, error) {
	user, _ := auth.CurrentUser()

	s, err := schema.Load(user.Database, cmd.Table)
	if err != nil {
		return nil, err
	}

	if !s.Can(user.Role, "write") {
		return nil, errors.New("teu boga hak nulis")
	}

	if err := s.ValidateRow(cmd.Data); err != nil {
		return nil, err
	}

	if err := storage.Append(cmd.Table, cmd.Data); err != nil {
		return nil, err
	}

	return &ExecutionResult{
		Message: fmt.Sprintf("✅ Data asup ka table '%s'", cmd.Table),
	}, nil
}

// ==========================================
// SELECT (Full Features: Filter, Sort, Limit, Agregasi)
// ==========================================
func execSelect(cmd *parser.Command) (*ExecutionResult, error) {
	// 1. SETUP & SECURITY
	user, _ := auth.CurrentUser()

	// Load Schema
	s, err := schema.Load(user.Database, cmd.Table)
	if err != nil {
		return nil, fmt.Errorf("gagal maca schema tabel '%s': %v", cmd.Table, err)
	}

	// Cek Hak Akses
	if !s.Can(user.Role, "read") {
		return nil, errors.New("akses ditolak: anjeun teu boga hak maca tabel ieu")
	}

	// Baca Data Raw
	rawRows, err := storage.ReadAll(cmd.Table)
	if err != nil {
		return nil, err
	}

	if len(rawRows) == 0 {
		return &ExecutionResult{Message: "Data kosong", Columns: s.GetFieldNames()}, nil
	}

	fieldNames := s.GetFieldNames()
	var filteredMaps []map[string]string // Data pikeun Agregasi (Map)
	var filteredSlices [][]string        // Data pikeun Sorting/Limit (Slice)

	// 2. FILTERING DATA
	for _, raw := range rawRows {
		if raw == "" { continue }
		cols := strings.Split(raw, "|")

		// --- LOGIKA FILTER WHERE (FIXED) ---
		matches := true
		if len(cmd.Where) > 0 {
			// Evaluasi kondisi pertama
			matches = evaluateOne(cols, s.Columns, cmd.Where[0])

			// Loop kondisi berikutnya (AND/OR)
			// Loop sampai len-1 karena kita cek i+1 di dalam
			for i := 0; i < len(cmd.Where)-1; i++ {
				cond := cmd.Where[i]
				if cond.LogicOp == "" { break }

				nextResult := evaluateOne(cols, s.Columns, cmd.Where[i+1])

				op := strings.ToUpper(cond.LogicOp)
				if op == "SARENG" || op == "AND" {
					matches = matches && nextResult
				} else if op == "ATAWA" || op == "OR" {
					matches = matches || nextResult
				}
			}
		}

		if !matches { continue } // Skip mun teu cocok

		// Lolos filter -> Simpen ke Slice (untuk fitur lama)
		filteredSlices = append(filteredSlices, cols)

		// Simpen ke Map (untuk fitur agregasi baru)
		rowMap := make(map[string]string)
		for i, val := range cols {
			if i < len(fieldNames) {
				rowMap[fieldNames[i]] = val
			}
		}
		filteredMaps = append(filteredMaps, rowMap)
	}

	// 3. PROSES AGREGASI ATAU PROYEKSI KOLOM
	selectedFields := cmd.Fields
	if len(selectedFields) == 0 || selectedFields[0] == "*" {
		selectedFields = fieldNames
	}

	isAggregateQuery := false
	var parsedCols []ParsedColumn

	for _, f := range selectedFields {
		pc := ParseColumnSelection(f) // Fungsi ti aggregator.go
		parsedCols = append(parsedCols, pc)
		if pc.IsAggregate {
			isAggregateQuery = true
		}
	}

	var finalResult [][]string
	var finalHeader []string

	// === CABANG A: QUERY AGREGASI (COUNT, SUM, AVG, dll) ===
	if isAggregateQuery {
		var resultRow []string

		for _, pc := range parsedCols {
			finalHeader = append(finalHeader, pc.OriginalText)

			if pc.IsAggregate {
				// Hitung Matematika pake data Map
				val, _ := CalculateAggregate(filteredMaps, pc)
				resultRow = append(resultRow, val)
			} else {
				// Mun Select biasa campur Agregasi (Implicit Group By - ambil baris pertama)
				if len(filteredMaps) > 0 {
					resultRow = append(resultRow, filteredMaps[0][pc.TargetCol])
				} else {
					resultRow = append(resultRow, "-")
				}
			}
		}
		finalResult = append(finalResult, resultRow)

	} else {
		// === CABANG B: QUERY BIASA (SORT & LIMIT) ===

		// 1. Sorting (RUNTUYKEUN)
		if cmd.OrderBy != "" {
			colIdx := indexOf(cmd.OrderBy, fieldNames)
			if colIdx != -1 {
				colType := s.Columns[colIdx].Type // Cek tipe data schema

				sort.Slice(filteredSlices, func(i, j int) bool {
					valA := filteredSlices[i][colIdx]
					valB := filteredSlices[j][colIdx]
					isLess := false

					switch colType {
					case "INT":
						a, _ := strconv.Atoi(valA)
						b, _ := strconv.Atoi(valB)
						isLess = a < b
					case "FLOAT":
						a, _ := strconv.ParseFloat(valA, 64)
						b, _ := strconv.ParseFloat(valB, 64)
						isLess = a < b
					default:
						isLess = valA < valB
					}

					if cmd.OrderDesc {
						return !isLess
					}
					return isLess
				})
			}
		}

		// 2. Pagination (SAKADAR & LIWATAN)
		totalRows := len(filteredSlices)
		start := 0
		end := totalRows

		if cmd.Offset > 0 {
			start = cmd.Offset
			if start > totalRows { start = totalRows }
		}

		if cmd.Limit > 0 {
			end = start + cmd.Limit
			if end > totalRows { end = totalRows }
		}

		slicedRows := filteredSlices[start:end]

		// 3. Proyeksi Kolom (Pilih kolom nu dipenta wungkul)
		for _, pc := range parsedCols {
			finalHeader = append(finalHeader, pc.TargetCol)
		}

		for _, rowSlice := range slicedRows {
			var rowData []string
			for _, pc := range parsedCols {
				idx := indexOf(pc.TargetCol, fieldNames)
				if idx != -1 && idx < len(rowSlice) {
					rowData = append(rowData, rowSlice[idx])
				} else {
					rowData = append(rowData, "NULL")
				}
			}
			finalResult = append(finalResult, rowData)
		}
	}

	return &ExecutionResult{
		Columns: finalHeader,
		Rows:    finalResult,
		Message: fmt.Sprintf("%d baris kapanggih", len(finalResult)),
	}, nil
}

// ==========================================
// UPDATE (OMEAN) - Fixed Logic Filter
// ==========================================
func execUpdate(cmd *parser.Command) (*ExecutionResult, error) {
	user, _ := auth.CurrentUser()
	s, err := schema.Load(user.Database, cmd.Table)
	if err != nil {
		return nil, err
	}
	if !s.Can(user.Role, "write") {
		return nil, errors.New("teu boga hak nulis (omean)")
	}

	rawRows, err := storage.ReadAll(cmd.Table)
	if err != nil {
		return nil, err
	}

	var newRows []string
	updatedCount := 0

	for _, raw := range rawRows {
		if raw == "" { continue }
		cols := strings.Split(raw, "|")

		// --- LOGIKA FILTER (Sama dengan execSelect) ---
		shouldUpdate := true
		if len(cmd.Where) > 0 {
			shouldUpdate = evaluateOne(cols, s.Columns, cmd.Where[0])
			for i := 0; i < len(cmd.Where)-1; i++ {
				cond := cmd.Where[i]
				if cond.LogicOp == "" { break }
				nextResult := evaluateOne(cols, s.Columns, cmd.Where[i+1])
				op := strings.ToUpper(cond.LogicOp)
				if op == "SARENG" || op == "AND" {
					shouldUpdate = shouldUpdate && nextResult
				} else if op == "ATAWA" || op == "OR" {
					shouldUpdate = shouldUpdate || nextResult
				}
			}
		}

		if shouldUpdate {
			for colName, newVal := range cmd.Updates {
				idx := indexOf(colName, s.GetFieldNames())
				if idx != -1 {
					cols[idx] = newVal
				}
			}
			updatedCount++
		}

		newRows = append(newRows, strings.Join(cols, "|"))
	}

	if err := storage.Rewrite(cmd.Table, newRows); err != nil {
		return nil, err
	}

	return &ExecutionResult{Message: fmt.Sprintf("✅ %d data geus diomean", updatedCount)}, nil
}

// ==========================================
// DELETE (MICEUN) - Fixed Logic Filter
// ==========================================
func execDelete(cmd *parser.Command) (*ExecutionResult, error) {
	user, _ := auth.CurrentUser()
	s, err := schema.Load(user.Database, cmd.Table)
	if err != nil {
		return nil, err
	}
	if !s.Can(user.Role, "write") {
		return nil, errors.New("teu boga hak nulis (miceun)")
	}

	rawRows, err := storage.ReadAll(cmd.Table)
	if err != nil {
		return nil, err
	}

	var newRows []string
	deletedCount := 0

	for _, raw := range rawRows {
		if raw == "" { continue }
		cols := strings.Split(raw, "|")

		// --- LOGIKA FILTER (Sama dengan execSelect) ---
		shouldDelete := true
		if len(cmd.Where) > 0 {
			shouldDelete = evaluateOne(cols, s.Columns, cmd.Where[0])
			for i := 0; i < len(cmd.Where)-1; i++ {
				cond := cmd.Where[i]
				if cond.LogicOp == "" { break }
				nextResult := evaluateOne(cols, s.Columns, cmd.Where[i+1])
				op := strings.ToUpper(cond.LogicOp)
				if op == "SARENG" || op == "AND" {
					shouldDelete = shouldDelete && nextResult
				} else if op == "ATAWA" || op == "OR" {
					shouldDelete = shouldDelete || nextResult
				}
			}
		}

		if shouldDelete {
			deletedCount++
			continue
		}

		newRows = append(newRows, raw)
	}

	if err := storage.Rewrite(cmd.Table, newRows); err != nil {
		return nil, err
	}

	return &ExecutionResult{Message: fmt.Sprintf("✅ %d data geus dipiceun", deletedCount)}, nil
}

// ==========================================
// HELPERS
// ==========================================

func indexOf(field string, fields []string) int {
	for i, f := range fields {
		if f == field {
			return i
		}
	}
	return -1
}

func evaluateOne(row []string, cols []schema.Column, cond parser.Condition) bool {
	idx := -1
	var colType string

	for i, c := range cols {
		if c.Name == cond.Field {
			idx = i
			colType = c.Type
			break
		}
	}

	if idx < 0 || idx >= len(row) {
		return false
	}

	return match(row[idx], cond.Operator, cond.Value, colType)
}

func match(a, op, b, colType string) bool {
	if strings.ToUpper(op) == "JIGA" {
		return strings.Contains(strings.ToLower(a), strings.ToLower(b))
	}

	switch colType {
	case "INT":
		numA, errA := strconv.Atoi(a)
		numB, errB := strconv.Atoi(b)

		if errA != nil || errB != nil {
			return false
		}

		switch op {
		case "=": return numA == numB
		case "!=": return numA != numB
		case ">": return numA > numB
		case "<": return numA < numB
		case ">=": return numA >= numB
		case "<=": return numA <= numB
		}

	case "FLOAT":
		fA, errA := strconv.ParseFloat(a, 64)
		fB, errB := strconv.ParseFloat(b, 64)

		if errA != nil || errB != nil {
			return false
		}

		switch op {
		case "=": return fA == fB
		case "!=": return fA != fB
		case ">": return fA > fB
		case "<": return fA < fB
		case ">=": return fA >= fB
		case "<=": return fA <= fB
		}

	case "BOOL":
		if op == "=" { return a == b }
		if op == "!=" { return a != b }
		return false

	case "STRING", "TEXT", "CHAR", "ENUM", "DATE":
		switch op {
		case "=": return a == b
		case "!=": return a != b
		case ">": return a > b
		case "<": return a < b
		case ">=": return a >= b
		case "<=": return a <= b
		}

	default:
		return false
	}

	return false
}