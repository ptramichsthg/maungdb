package storage

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/febrd/maungdb/internal/config"
)

func Init() error {
	// main data dir
	if err := os.MkdirAll(config.DataDir, 0755); err != nil {
		return err
	}

	// system dir
	systemPath := filepath.Join(config.DataDir, config.SystemDir)
	if err := os.MkdirAll(systemPath, 0755); err != nil {
		return err
	}

	// schema dir
	schemaPath := filepath.Join(config.DataDir, config.SchemaDir)
	if err := os.MkdirAll(schemaPath, 0755); err != nil {
		return err
	}

	// init default user
	return initDefaultUser(systemPath)
}

// =======================
// TABLE HANDLING
// =======================

func findTableFile(table string) (string, error) {
	for _, ext := range config.AllowedExt {
		path := filepath.Join(config.DataDir, table+ext)
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}
	return "", errors.New("table teu kapanggih")
}

func createTableIfNotExist(table string) (string, error) {
	ext := config.AllowedExt[0] // default .mg
	path := filepath.Join(config.DataDir, table+ext)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		file, err := os.Create(path)
		if err != nil {
			return "", err
		}
		file.Close()
	}

	return path, nil
}

func Append(table, data string) error {
	path, err := findTableFile(table)
	if err != nil {
		path, err = createTableIfNotExist(table)
		if err != nil {
			return err
		}
	}

	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(data + "\n")
	return err
}

func ReadAll(table string) ([]string, error) {
	path, err := findTableFile(table)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var rows []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		rows = append(rows, scanner.Text())
	}

	return rows, nil
}

// =======================
// USER SYSTEM
// =======================

func initDefaultUser(systemPath string) error {
	userFile := filepath.Join(systemPath, "users.maung")

	if _, err := os.Stat(userFile); err == nil {
		return nil // already exists
	}

	file, err := os.Create(userFile)
	if err != nil {
		return err
	}
	defer file.Close()

	line := strings.Join([]string{
		config.DefaultUser,
		config.DefaultPass,
		config.DefaultRole,
	}, "|")

	_, err = file.WriteString(line + "\n")
	return err
}
