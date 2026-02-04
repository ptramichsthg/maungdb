package schema

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/febrd/maungdb/internal/config"
)

type Column struct {
	Name string
	Type string // "STRING", "INT"
}

type Definition struct {
	Columns []Column
	Perms   map[string][]string
}

// Create schema anyar kalayan tipe data (format fields: "nama:string", "umur:int")
func Create(database, table string, fieldsRaw []string, perms map[string][]string) error {
	path := filepath.Join(config.DataDir, "db_"+database, table+".schema")

	// Parse fields "name:type"
	var headerParts []string
	for _, f := range fieldsRaw {
		parts := strings.Split(f, ":")
		if len(parts) != 2 {
			return errors.New("format salah, gunakeun 'kolom:tipe' (conto: umur:int)")
		}
		
		validType := strings.ToUpper(parts[1])
		if validType != "STRING" && validType != "INT" {
			return errors.New("tipe data teu didukung: " + parts[1])
		}

		headerParts = append(headerParts, parts[0]+":"+validType)
	}

	// 1. Tulis Header (fields)
	content := strings.Join(headerParts, ",") + "\n"

	// 2. Tulis Permissions
	for role, actions := range perms {
		content += fmt.Sprintf("%s=%s\n", role, strings.Join(actions, ","))
	}

	return os.WriteFile(path, []byte(content), 0644)
}

// Load maca schema jeung tipe datana
func Load(database, table string) (*Definition, error) {
	path := filepath.Join(config.DataDir, "db_"+database, table+".schema")
	
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.New("table teu kapanggih")
	}

	lines := strings.Split(string(content), "\n")
	if len(lines) < 1 {
		return nil, errors.New("schema ruksak")
	}

	// Parse Header: "nama:STRING,umur:INT"
	rawCols := strings.Split(lines[0], ",")
	var columns []Column

	for _, rc := range rawCols {
		parts := strings.Split(rc, ":")
		if len(parts) == 2 {
			columns = append(columns, Column{Name: parts[0], Type: parts[1]})
		} else {
			// Fallback mun schema heubeul (dianggap STRING)
			columns = append(columns, Column{Name: parts[0], Type: "STRING"})
		}
	}

	def := &Definition{
		Columns: columns,
		Perms:   make(map[string][]string),
	}

	// Parse Permissions
	for _, line := range lines[1:] {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "=")
		if len(parts) == 2 {
			def.Perms[parts[0]] = strings.Split(parts[1], ",")
		}
	}

	return def, nil
}

// ValidateRow mastikeun data nu asup sesuai jeung tipe kolom
func (d *Definition) ValidateRow(data string) error {
	values := strings.Split(data, "|")
	if len(values) != len(d.Columns) {
		return errors.New("jumlah kolom teu sesuai")
	}

	for i, col := range d.Columns {
		val := strings.TrimSpace(values[i])
		
		if col.Type == "INT" {
			if _, err := strconv.Atoi(val); err != nil {
				return fmt.Errorf("kolom '%s' kudu angka (INT), tapi meunang '%s'", col.Name, val)
			}
		}
		// STRING narima naon wae
	}

	return nil
}

func (d *Definition) Can(role, action string) bool {
	if role == "supermaung" {
		return true
	}

	// Cek action aya teu
	allowedRoles, ok := d.Perms[action]
	if !ok {
		return false
	}

	// Cek role aya dina permission
	for _, r := range allowedRoles {
		if r == role {
			return true
		}
	}

	return false
}

// Helper nyokot daptar ngaran kolom (keur Executor)
func (d *Definition) GetFieldNames() []string {
	var names []string
	for _, c := range d.Columns {
		names = append(names, c.Name)
	}
	return names
}