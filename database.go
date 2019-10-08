package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/joho/godotenv/autoload"
)

var (
	// Db is the open database connection
	Db  *sql.DB
	err error
)

// OpenDatabase opens the database connection
func OpenDatabase() {
	Db, err = sql.Open("mysql", os.Getenv("MYSQL_DSN"))
	if err != nil {
		log.Fatal(err)
	}
}
