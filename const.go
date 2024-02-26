package main

import (
	"net/http"
)

const (
	API_BASE_PATH                       = "/api/v1"
	ADMIN_PASSWORD_SETTINGS_KEY         = "ADMIN_PASSWORD"
	session_secret_key                  = "SESSION_SECRET"
	CORRELATION_API_SECRET_SETTINGS_KEY = "CORRELATION_API_KEY"
	CHAINLOAD_URI_SETTINGS_KEY          = "CHAINLOAD_URI"
	PAGES_TO_COLLECT_SETTINGS_KEY       = "PAGES_TO_COLLECT"
	SEND_ALERTS_SETTINGS_KEY            = "SEND_ALERTS"
	csrf_header_name                    = "X-CSRF-Buster"
)

func get_host(request *http.Request) string {
	host := get_env("DOMAIN")
	if host == "" {
		host = "https://" + request.Host
	}

	return host
}

func get_pages_to_collect() string {
	db := establish_database_connection()
	defer db.Close()

	var pages_to_collect string
	db.QueryRow("SELECT value FROM settings WHERE key = ?", PAGES_TO_COLLECT_SETTINGS_KEY).Scan(&pages_to_collect)
	return "[" + pages_to_collect + "]"
}

func get_chainload_uri() string {
	db := establish_database_connection()
	defer db.Close()

	var chainload_uris string
	db.QueryRow("SELECT value FROM settings WHERE key = ?", CHAINLOAD_URI_SETTINGS_KEY).Scan(&chainload_uris)
	return chainload_uris
}

func get_screenshot_directory() string {
	return "./screenshots"
}

func get_sqlite_database_path() string {
	return "./db/xsshunter-go.db"
}
