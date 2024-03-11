package main

import (
	"net/http"
	"os"
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

var constant map[string]string

var is_postgres bool = os.Getenv("DATABASE_URL") != ""

var pages_to_collect string
var chainload_uris string
var send_alerts bool

func initalize_constant() {
	constant = make(map[string]string)
}

func get_host(request *http.Request) string {
	host := get_env("DOMAIN")
	if host == "" {
		host = "https://" + request.Host
	}
	return host
}

func get_pages_to_collect() string {
	return pages_to_collect
}

func set_pages_to_collect() {
	db := establish_database_connection()
	defer db.Close()

	var pages_to_collect_value string
	db.QueryRow("SELECT value FROM settings WHERE key = ?", PAGES_TO_COLLECT_SETTINGS_KEY).Scan(&pages_to_collect_value)
	pages_to_collect = "[" + pages_to_collect_value + "]"
}

func get_chainload_uri() string {
	return chainload_uris
}

func set_chainload_uri() {
	db := establish_database_connection()
	defer db.Close()

	db.QueryRow("SELECT value FROM settings WHERE key = ?", CHAINLOAD_URI_SETTINGS_KEY).Scan(&chainload_uris)
}

func get_send_alerts() bool {
	return send_alerts
}

func set_send_alerts() {
	db := establish_database_connection()
	defer db.Close()

	db.QueryRow("SELECT value FROM settings WHERE key = ?", SEND_ALERTS_SETTINGS_KEY).Scan(&send_alerts)
}

func get_screenshot_directory() string {
	return "./screenshots"
}

func get_sqlite_database_path() string {
	return "./db/xsshunter-go.db"
}
