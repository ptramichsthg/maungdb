package schema

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/febrd/maungdb/internal/config"
)

func GetColumns(database, table string) ([]string, error) {
	path := filepath.Join(
		config.DataDir,
		"db_"+database,
		table+".schema",
	)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(data), "\n")
	if len(lines) == 0 {
		return nil, nil
	}

	defs := strings.Split(lines[0], "|")

	cols := []string{}
	for _, d := range defs {
		kv := strings.Split(d, ":")
		if len(kv) == 2 {
			cols = append(cols, kv[0])
		}
	}

	return cols, nil
}
