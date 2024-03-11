package main

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"os"
	"strconv"
	"strings"

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
	if constant[key] == "" {
		constant[key] = os.Getenv(key)
	}
	return constant[key]
}

func parameter_to_int(input string, default_int int) int {
	if input == "" {
		return default_int
	}
	value, err := strconv.Atoi(input)
	if err != nil {
		return default_int
	}
	return value
}

func update_setting(setting_key string, setting_value string) {
	db := establish_database_connection()
	defer db.Close()
	db.Exec("UPDATE settings SET value = :value WHERE key = :key", setting_value, setting_key)
}

func make_folder_if_not_exists(folder string) {
	if !checkFileExists(folder) {
		os.MkdirAll(folder, 0755)
	}
}

func generate_screenshot_url(request *http.Request, screenshot_id string) string {
	if get_env("SCREENSHOTS_REQUIRE_AUTH") == "true" {
		return ""
		// return get_host(request) + "/screenshot/" + screenshot_id + "?auth=" +
	}
	return get_host(request) + "/screenshots/" + screenshot_id + ".png"
}

func get_client_ip(request *http.Request) string {
	clientIP := request.Header.Get("X-Forwarded-For")
	if clientIP == "" {
		return request.RemoteAddr
	}

	ips := strings.Split(clientIP, ",")
	if len(ips) > 0 {
		clientIP = ips[0]
	}
	return clientIP
}

// func remember(variable *string, reload bool, function func() string) string {
// 	if reload || *variable == "" {
// 		*variable = function()
// 	}
// 	return *variable
// }
