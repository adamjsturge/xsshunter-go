package main

import (
	"log"
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

// var constant = make(map[string]string)

var is_postgres bool = get_env("DATABASE_URL") != ""

var pages_to_collect string
var chainload_uris string
var send_alerts bool

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
	pages_to_collect_value, err := db_single_item_query("SELECT value FROM settings WHERE key = $1", PAGES_TO_COLLECT_SETTINGS_KEY).toString()
	if err != nil {
		log.Fatal("Fatal Error on set page collect:", err)
	}

	pages_to_collect = "[" + pages_to_collect_value + "]"
}

func get_chainload_uri() string {
	return chainload_uris
}

func set_chainload_uri() {
	chainload_uris_value, err := db_single_item_query("SELECT value FROM settings WHERE key = $1", CHAINLOAD_URI_SETTINGS_KEY).toString()
	if err != nil {
		log.Fatal("Fatal Error on set chainload uri:", err)
	}
	chainload_uris = chainload_uris_value
}

func get_send_alerts() bool {
	return send_alerts
}

func set_send_alerts() {
	send_alerts_value, err := db_single_item_query("SELECT value FROM settings WHERE key = $1", SEND_ALERTS_SETTINGS_KEY).toBool()
	if err != nil {
		log.Fatal("Fatal Error on set alert:", err)
	}
	send_alerts = send_alerts_value
}

func get_screenshot_directory() string {
	return "./screenshots"
}

func get_sqlite_database_path() string {
	return "./db/xsshunter-go.db"
}
