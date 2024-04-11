package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

type UserXSSPayloads struct {
	ID          uint   `json:"id"`
	Payload     string `json:"payload"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Author      string `json:"author"`
	Author_link string `json:"author_link"`
}

func authCheckHandler(w http.ResponseWriter, r *http.Request) {
	set_secure_headers(w, r)
	is_authenticated := get_and_validate_jwt(r)
	if !is_authenticated {
		http.Error(w, "Not authenticated", http.StatusUnauthorized)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func settingsHandler(w http.ResponseWriter, r *http.Request) {
	set_secure_headers(w, r)
	is_authenticated := get_and_validate_jwt(r)
	if !is_authenticated {
		http.Error(w, "Not authenticated", http.StatusUnauthorized)
	}
	if r.Method == "GET" {
		db := establish_database_connection()
		defer db.Close()

		rows, err := db.Query("SELECT key, value FROM settings WHERE key IN ($1, $2, $3, $4)", CORRELATION_API_SECRET_SETTINGS_KEY, CHAINLOAD_URI_SETTINGS_KEY, PAGES_TO_COLLECT_SETTINGS_KEY, SEND_ALERTS_SETTINGS_KEY)
		if err != nil {
			http.Error(w, "Error querying database", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		settings := map[string]string{}
		for rows.Next() {
			var key, value string
			err = rows.Scan(&key, &value)
			if err != nil {
				http.Error(w, "Error scanning database", http.StatusInternalServerError)
				return
			}
			settings[key] = value
		}
		settings[ADMIN_PASSWORD_SETTINGS_KEY] = ""

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(settings)
		if err != nil {
			http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
			return
		}
	} else if r.Method == "PUT" {
		var setting_key = r.FormValue("key")
		var setting_value = r.FormValue("value")
		if setting_key == "" {
			http.Error(w, "Invalid key", http.StatusBadRequest)
			return
		}

		switch setting_key {
		case ADMIN_PASSWORD_SETTINGS_KEY:
			hashed_password, err := hash_string(setting_value)
			if err != nil {
				http.Error(w, "Error hashing password", http.StatusInternalServerError)
				return
			}
			update_setting(ADMIN_PASSWORD_SETTINGS_KEY, hashed_password)
		case CORRELATION_API_SECRET_SETTINGS_KEY:
			update_setting(CORRELATION_API_SECRET_SETTINGS_KEY, setting_value)
		case CHAINLOAD_URI_SETTINGS_KEY:
			update_setting(CHAINLOAD_URI_SETTINGS_KEY, setting_value)
			set_chainload_uri()
		case PAGES_TO_COLLECT_SETTINGS_KEY:
			update_setting(PAGES_TO_COLLECT_SETTINGS_KEY, setting_value)
			set_pages_to_collect()
		case SEND_ALERTS_SETTINGS_KEY:
			update_setting(SEND_ALERTS_SETTINGS_KEY, setting_value)
			set_send_alerts()
		default:
			http.Error(w, "Invalid key", http.StatusBadRequest)
		}
	} else {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	set_secure_headers(w, r)
	if r.Method != "POST" {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}

	password, err := db_query_row("SELECT value FROM settings WHERE key = $1", ADMIN_PASSWORD_SETTINGS_KEY).toString()
	if err != nil {
		http.Error(w, "Error querying database", http.StatusInternalServerError)
		return
	}

	if password == "" {
		http.Error(w, "No password set", http.StatusInternalServerError)
		return
	}

	if check_hash(r.FormValue("password"), password) {
		generate_and_set_jwt(w)
		w.WriteHeader(http.StatusOK)
	} else {
		http.Error(w, "Invalid password", http.StatusUnauthorized)
	}
}

func payloadFiresHandler(w http.ResponseWriter, r *http.Request) {
	set_secure_headers(w, r)
	is_authenticated := get_and_validate_jwt(r)
	if !is_authenticated {
		http.Error(w, "Not authenticated", http.StatusUnauthorized)
	}
	if r.Method == "GET" {
		page_string := r.URL.Query().Get("page")
		limit_string := r.URL.Query().Get("limit")
		page := parameter_to_int(page_string, 1) - 1
		limit := parameter_to_int(limit_string, 10)
		offset := page * limit

		db := establish_database_connection()
		defer db.Close()

		rows, err := db.Query("SELECT id, url, ip_address, referer, user_agent, cookies, title, dom, text, origin, screenshot_id, was_iframe, browser_timestamp FROM payload_fire_results ORDER BY created_at DESC LIMIT $1 OFFSET $2", limit, offset)
		if err != nil {
			http.Error(w, "Error querying database", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		payload_fires := []PayloadFireResults{}
		for rows.Next() {
			var payload_fire PayloadFireResults
			err = rows.Scan(&payload_fire.ID, &payload_fire.Url, &payload_fire.Ip_address, &payload_fire.Referer, &payload_fire.User_agent, &payload_fire.Cookies, &payload_fire.Title, &payload_fire.Dom, &payload_fire.Text, &payload_fire.Origin, &payload_fire.Screenshot_id, &payload_fire.Was_iframe, &payload_fire.Browser_timestamp)
			if err != nil {
				http.Error(w, "Error scanning database", http.StatusInternalServerError)
				return
			}
			payload_fires = append(payload_fires, payload_fire)
		}
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(payload_fires)
		if err != nil {
			http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
			return
		}
	} else if r.Method == "DELETE" {
		ids_to_delete := r.FormValue("ids")
		if len(ids_to_delete) == 0 {
			http.Error(w, "No ids to delete", http.StatusBadRequest)
			return
		}
		db := establish_database_connection()
		defer db.Close()

		rows, err := db.Query("SELECT screenshot_id FROM payload_fire_results WHERE id IN ($1)", ids_to_delete)
		if err != nil {
			http.Error(w, "Error querying database", http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		for rows.Next() {
			var screenshot_id string
			err = rows.Scan(&screenshot_id)
			if err != nil {
				http.Error(w, "Error scanning database", http.StatusInternalServerError)
				return
			}
			payload_fire_image_filename := get_screenshot_directory() + "/" + screenshot_id + ".png.gz"
			err = os.Remove(payload_fire_image_filename)
			if err != nil {
				http.Error(w, "Error deleting payload fire image", http.StatusInternalServerError)
				return
			}
			_, err = db_execute("DELETE FROM payload_fire_results WHERE screenshot_id = $1", screenshot_id)
			if err != nil {
				http.Error(w, "Error deleting payload fires", http.StatusInternalServerError)
				return
			}
		}
	} else {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
	}
}

func collectedPagesHandler(w http.ResponseWriter, r *http.Request) {
	set_secure_headers(w, r)
	is_authenticated := get_and_validate_jwt(r)
	if !is_authenticated {
		http.Error(w, "Not authenticated", http.StatusUnauthorized)
	}
	if r.Method == "GET" {
		page_string := r.URL.Query().Get("page")
		limit_string := r.URL.Query().Get("limit")
		page := parameter_to_int(page_string, 1) - 1
		limit := parameter_to_int(limit_string, 10)
		offset := page * limit

		db := establish_database_connection()
		defer db.Close()

		rows, err := db.Query("SELECT id, uri FROM collected_pages ORDER BY created_at DESC LIMIT $1 OFFSET $2", limit, offset)
		if err != nil {
			http.Error(w, "Error querying database", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		collected_pages := []CollectedPages{}
		for rows.Next() {
			var collected_page CollectedPages
			err = rows.Scan(&collected_page.ID, &collected_page.Uri)
			if err != nil {
				http.Error(w, "Error scanning database", http.StatusInternalServerError)
				return
			}
			collected_pages = append(collected_pages, collected_page)
		}
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(collected_pages)
		if err != nil {
			http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
			return
		}
	} else if r.Method == "DELETE" {
		ids_to_delete := r.FormValue("ids")
		if len(ids_to_delete) == 0 {
			http.Error(w, "No ids to delete", http.StatusBadRequest)
			return
		}
		_, err := db_execute("DELETE FROM collected_pages WHERE id IN ($1)", ids_to_delete)
		if err != nil {
			http.Error(w, "Error deleting collected pages", http.StatusInternalServerError)
			return
		}
	} else {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
	}
}

func recordInjectionHandler(w http.ResponseWriter, r *http.Request) {
	set_secure_headers(w, r)

	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		log.Fatal(err)
	}
	owner_correlation_key := r.FormValue("owner_correlation_key")
	if owner_correlation_key == "" {
		http.Error(w, "No owner_correlation_key", http.StatusBadRequest)
		return
	}

	is_authenticated, errQuery := db_query_row("SELECT 1 FROM settings WHERE key = $1 AND value = $2", CORRELATION_API_SECRET_SETTINGS_KEY, owner_correlation_key).toBool()
	if errQuery != nil {
		fmt.Println("Error querying database: ", errQuery)
		http.Error(w, "Error", http.StatusUnauthorized)
		return
	}

	if !is_authenticated {
		http.Error(w, "Not authenticated", http.StatusUnauthorized)
	}
	if r.Method != "POST" {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}

	_, err = db_execute("INSERT INTO injection_requests (injection_key, request) VALUES ($1, $2)", r.FormValue("injection_key"), r.FormValue("request"))
	if err != nil {
		http.Error(w, "Error inserting injection request", http.StatusInternalServerError)
		return
	}

	_, err = w.Write([]byte("Injection recorded"))
	if err != nil {
		http.Error(w, "Error writing response", http.StatusInternalServerError)
	}
}

func userPayloadsHandler(w http.ResponseWriter, r *http.Request) {
	set_secure_headers(w, r)
	is_authenticated := get_and_validate_jwt(r)
	if !is_authenticated {
		http.Error(w, "Not authenticated", http.StatusUnauthorized)
	}
	if r.Method == "GET" {
		// page_string := r.URL.Query().Get("page")
		// limit_string := r.URL.Query().Get("limit")
		// page := parameter_to_int(page_string, 1) - 1
		// limit := parameter_to_int(limit_string, 10)
		// offset := page * limit

		db := establish_database_connection()
		defer db.Close()

		rows, err := db.Query("SELECT id, payload, title, description, author, author_link FROM user_xss_payloads ORDER BY created_at ASC")
		if err != nil {
			http.Error(w, "Error querying database", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		user_payloads := []UserXSSPayloads{}
		for rows.Next() {
			var user_payload UserXSSPayloads
			err = rows.Scan(&user_payload.ID, &user_payload.Payload, &user_payload.Title, &user_payload.Description, &user_payload.Author, &user_payload.Author_link)
			if err != nil {
				http.Error(w, "Error scanning database", http.StatusInternalServerError)
				return
			}
			user_payloads = append(user_payloads, user_payload)
		}
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(user_payloads)
		if err != nil {
			http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
			return
		}

	} else if r.Method == "POST" {
		db := establish_database_connection()
		defer db.Close()

		stmt, _ := db.Prepare(`INSERT INTO user_xss_payloads (payload, title, description, author, author_link) VALUES ($1, $2, $3, $4, $5)`)
		_, err := stmt.Exec(r.FormValue("payload"), r.FormValue("title"), r.FormValue("description"), r.FormValue("author"), r.FormValue("author_link"))
		if err != nil {
			http.Error(w, "Error inserting user payload", http.StatusInternalServerError)
			return
		}
	} else if r.Method == "DELETE" {
		ids_to_delete := r.FormValue("ids")
		if len(ids_to_delete) == 0 {
			http.Error(w, "No ids to delete", http.StatusBadRequest)
			return
		}
		_, err := db_execute("DELETE FROM user_xss_payloads WHERE id IN ($1)", ids_to_delete)
		if err != nil {
			http.Error(w, "Error deleting user payloads", http.StatusInternalServerError)
			return
		}
	} else {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
	}
}

func userPayloadImporterHandler(w http.ResponseWriter, r *http.Request) {
	set_secure_headers(w, r)
	is_authenticated := get_and_validate_jwt(r)
	if !is_authenticated {
		http.Error(w, "Not authenticated", http.StatusUnauthorized)
	}
	if r.Method != "POST" {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}
	db := establish_database_connection()
	defer db.Close()

	var user_payloads []UserXSSPayloads
	err := json.NewDecoder(r.Body).Decode(&user_payloads)
	if err != nil {
		http.Error(w, "Error decoding JSON", http.StatusInternalServerError)
		return
	}
	for _, user_payload := range user_payloads {
		stmt, _ := db.Prepare(`INSERT INTO user_xss_payloads (payload, title, description, author, author_link) VALUES ($1, $2, $3, $4, $5)`)
		_, err := stmt.Exec(user_payload.Payload, user_payload.Title, user_payload.Description, user_payload.Author, user_payload.Author_link)
		if err != nil {
			http.Error(w, "Error inserting user payload", http.StatusInternalServerError)
			return
		}
	}
}
