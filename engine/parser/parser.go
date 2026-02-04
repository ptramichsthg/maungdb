package parser

import (
	"errors"
	"strings"
)

func Parse(input string) (*Command, error) {
	tokens := strings.Fields(input)
	if len(tokens) < 2 {
		return nil, errors.New("query teu valid")
	}

	switch tokens[0] {
	case "simpen":
		return parseInsert(tokens)
	case "tingali":
		return parseSelect(tokens)
	default:
		return nil, errors.New("paréntah teu dikenal")
	}
}

func parseInsert(tokens []string) (*Command, error) {
	if len(tokens) < 3 {
		return nil, errors.New("format simpen salah: simpen <table> <data>")
	}
	return &Command{
		Type:  CmdInsert,
		Table: tokens[1],
		Data:  tokens[2],
	}, nil
}

func parseSelect(tokens []string) (*Command, error) {
	cmd := &Command{
		Type:  CmdSelect,
		Table: tokens[1],
		Where: []Condition{},
	}

	if len(tokens) > 2 {
		if tokens[2] != "dimana" {
			return nil, errors.New("keyword salah, kedahna 'dimana'")
		}

		// Parser logic pikeun multi-kondisi (AND/OR)
		// token nu sesa: [col, op, val, (DAN/ATAU), col, op, val, ...]
		remaining := tokens[3:]
		
		for len(remaining) >= 3 {
			cond := Condition{
				Field:    remaining[0],
				Operator: remaining[1],
				Value:    remaining[2],
				LogicOp:  "", // Default kosong
			}

			// Cek naha aya logika saterusna (DAN/ATAU)
			if len(remaining) > 3 {
				logic := strings.ToUpper(remaining[3])
				if logic == "SARENG" || logic == "ATAWA" || logic == "sareng" || logic == "atawa" {
					cond.LogicOp = logic
					remaining = remaining[4:] // Geser ka kondisi saterusna
				} else {
					return nil, errors.New("operator logika teu dikenal: " + remaining[3])
				}
			} else {
				remaining = nil // Rengse
			}

			cmd.Where = append(cmd.Where, cond)
		}
	}

	return cmd, nil
}