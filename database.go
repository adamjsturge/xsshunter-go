package main

import (
	"log"
	"os"
	"fmt"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Settings struct {
	ID uint
	key *string
	value *string
}

type PayloadFireResults struct {
	ID uint
	url string
	ip_address string
	referer string
	user_agent string
	cookies string
	title string
	dom string
	text string
	origin string
	screenshot_id string
	was_iframe bool
	browser_timestamp uint
}

type CollectedPages struct {
	ID uint
	uri string
	html string
}

type InjectionRequests struct {
	ID uint
	request string
	injection_key string
}

func establish_database_connection() *gorm.DB {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	postgres_db := os.Getenv("POSTGRES_DSN")
	if postgres_db != "" {
		return establish_postgres_connection()
	} else {
		return establish_sqlite_connection()
	}
}

func establish_postgres_connection() *gorm.DB {
	dsn := os.Getenv("POSTGRES_DSN")

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect to database")
	}
	return db
}

func establish_sqlite_connection() *gorm.DB {
	sqlite_path := os.Getenv("SQLITE_PATH")
	db, err := gorm.Open(sqlite.Open(sqlite_path), &gorm.Config{})
	if err != nil {
		panic("failed to connect to database")
	}
	return db
}

func initialize_users() {
	new_password := get_secure_random_string(32)

	new_user := setup_admin_user(new_password)

	if new_user {
		return
	}

	banner_message := get_default_user_created_banner(new_password)
	fmt.Println(banner_message)
}

func setup_admin_user(password string) {

}

func get_default_user_created_banner(password string) string {
	return `============================================================================
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
   
	   XSS Hunter Express has only one user for the instance. Do not
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
