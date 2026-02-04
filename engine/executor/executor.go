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
	default:
		return nil, errors.New("command teu didukung")
	}
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

	// Validasi Tipe Data (INT vs STRING)
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
	fieldNames := s.GetFieldNames() // Helper nu anyar

	for _, raw := range rawRows {
		cols := strings.Split(raw, "|")

		if len(cmd.Where) == 0 {
			parsedRows = append(parsedRows, cols)
			continue
		}

		// Logic AND / OR
		matchAll := true
		
		// Initial check keur elemen kahiji
		currentMatch := evaluateOne(cols, s.Columns, cmd.Where[0])

		for i := 0; i < len(cmd.Where); i++ {
			cond := cmd.Where[i]
			
			if cond.LogicOp == "" {
				matchAll = currentMatch
				break
			}

			nextResult := evaluateOne(cols, s.Columns, cmd.Where[i+1])

			if cond.LogicOp == "DAN" {
				currentMatch = currentMatch && nextResult
			} else if cond.LogicOp == "ATAU" {
				currentMatch = currentMatch || nextResult
			}
		}

		if matchAll {
			parsedRows = append(parsedRows, cols)
		}
	}

	return &ExecutionResult{
		Columns: fieldNames,
		Rows:    parsedRows,
	}, nil
}

// evaluateOne ayeuna narima []schema.Column supaya nyaho tipe datana
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

// match ayeuna ngalakukeun Type Casting
func match(a, op, b, colType string) bool {
	// Lamun tipe INT, convert heula
	if colType == "INT" {
		numA, errA := strconv.Atoi(a)
		numB, errB := strconv.Atoi(b)

		// Lamun gagal convert, anggap false (data ruksak)
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
		default: return false
		}
	}

	// Default: STRING Comparison
	switch op {
	case "=": return a == b
	case "!=": return a != b
	case ">": return a > b
	case "<": return a < b
	case ">=": return a >= b
	case "<=": return a <= b
	default: return false
	}
}