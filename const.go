package main

func get_pages_to_collect() []string {

}

func get_chainload_uri() []string {
	db := establish_database_connection()

}

func get_screenshot_directory() string {
	return get_env("SCREENSHOT_DIRECTORY")
}
