package executor

import (
	"fmt"
	"strings"

	"github.com/febrd/maungdb/engine/auth"
	"github.com/febrd/maungdb/engine/schema"
	"github.com/febrd/maungdb/engine/storage"
)


func ValidateConstraints(d *schema.Definition, tableName string, rowData string) error {
	newCols := strings.Split(rowData, "|")

	if len(newCols) != len(d.Columns) {
		return fmt.Errorf("jumlah kolom teu sesuai (harap: %d, dikirim: %d)", len(d.Columns), len(newCols))
	}

	user, _ := auth.CurrentUser()

	for i, col := range d.Columns {
		val := strings.TrimSpace(newCols[i]) 
		if col.IsNotNull {
			if val == "" || strings.ToUpper(val) == "NULL" {
				return fmt.Errorf("kolom '%s' teu kenging kosong (NOT NULL)", col.Name)
			}
		}

		if col.IsPrimary || col.IsUnique {
			if val != "" {
				isDup, err := checkDuplicate(tableName, i, val)
				if err != nil {
					return fmt.Errorf("gagal cek duplikasi: %v", err)
				}
				if isDup {
					constraintType := "UNIQUE"
					if col.IsPrimary {
						constraintType = "PRIMARY KEY"
					}
					return fmt.Errorf("pelanggaran %s di kolom '%s': data '%s' parantos aya", constraintType, col.Name, val)
				}
			}
		}

		if col.ForeignKey != "" && val != "" && strings.ToUpper(val) != "NULL" {
			parts := strings.Split(col.ForeignKey, ".")
			if len(parts) != 2 {
				return fmt.Errorf("definisi FK salah di kolom %s (format kedah: tabel.kolom)", col.Name)
			}

			targetTable := strings.ToLower(strings.TrimSpace(parts[0]))			
			targetCol := strings.TrimSpace(parts[1])
			exists, err := checkForeignKeyExists(user.Database, targetTable, targetCol, val)
			if err != nil {
				return fmt.Errorf("gagal validasi FK: %v", err)
			}
			if !exists {
				return fmt.Errorf("violation foreign key: data '%s' teu kapanggih di tabel induk '%s.%s'", val, targetTable, targetCol)
			}
		}
	}

	return nil
}

func checkDuplicate(tableName string, colIndex int, value string) (bool, error) {
	rows, err := storage.ReadAll(tableName)
	if err != nil {
		return false, nil 
	}

	for _, row := range rows {
		if row == "" { continue }
		cols := strings.Split(row, "|")

		if colIndex < len(cols) {
			if strings.TrimSpace(cols[colIndex]) == value {
				return true, nil 
			}
		}
	}
	return false, nil
}

func checkForeignKeyExists(dbName string, targetTable string, targetColName string, value string) (bool, error) {
	targetDef, err := schema.Load(dbName, targetTable)
	if err != nil {
		return false, fmt.Errorf("tabel induk '%s' teu kapanggih (error: %v)", targetTable, err)
	}

	targetIndex := targetDef.GetColumnIndex(targetColName)
	if targetIndex == -1 {
		for i, col := range targetDef.Columns {
			if strings.EqualFold(col.Name, targetColName) {
				targetIndex = i
				break
			}
		}
	}

	if targetIndex == -1 {
		return false, fmt.Errorf("kolom '%s' teu aya di tabel induk '%s' (pastikeun ejaan leres)", targetColName, targetTable)
	}
	found, err := checkDuplicate(targetTable, targetIndex, value)
	if err != nil {
		return false, err
	}

	return found, nil
}