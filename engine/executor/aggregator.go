package executor

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

type AggregateFunc string

const (
	FuncCount = "JUMLAH"       
	FuncSum   = "TOTAL"       
	FuncAvg   = "RATA"         
	FuncMax   = "PANGGEDENA"   
	FuncMin   = "PANGLEUTIKNA" 
)

type ParsedColumn struct {
	OriginalText string
	IsAggregate  bool
	FuncType     AggregateFunc
	TargetCol    string 
}

func ParseColumnSelection(rawCol string) ParsedColumn {
	rawCol = strings.TrimSpace(rawCol)

	if strings.Contains(rawCol, "(") && strings.Contains(rawCol, ")") {
		start := strings.Index(rawCol, "(")
		end := strings.LastIndex(rawCol, ")")
		
		funcName := strings.ToUpper(strings.TrimSpace(rawCol[:start]))
		target := strings.TrimSpace(rawCol[start+1 : end])

		var aggType AggregateFunc
		isValid := true

		switch funcName {
		case FuncCount: aggType = FuncCount
		case FuncSum:   aggType = FuncSum
		case FuncAvg:   aggType = FuncAvg
		case FuncMax:   aggType = FuncMax
		case FuncMin:   aggType = FuncMin
		default:
			isValid = false
		}

		if isValid {
			return ParsedColumn{
				OriginalText: rawCol,
				IsAggregate:  true,
				FuncType:     aggType,
				TargetCol:    target,
			}
		}
	}

	return ParsedColumn{
		OriginalText: rawCol,
		IsAggregate:  false,
		TargetCol:    rawCol,
	}
}

func CalculateAggregate(rows []map[string]string, parsedCol ParsedColumn) (string, error) {
	if len(rows) == 0 {
		return "0", nil
	}

	if parsedCol.FuncType == FuncCount {
		return fmt.Sprintf("%d", len(rows)), nil
	}

	var sum float64
	var count float64
	var maxVal = -math.MaxFloat64
	var minVal = math.MaxFloat64
	
	for _, row := range rows {
		valStr, ok := row[parsedCol.TargetCol]
		if !ok {
			continue 
		}

		val, err := strconv.ParseFloat(valStr, 64)
		if err != nil {
			continue 
		}

		switch parsedCol.FuncType {
		case FuncSum, FuncAvg:
			sum += val
			count++
		case FuncMax:
			if val > maxVal { maxVal = val }
		case FuncMin:
			if val < minVal { minVal = val }
		}
	}

	switch parsedCol.FuncType {
	case FuncSum:
		return fmt.Sprintf("%.2f", sum), nil
	case FuncAvg:
		if count == 0 { return "0", nil }
		return fmt.Sprintf("%.2f", sum/count), nil
	case FuncMax:
		if maxVal == -math.MaxFloat64 { return "0", nil }
		return fmt.Sprintf("%.2f", maxVal), nil
	case FuncMin:
		if minVal == math.MaxFloat64 { return "0", nil }
		return fmt.Sprintf("%.2f", minVal), nil
	}

	return "Error", nil
}