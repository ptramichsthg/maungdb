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

	switch strings.ToUpper(tokens[0]) {
	case "SIMPEN":
		return parseInsert(tokens)
	case "TINGALI":
		return parseSelect(tokens)
	case "OMEAN": // UPDATE
		return parseUpdate(tokens)
	case "MICEUN": // DELETE
		return parseDelete(tokens)
	default:
		return nil, errors.New("paréntah teu dikenal")
	}
}

// ... parseInsert & parseSelect (tetep sami) ...

// Sintaks: OMEAN <table_name> JADI <col>=<val> DIMANA ...
func parseUpdate(tokens []string) (*Command, error) {
	if len(tokens) < 4 || strings.ToUpper(tokens[2]) != "JADI" {
		return nil, errors.New("format OMEAN salah: OMEAN <table> JADI <col>=<val> DIMANA ...")
	}

	// Parse "col=val"
	pairs := strings.Split(tokens[3], "=")
	if len(pairs) != 2 {
		return nil, errors.New("format update salah, gunakeun col=val")
	}

	cmd := &Command{
		Type:    CmdUpdate,
		Table:   tokens[1],
		Updates: map[string]string{pairs[0]: pairs[1]},
		Where:   []Condition{},
	}

	// Parse WHERE (dimana)
	if len(tokens) > 4 {
		if strings.ToUpper(tokens[4]) != "DIMANA" {
			return nil, errors.New("kedah nganggo DIMANA")
		}
		whereCmd, err := parseWhere(tokens[5:]) // Reuse logic WHERE
		if err != nil {
			return nil, err
		}
		cmd.Where = whereCmd.Where
	}

	return cmd, nil
}

// Sintaks: MICEUN TI <table_name> DIMANA ...
func parseDelete(tokens []string) (*Command, error) {
	if len(tokens) < 3 || strings.ToUpper(tokens[1]) != "TI" {
		return nil, errors.New("format MICEUN salah: MICEUN TI <table> DIMANA ...")
	}

	cmd := &Command{
		Type:  CmdDelete,
		Table: tokens[2],
		Where: []Condition{},
	}

	if len(tokens) > 3 {
		if strings.ToUpper(tokens[3]) != "DIMANA" {
			return nil, errors.New("kedah nganggo DIMANA")
		}
		whereCmd, err := parseWhere(tokens[4:])
		if err != nil {
			return nil, err
		}
		cmd.Where = whereCmd.Where
	}

	return cmd, nil
}

// Helper misahkeun logic WHERE supaya bisa dipake ku SELECT, UPDATE, DELETE
func parseWhere(tokens []string) (*Command, error) {
	cmd := &Command{Where: []Condition{}}
	remaining := tokens
	
	for len(remaining) >= 3 {
		cond := Condition{
			Field:    remaining[0],
			Operator: remaining[1],
			Value:    remaining[2],
			LogicOp:  "",
		}
		if len(remaining) > 3 {
			logic := strings.ToUpper(remaining[3])
			if logic == "SARENG" || logic == "ATAWA" || logic == "sareng" || logic == "atawa" {
				cond.LogicOp = logic
				remaining = remaining[4:]
			} else {
				remaining = nil
			}
		} else {
			remaining = nil
		}
		cmd.Where = append(cmd.Where, cond)
	}
	return cmd, nil
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
		if tokens[2] != "dimana" || tokens[2] != "DIMANA" {
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