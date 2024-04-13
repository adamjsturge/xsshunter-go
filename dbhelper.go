package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

type Result struct {
	value interface{}
	err   error
}

type ResultRow struct {
	values []interface{}
	err    error
}

func (r Result) toString() (string, error) {
	if r.err != nil {
		return "", r.err
	}
	if str, ok := r.value.(string); ok {
		return str, nil
	}
	return "", fmt.Errorf("failed to convert result to string")
}

func (r Result) toInt() (int, error) {
	if r.err != nil {
		return 0, r.err
	}
	if num, ok := r.value.(int64); ok {
		return int(num), nil
	}
	return 0, fmt.Errorf("failed to convert result to int")
}

func (r Result) toBool() (bool, error) {
	if r.err != nil {
		return false, r.err
	}
	switch v := r.value.(type) {
	case bool:
		return v, nil
	case string:
		lower := strings.ToLower(v)
		if lower == "true" || lower == "1" {
			return true, nil
		} else if lower == "false" || lower == "0" || lower == "" {
			return false, nil
		}
	case int:
		return v == 1, nil
	case int64:
		return v == 1, nil
	}
	return false, fmt.Errorf("failed to convert result to bool")
}

func (r Result) toMap() (map[string]Result, error) {
	if r.err != nil {
		return nil, r.err
	}
	if m, ok := r.value.(map[string]Result); ok {
		return m, nil
	}
	return nil, fmt.Errorf("failed to convert result to map")
}

func db_single_item_query(query string, args ...any) Result {
	db := establish_database_connection()
	defer db.Close()

	var result interface{}
	err := db.QueryRow(query, args...).Scan(&result)
	if err != nil {
		if err == sql.ErrNoRows {
			return Result{err: nil}
		}
		return Result{err: err}
	}
	return Result{value: result}
}

func db_multi_item_query(query string, args ...interface{}) ([]map[string]Result, error) {
	db := establish_database_connection()
	defer db.Close()

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []map[string]Result
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))

		for i := 0; i < len(columns); i++ {
			valuePtrs[i] = &values[i]
		}

		err = rows.Scan(valuePtrs...)
		if err != nil {
			return nil, err
		}

		result := make(map[string]Result)
		for i, col := range columns {
			val := valuePtrs[i].(*interface{})
			result[col] = Result{value: *val}
		}

		results = append(results, result)
	}

	return results, nil
}

// func db_query(query string, args ...any) (any, error) {
// 	db := establish_database_connection()
// 	defer db.Close()

// 	var result any
// 	rows, err := db.Query(query, args)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer rows.Close()

// 	for rows.Next() {
// 		err = rows.Scan(&result)
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 	}

// 	return result, err
// }

func db_prepare_execute(query string, args ...any) (sql.Result, error) {
	db := establish_database_connection()
	defer db.Close()

	stmt, _ := db.Prepare(query)
	result, err := stmt.Exec(args...)
	if err != nil {
		log.Printf("%q: %s\n", err, query)
	}

	return result, err
}

func db_execute(query string, args ...any) (sql.Result, error) {
	db := establish_database_connection()
	defer db.Close()

	result, err := db.Exec(query, args...)
	if err != nil {
		log.Printf("%q: %s\n", err, query)
	}

	return result, err
}

func initialize_database() {
	if is_postgres {
		initialize_postgres_database()
	} else {
		initialize_sqlite_database()
	}

	do_migrations()
	initialize_settings()
}

func establish_database_connection() *sql.DB {
	if is_postgres {
		return establish_postgres_database_connection()
	}
	return establish_sqlite_database_connection()
}

func initialize_sqlite_database() {
	if _, err := os.Stat("db"); os.IsNotExist(err) {
		err = os.MkdirAll("db", 0750)
		if err != nil {
			log.Fatal(err)
		}
	}
	create_sqlite_tables()
}

func initialize_postgres_database() {
	create_postgres_tables()
}

func establish_sqlite_database_connection() *sql.DB {
	dbPath := get_sqlite_database_path()
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func establish_postgres_database_connection() *sql.DB {
	db, err := sql.Open("postgres", get_env("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	return db
}
