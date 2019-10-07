package main

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

var (
	// Db is the open database connection
	Db  *sql.DB
	err error
)

// OpenDatabase opens the database connection
func OpenDatabase() {
	Db, err = sql.Open("mysql", "root@tcp(localhost)/parkleaks")
	if err != nil {
		log.Fatal(err)
	}
}
