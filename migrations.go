package main

func do_migrations() {
	// migration_one()

}

// func check_if_migrations_has_ran(migration_name string) bool {
// 	has_migration, err := db_query_row("SELECT COUNT(1) FROM migrations WHERE name = $1", migration_name).toInt()
// 	if err != nil {
// 		log.Fatal(err)
// 		return false
// 	}

// 	return has_migration == 1
// }

// func record_migration(migration_name string) {

// }

// func migration_handler(name string, pgStmt string, sqliteStmt string) {
// 	if check_if_migrations_has_ran(name) {
// 		return
// 	}

// 	var sqlStmt string
// 	if is_postgres {
// 		sqlStmt = pgStmt
// 	} else {
// 		sqlStmt = sqliteStmt
// 	}

// 	_, err := db_execute(sqlStmt)
// 	if err != nil {
// 		log.Fatal("Migration ", err)
// 	}

// 	record_migration(name)
// }

// func migration_one() {
// 	name := "20240410_add_injection_request_id"

// 	pgStmt := `

// 	`

// 	sqliteStmt := `

// 	`

// 	migration_handler(name, pgStmt, sqliteStmt)
// }
