package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
)

type Result struct {
	value interface{}
	err   error
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

func db_query_row(query string, args ...any) Result {
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

// func db_query(query string, args ...any) (any, error) {
// 	db := establish_database_connection()
// 	defer db.Close()

// 	var result any
// 	err := db.QueryRow(query, args).Scan(&result)
// 	if err != nil {
// 		log.Fatal(err)
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
