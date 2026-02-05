package storage

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	 "encoding/csv"
	"github.com/febrd/maungdb/engine/auth"
	"github.com/febrd/maungdb/internal/config"
	"golang.org/x/crypto/bcrypt"
	
)

var mutex sync.Mutex


func Init() error {
	if err := os.MkdirAll(config.DataDir, 0755); err != nil {
		return err
	}

	systemPath := filepath.Join(config.DataDir, config.SystemDir)
	if err := os.MkdirAll(systemPath, 0755); err != nil {
		return err
	}

	if err := initDefaultUser(systemPath); err != nil {
		return err
	}

	return nil
}

func tablePath(database, table string) (string, error) {
	dbPath := filepath.Join(config.DataDir, "db_"+database)

	if _, err := os.Stat(dbPath); err != nil {
		return "", errors.New("database teu kapanggih")
	}

	for _, ext := range config.AllowedExt {
		p := filepath.Join(dbPath, table+ext)
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}

	return filepath.Join(dbPath, table+config.AllowedExt[0]), nil
}

func Append(table, data string) error {
	u, err := auth.CurrentUser()
	if err != nil {
		return err
	}

	if u.Database == "" {
		return errors.New("can use database heula")
	}

	path, err := tablePath(u.Database, table)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(
		path,
		os.O_CREATE|os.O_APPEND|os.O_WRONLY,
		0644,
	)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(data + "\n")
	return err
}

func ReadAll(table string) ([]string, error) {
	u, err := auth.CurrentUser()
	if err != nil {
		return nil, err
	}

	if u.Database == "" {
		return nil, errors.New("can use database heula")
	}

	path, err := tablePath(u.Database, table)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, errors.New("table teu kapanggih")
	}
	defer file.Close()
	var rows []string
	sc := bufio.NewScanner(file)
	for sc.Scan() {
		rows = append(rows, sc.Text())
	}

	return rows, nil
}

func InitTableFile(database, table string) error {
	dbPath := filepath.Join(config.DataDir, "db_"+database)
	path := filepath.Join(dbPath, table+".mg")

	if _, err := os.Stat(path); err == nil {
		return nil
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}

	return file.Close()
}

func initDefaultUser(systemPath string) error {
	userFile := filepath.Join(systemPath, "users.maung")

	if _, err := os.Stat(userFile); err == nil {
		return nil 
	}

	hash, _ := bcrypt.GenerateFromPassword(
		[]byte(config.DefaultPass),
		bcrypt.DefaultCost,
	)

	line := strings.Join([]string{
		config.DefaultUser,
		string(hash),
		config.DefaultRole,
		"*",
	}, "|") + "\n"

	return os.WriteFile(userFile, []byte(line), 0644)
}


func Rewrite(table string, rows []string) error {
	u, err := auth.CurrentUser()
	if err != nil {
		return err
	}
	
	path, err := tablePath(u.Database, table)
	if err != nil {
		return err
	}

	content := strings.Join(rows, "\n")
	
	
	if len(rows) > 0 {
		content += "\n" 
	}

	return os.WriteFile(path, []byte(content), 0644)
}

func ExportCSV(table string) (string, error) {
	mutex.Lock()
	defer mutex.Unlock()

	rows, err := ReadAll(table)
	if err != nil {
		return "", err
	}

	filename := filepath.Join(config.DataDir, table+".csv")
	file, err := os.Create(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, row := range rows {
		if row == "" { continue }
		cols := strings.Split(row, "|")
		if err := writer.Write(cols); err != nil {
			return "", err
		}
	}

	return filename, nil
}

func ImportCSV(table string, filePath string) (int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return 0, err
	}

	count := 0
	for _, record := range records {
		rowStr := strings.Join(record, "|")
		if err := Append(table, rowStr); err == nil {
			count++
		}
	}
	return count, nil
}