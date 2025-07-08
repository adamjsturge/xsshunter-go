package main

import "log"

func do_migrations() {
	migration_one()
	migration_two()

}

func check_if_migrations_has_ran(migration_name string) bool {
	has_migration, err := db_single_item_query("SELECT 1 FROM migrations WHERE name = $1", migration_name).toBool()
	if err != nil {
		log.Fatal("Fatal Error on check migration ran:", err)
		return false
	}

	return has_migration
}

func record_migration(migration_name string) {
	_, err := db_execute("INSERT INTO migrations (name) VALUES ($1)", migration_name)
	if err != nil {
		log.Fatal("Fatal Error on insert migration:", err)
	}
}

func migration_handler(name string, pgStmt string, sqliteStmt string) {
	if check_if_migrations_has_ran(name) {
		return
	}

	var sqlStmt string
	if is_postgres {
		sqlStmt = pgStmt
	} else {
		sqlStmt = sqliteStmt
	}

	_, err := db_execute(sqlStmt)
	if err != nil {
		log.Fatal("Migration ", err)
	}

	record_migration(name)
}

func migration_one() {
	name := "20240410_add_injection_request_id"

	pgStmt := `
		ALTER TABLE payload_fire_results ADD COLUMN injection_requests_id INTEGER DEFAULT NULL;
		ALTER TABLE payload_fire_results ADD CONSTRAINT fk_injection_requests_id FOREIGN KEY (injection_requests_id) REFERENCES injection_requests(id);
	`

	sqliteStmt := `
		ALTER TABLE payload_fire_results ADD COLUMN injection_requests_id INTEGER DEFAULT NULL;
	`

	migration_handler(name, pgStmt, sqliteStmt)
}

func migration_two() {
	name := "20241215_add_probe_id"

	pgStmt := `
		ALTER TABLE payload_fire_results ADD COLUMN probe_id TEXT DEFAULT '';
	`

	sqliteStmt := `
		ALTER TABLE payload_fire_results ADD COLUMN probe_id TEXT DEFAULT '';
	`

	migration_handler(name, pgStmt, sqliteStmt)
}
