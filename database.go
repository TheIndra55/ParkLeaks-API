package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

var (
	// Db is the open database connection
	Db *sql.DB
	// Names is a list of possible names for new users
	Names []string
	err   error
)

// OpenDatabase opens the database connection
func OpenDatabase() {
	Db, err = sql.Open("mysql", os.Getenv("MYSQL_DSN"))
	if err != nil {
		log.Fatal(err)
	}

	Db.SetMaxOpenConns(20)
}

// GetNames fetches all names from database and stores them in memory
func GetNames() {
	rows, err := Db.Query("SELECT `name` from `names`")
	defer rows.Close()

	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {
		var name string
		rows.Scan(&name)

		Names = append(Names, name)
	}
}
