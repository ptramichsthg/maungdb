package executor

import (
	"errors"
	"fmt"
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

// ===========================
// 1. INSERT (SIMPEN)
// ===========================
func execInsert(cmd *parser.Command) (*ExecutionResult, error) {
	user, _ := auth.CurrentUser()

	s, err := schema.Load(user.Database, cmd.Table)
	if err != nil {
		return nil, err
	}

	if !s.Can(user.Role, "write") {
		return nil, errors.New("teu boga hak nulis")
	}

	// Validasi Tipe Data sesuai Schema
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

// ===========================
// 2. SELECT (TINGALI)
// ===========================
func execSelect(cmd *parser.Command) (*ExecutionResult, error) {
	user, _ := auth.CurrentUser()

	s, err := schema.Load(user.Database, cmd.Table)
	if err != nil {
		return nil, err
	}

	if !s.Can(user.Role, "read") {
		return nil, errors.New("teu boga hak maca")
	}

	rawRows, err := storage.ReadAll(cmd.Table)
	if err != nil {
		return nil, err
	}

	var parsedRows [][]string
	fieldNames := s.GetFieldNames()

	for _, raw := range rawRows {
		if raw == "" { continue } // Skip baris kosong
		cols := strings.Split(raw, "|")

		// Lamun euweuh WHERE, asupkeun kabeh
		if len(cmd.Where) == 0 {
			parsedRows = append(parsedRows, cols)
			continue
		}

		// Logic WHERE (Multi Condition)
		matchAll := true
		
		// Evaluasi kondisi kahiji
		currentMatch := evaluateOne(cols, s.Columns, cmd.Where[0])

		// Loop pikeun kondisi saterusna (DAN / ATAU)
		for i := 0; i < len(cmd.Where); i++ {
			cond := cmd.Where[i]

			// Mun ieu kondisi terakhir, hasilna geus dicekel currentMatch
			if cond.LogicOp == "" {
				matchAll = currentMatch
				break
			}

			// Evaluasi kondisi saterusna
			if i+1 < len(cmd.Where) {
				nextResult := evaluateOne(cols, s.Columns, cmd.Where[i+1])

				if cond.LogicOp == "sareng" || cond.LogicOp == "SARENG" {
					currentMatch = currentMatch && nextResult
				} else if cond.LogicOp == "ATAWA" || cond.LogicOp == "atawa" {
					currentMatch = currentMatch || nextResult
				}
			}
		}
		matchAll = currentMatch

		if matchAll {
			parsedRows = append(parsedRows, cols)
		}
	}

	return &ExecutionResult{
		Columns: fieldNames,
		Rows:    parsedRows,
	}, nil
}

// ===========================
// 3. UPDATE (OMEAN)
// ===========================
func execUpdate(cmd *parser.Command) (*ExecutionResult, error) {
	user, _ := auth.CurrentUser()
	s, err := schema.Load(user.Database, cmd.Table)
	if err != nil { return nil, err }
	if !s.Can(user.Role, "write") { return nil, errors.New("teu boga hak nulis (omean)") }

	rawRows, err := storage.ReadAll(cmd.Table)
	if err != nil { return nil, err }

	var newRows []string
	updatedCount := 0

	for _, raw := range rawRows {
		if raw == "" { continue }
		cols := strings.Split(raw, "|")

		// Cek WHERE
		shouldUpdate := false
		if len(cmd.Where) == 0 {
			shouldUpdate = true // Mun euweuh where, update kabeh!
		} else {
			// Reuse logic evaluateOne (Sederhana: anggap AND kabeh keur update MVP)
			// Di dieu urang pake logic nu sami jeung Select
			currentMatch := evaluateOne(cols, s.Columns, cmd.Where[0])
			shouldUpdate = currentMatch
		}

		if shouldUpdate {
			// Lakukan perubahan data
			for colName, newVal := range cmd.Updates {
				idx := indexOf(colName, s.GetFieldNames())
				if idx != -1 {
					// Disini idealna validasi tipe data heula
					cols[idx] = newVal 
				}
			}
			updatedCount++
		}
		
		newRows = append(newRows, strings.Join(cols, "|"))
	}

	// Tulis ulang file database
	if err := storage.Rewrite(cmd.Table, newRows); err != nil {
		return nil, err
	}

	return &ExecutionResult{Message: fmt.Sprintf("✅ %d data geus diomean", updatedCount)}, nil
}

// ===========================
// 4. DELETE (MICEUN)
// ===========================
func execDelete(cmd *parser.Command) (*ExecutionResult, error) {
	user, _ := auth.CurrentUser()
	s, err := schema.Load(user.Database, cmd.Table)
	if err != nil { return nil, err }
	if !s.Can(user.Role, "write") { return nil, errors.New("teu boga hak nulis (miceun)") }

	rawRows, err := storage.ReadAll(cmd.Table)
	if err != nil { return nil, err }

	var newRows []string
	deletedCount := 0

	for _, raw := range rawRows {
		if raw == "" { continue }
		cols := strings.Split(raw, "|")

		shouldDelete := false
		if len(cmd.Where) > 0 {
			shouldDelete = evaluateOne(cols, s.Columns, cmd.Where[0])
		}

		if shouldDelete {
			deletedCount++
			continue // Skip (ulah diasupkeun ka newRows) -> Ieu nu ngahapus
		}
		
		newRows = append(newRows, raw)
	}

	// Tulis ulang file database
	if err := storage.Rewrite(cmd.Table, newRows); err != nil {
		return nil, err
	}

	return &ExecutionResult{Message: fmt.Sprintf("✅ %d data geus dipiceun", deletedCount)}, nil
}

// ===========================
// HELPERS
// ===========================

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

func indexOf(field string, fields []string) int {
	for i, f := range fields {
		if f == field {
			return i
		}
	}
	return -1
}
// match ayeuna ngalakukeun Type Casting Lengkep (INT, FLOAT, BOOL, DATE, ENUM, jsb)
func match(a, op, b, colType string) bool {
	switch colType {
	case "INT":
		numA, errA := strconv.Atoi(a)
		numB, errB := strconv.Atoi(b)
		// Mun gagal convert (misal data ruksak), anggap false
		if errA != nil || errB != nil { return false }

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
		if errA != nil || errB != nil { return false }

		switch op {
		case "=": return fA == fB
		case "!=": return fA != fB
		case ">": return fA > fB
		case "<": return fA < fB
		case ">=": return fA >= fB
		case "<=": return fA <= fB
		}

	case "BOOL":
		// Bool ngan ukur bisa Cek Sarua (=) atawa Teu Sarua (!=)
		// Teu asup akal mun "true > false"
		if op == "=" { return a == b }
		if op == "!=" { return a != b }
		return false

	// GROUP: Tipe data nu dibandingkeun secara Text / Leksikal
	// DATE (YYYY-MM-DD) aman dibandingkeun siga string
	// CHAR, ENUM, TEXT, STRING sami sadayana
	case "STRING", "TEXT", "CHAR", "ENUM", "DATE":
		switch op {
		case "=": return a == b
		case "!=": return a != b
		case ">": return a > b
		case "<": return a < b
		case ">=": return a >= b
		case "<=": return a <= b
		}

	// Default Fallback (bisi aya tipe nu kaliwat)
	default:
		return false
	}
	
	return false
}