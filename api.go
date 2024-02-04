package main

import (
	"net/http"
)

func authCheckHandler(w http.ResponseWriter, r *http.Request) {
	is_authenticated := get_and_validate_jwt(r)
	if !is_authenticated {
		http.Error(w, "Not authenticated", http.StatusUnauthorized)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func settingsHandler(w http.ResponseWriter, r *http.Request) {
	is_authenticated := get_and_validate_jwt(r)
	if !is_authenticated {
		http.Error(w, "Not authenticated", http.StatusUnauthorized)
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	db := establish_database_connection()
	defer db.Close()

	var password string
	db.QueryRow("SELECT value FROM settings WHERE key = ?", ADMIN_PASSWORD_SETTINGS_KEY).Scan(&password)

	if password == "" {
		http.Error(w, "No password set", http.StatusInternalServerError)
		return
	}

	if check_hash(r.FormValue("password"), password) {
		generate_and_set_jwt(w)
	}
}

func payloadFiresHandler(w http.ResponseWriter, r *http.Request) {
	is_authenticated := get_and_validate_jwt(r)
	if !is_authenticated {
		http.Error(w, "Not authenticated", http.StatusUnauthorized)
	}
}

func collected_pages(w http.ResponseWriter, r *http.Request) {
	is_authenticated := get_and_validate_jwt(r)
	if !is_authenticated {
		http.Error(w, "Not authenticated", http.StatusUnauthorized)
	}
}
