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

type SingleResult struct {
	value interface{}
	err   error
}

type Result struct {
	value interface{}
}

type ResultsObjectArray []ResultsObject

type ResultsObject map[string]Result

//lint:ignore U1000 Ignore unused function temporarily for debugging
func db_select(query string, args ...any) (ResultsObjectArray, error) {
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

	columnLength := len(columns)

	values := make([]interface{}, columnLength)
	valuePtrs := make([]interface{}, columnLength)
	for i := range columns {
		valuePtrs[i] = &values[i]
	}

	resultsArray := make(ResultsObjectArray, 0)

	for rows.Next() {
		scanValues := make([]interface{}, columnLength)
		copy(scanValues, valuePtrs)

		err := rows.Scan(scanValues...)
		if err != nil {
			return nil, err
		}

		resultsObject := make(ResultsObject)
		for i, column := range columns {
			resultsObject[column] = Result{value: values[i]}
		}

		resultsArray = append(resultsArray, resultsObject)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return resultsArray, nil
}

func (r SingleResult) toString() (string, error) {
	if r.err != nil {
		return "", r.err
	}
	return toString(r.value)
}

func (r SingleResult) toInt() (int, error) {
	if r.err != nil {
		return 0, r.err
	}
	return toInt(r.value)
}

func (r SingleResult) toBool() (bool, error) {
	if r.err != nil {
		return false, r.err
	}
	return toBool(r.value)
}

//lint:ignore U1000 Ignore unused function temporarily for debugging
func (r Result) toString() (string, error) {
	return toString(r.value)
}

//lint:ignore U1000 Ignore unused function temporarily for debugging
func (r Result) toInt() (int, error) {
	return toInt(r.value)
}

//lint:ignore U1000 Ignore unused function temporarily for debugging
func (r Result) toBool() (bool, error) {
	return toBool(r.value)
}

func toString(value interface{}) (string, error) {
	if value == nil {
		return "", nil
	}
	if str, ok := value.(string); ok {
		return str, nil
	}
	return "", fmt.Errorf("failed to convert result to string")
}

func toInt(value interface{}) (int, error) {
	if value == nil {
		return 0, nil
	}
	if num, ok := value.(int64); ok {
		return int(num), nil
	}
	return 0, fmt.Errorf("failed to convert result to int")
}

func toBool(value interface{}) (bool, error) {
	switch v := value.(type) {
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
	case nil:
		return false, nil
	}
	return false, fmt.Errorf("failed to convert result to bool")
}

func db_single_item_query(query string, args ...any) SingleResult {
	db := establish_database_connection()
	defer db.Close()

	var result interface{}
	err := db.QueryRow(query, args...).Scan(&result)
	if err != nil {
		if err == sql.ErrNoRows {
			return SingleResult{value: nil, err: nil}
		}
		return SingleResult{value: nil, err: err}
	}

	return SingleResult{value: result, err: nil}
}

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
