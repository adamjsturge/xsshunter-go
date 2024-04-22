package main

import (
	"testing"
)

func BenchmarkRawQuery(b *testing.B) {
	db := establish_database_connection()
	defer db.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rows, err := db.Query("SELECT id, request FROM injection_requests WHERE injection_key = $1", "e46304368666d45eb27cde23e564e828f29e167b")
		if err != nil {
			b.Fatal(err)
		}
		defer rows.Close()

		for rows.Next() {
			var id int
			var request string
			err := rows.Scan(&id, &request)
			if err != nil {
				b.Fatal(err)
			}
		}
	}
}

func BenchmarkDbSelect(b *testing.B) {
	db := establish_database_connection()
	defer db.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		injection_requests, err := db_select("SELECT id, request FROM injection_requests WHERE injection_key = $1", "e46304368666d45eb27cde23e564e828f29e167b")
		if err != nil {
			b.Fatal(err)
		}

		_, _ = injection_requests[0]["id"].toInt()
		_, _ = injection_requests[0]["request"].toString()
	}
}

func BenchmarkDBRawQuery(b *testing.B) {
	query := "SELECT 1 FROM payload_fire_results WHERE id = ?"
	args := []interface{}{1}

	for i := 0; i < b.N; i++ {
		db := establish_database_connection()
		defer db.Close()

		var result bool
		err := db.QueryRow(query, args...).Scan(&result)
		if err != nil {
			panic(err)
		}

		_ = result
	}
}

func BenchmarkDbSingleItemQuery(b *testing.B) {
	query := "SELECT 1 FROM payload_fire_results WHERE id = ?"
	args := []interface{}{1}

	for i := 0; i < b.N; i++ {
		_, _ = db_single_item_query(query, args...).toBool()
	}
}
