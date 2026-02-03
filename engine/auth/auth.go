package auth

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/febrd/maungdb/internal/config"
)

type User struct {
	Username string
	Role     string
}

var currentUser *User


func Login(username, password string) error {
	userFile := filepath.Join(
		config.DataDir,
		config.SystemDir,
		"users.maung",
	)

	file, err := os.Open(userFile)
	if err != nil {
		return errors.New("system user file teu kapanggih")
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "|")
		if len(parts) != 3 {
			continue
		}

		if parts[0] == username && parts[1] == password {
			currentUser = &User{
				Username: username,
				Role:     parts[2],
			}
			return nil
		}
	}

	return errors.New("login gagal")
}

func CurrentUser() *User {
	return currentUser
}


func RequireRole(minRole string) error {
	if currentUser == nil {
		return errors.New("can login heula")
	}

	userLevel, ok1 := config.Roles[currentUser.Role]
	requiredLevel, ok2 := config.Roles[minRole]

	if !ok1 || !ok2 {
		return errors.New("role teu dikenal")
	}

	if userLevel > requiredLevel {
		return errors.New("hak aksés teu cukup")
	}

	return nil
}
