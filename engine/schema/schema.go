package schema

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/febrd/maungdb/internal/config"
)

type Schema struct {
	Table       string              `json:"table"`
	Fields      []string            `json:"fields"`
	Permissions map[string][]string `json:"permissions"`
}

func Load(table string) (*Schema, error) {
	path := filepath.Join(
		config.DataDir,
		config.SchemaDir,
		table+".tpk",
	)

	file, err := os.Open(path)
	if err != nil {
		return nil, errors.New("schema teu kapanggih")
	}
	defer file.Close()

	var s Schema
	if err := json.NewDecoder(file).Decode(&s); err != nil {
		return nil, err
	}

	return &s, nil
}

func (s *Schema) ValidateRow(row string) error {
	values := splitRow(row)
	if len(values) != len(s.Fields) {
		return errors.New("jumlah kolom teu saluyu jeung schema")
	}
	return nil
}

func splitRow(row string) []string {
	var res []string
	current := ""

	for _, c := range row {
		if c == '|' {
			res = append(res, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	res = append(res, current)
	return res
}

func (s *Schema) Can(role, action string) bool {
	roles, ok := s.Permissions[action]
	if !ok {
		return false
	}

	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}
