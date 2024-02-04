package main

import (
	"crypto/rand"
	"encoding/hex"
	"os"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

func get_secure_random_string(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func hash_string(input string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(input), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func check_hash(input string, hashed string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(input))
	return err == nil
}

// func generate_log(input string) {
// 	datetime :=
// }

func checkFileExists(filepath string) bool {
	_, err := os.Stat(filepath)
	return !os.IsNotExist(err)
}

func get_env(key string) string {
	err := godotenv.Load()
	if err != nil {
		return ""
	}
	return os.Getenv(key)
}
