package main

import (
	"fmt"
	"log"
)

type Settings struct {
	ID    uint
	Key   *string
	Value *string
}

type PayloadFireResults struct {
	ID                    string `json:"id"`
	Url                   string `json:"url"`
	Ip_address            string `json:"ip_address"`
	Referer               string `json:"referer"`
	User_agent            string `json:"user_agent"`
	Cookies               string `json:"cookies"`
	Title                 string `json:"title"`
	Dom                   string `json:"dom"`
	Text                  string `json:"text"`
	Origin                string `json:"origin"`
	Screenshot_id         string `json:"screenshot_id"`
	Was_iframe            bool   `json:"was_iframe"`
	Browser_timestamp     uint   `json:"browser_timestamp"`
	Correlated_request    string `json:"correlated_request"`
	Injection_requests_id *int   `json:"injection_requests_id"`
}

type CollectedPages struct {
	ID   uint
	Uri  string
	Html string
}

type InjectionRequests struct {
	ID            uint
	Request       string
	Injection_key string
}

func create_sqlite_tables() {
	db := establish_sqlite_database_connection()
	defer db.Close()

	sqlStmt := `
	CREATE TABLE IF NOT EXISTS settings (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		key TEXT,
		value TEXT
	);
	CREATE TABLE IF NOT EXISTS payload_fire_results (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		url TEXT,
		ip_address TEXT,
		referer TEXT,
		user_agent TEXT,
		cookies TEXT,
		title TEXT,
		dom TEXT,
		text TEXT,
		origin TEXT,
		screenshot_id TEXT,
		was_iframe BOOLEAN,
		browser_timestamp UNSIGNED INT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS collected_pages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		uri TEXT,
		html TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS injection_requests (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		request TEXT,
		injection_key TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS user_xss_payloads (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		payload TEXT,
		title TEXT,
		description TEXT,
		author TEXT,
		author_link TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS migrations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err := db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}
}

func create_postgres_tables() {
	db := establish_postgres_database_connection()
	defer db.Close()

	sqlStmt := `
	CREATE TABLE IF NOT EXISTS settings (
		id SERIAL PRIMARY KEY,
		key TEXT,
		value TEXT
	);
	CREATE TABLE IF NOT EXISTS payload_fire_results (
		id SERIAL PRIMARY KEY,
		url TEXT,
		ip_address TEXT,
		referer TEXT,
		user_agent TEXT,
		cookies TEXT,
		title TEXT,
		dom TEXT,
		text TEXT,
		origin TEXT,
		screenshot_id TEXT,
		was_iframe BOOLEAN,
		browser_timestamp BIGINT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS collected_pages (
		id SERIAL PRIMARY KEY,
		uri TEXT,
		html TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS injection_requests (
		id SERIAL PRIMARY KEY,
		request TEXT,
		injection_key TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS user_xss_payloads (
		id SERIAL PRIMARY KEY,
		payload TEXT,
		title TEXT,
		description TEXT,
		author TEXT,
		author_link TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS migrations (
		id SERIAL PRIMARY KEY,
		name TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err := db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}
}

func initialize_settings() {
	initialize_users()
	initialize_configs()
	initialize_correlation_api()
	initialize_setting_helper(PAGES_TO_COLLECT_SETTINGS_KEY, "")
	set_pages_to_collect()
	initialize_setting_helper(SEND_ALERTS_SETTINGS_KEY, "true")
	set_send_alerts()
	initialize_setting_helper(CHAINLOAD_URI_SETTINGS_KEY, "")
	set_chainload_uri()
}

func initialize_users() {
	new_password, err := get_secure_random_string(32)
	if err != nil {
		log.Fatal(err)
	}

	new_user := setup_admin_user(new_password)

	if new_user {
		return
	}

	banner_message := get_default_user_created_banner(new_password)
	fmt.Println(banner_message)
}

func setup_admin_user(password string) bool {
	db := establish_database_connection()
	defer db.Close()

	hashed_password, err := hash_string(password)
	if err != nil {
		log.Fatal(err)
	}

	return initialize_setting_helper(ADMIN_PASSWORD_SETTINGS_KEY, hashed_password)
}

func initialize_configs() {
	session_secret, err := get_secure_random_string(64)
	if err != nil {
		log.Fatal(err)
	}
	initialize_setting_helper(session_secret_key, session_secret)
}

func initialize_correlation_api() {
	api_key, err := get_secure_random_string(64)
	if err != nil {
		log.Fatal(err)
	}
	initialize_setting_helper(CORRELATION_API_SECRET_SETTINGS_KEY, api_key)
}

func initialize_setting_helper(key string, value string) bool {
	db := establish_database_connection()
	defer db.Close()

	has_setting, setting_err := db_single_item_query("SELECT 1 FROM settings WHERE key = $1", key).toBool()
	if setting_err != nil {
		log.Fatal(setting_err)
	}
	if !has_setting {
		_, err := db.Exec("INSERT INTO settings (key, value) VALUES ($1, $2)", key, value)
		if err != nil {
			log.Fatal(err)
		}
		return false
	}
	return true
}

func get_default_user_created_banner(password string) string {
	return `
   ============================================================================
    █████╗ ████████╗████████╗███████╗███╗   ██╗████████╗██╗ ██████╗ ███╗   ██╗
   ██╔══██╗╚══██╔══╝╚══██╔══╝██╔════╝████╗  ██║╚══██╔══╝██║██╔═══██╗████╗  ██║
   ███████║   ██║      ██║   █████╗  ██╔██╗ ██║   ██║   ██║██║   ██║██╔██╗ ██║
   ██╔══██║   ██║      ██║   ██╔══╝  ██║╚██╗██║   ██║   ██║██║   ██║██║╚██╗██║
   ██║  ██║   ██║      ██║   ███████╗██║ ╚████║   ██║   ██║╚██████╔╝██║ ╚████║
   ╚═╝  ╚═╝   ╚═╝      ╚═╝   ╚══════╝╚═╝  ╚═══╝   ╚═╝   ╚═╝ ╚═════╝ ╚═╝  ╚═══╝
																			  
   vvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvv
	   An admin user (for the admin control panel) has been created
	   with the following password:
   
	   PASSWORD: ` + password + `
   
	   XSS Hunter Go has only one user for the instance. Do not
	   share this password with anyone who you don't trust. Save it
	   in your password manager and don't change it to anything that
	   is bruteforcable.
   
   ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
    █████╗ ████████╗████████╗███████╗███╗   ██╗████████╗██╗ ██████╗ ███╗   ██╗
   ██╔══██╗╚══██╔══╝╚══██╔══╝██╔════╝████╗  ██║╚══██╔══╝██║██╔═══██╗████╗  ██║
   ███████║   ██║      ██║   █████╗  ██╔██╗ ██║   ██║   ██║██║   ██║██╔██╗ ██║
   ██╔══██║   ██║      ██║   ██╔══╝  ██║╚██╗██║   ██║   ██║██║   ██║██║╚██╗██║
   ██║  ██║   ██║      ██║   ███████╗██║ ╚████║   ██║   ██║╚██████╔╝██║ ╚████║
   ╚═╝  ╚═╝   ╚═╝      ╚═╝   ╚══════╝╚═╝  ╚═══╝   ╚═╝   ╚═╝ ╚═════╝ ╚═╝  ╚═══╝
																			  
   ============================================================================`
}
