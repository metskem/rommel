package db

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/metskem/rommel/joeradiostats/conf"
	"log"
	"net/url"
	"os"
)

var Database *sql.DB

func Initdb() {
	var DbExists bool
	var err error
	dbURL, err := url.Parse(conf.DatabaseURL)
	if err != nil {
		log.Fatalf("failed parsing database url %s, error: %s", conf.DatabaseURL, err.Error())
	}
	if _, err = os.Stat(dbURL.Opaque); err != nil && dbURL.Scheme == "file" {
		log.Printf("database %s does not exist, creating it...\n", dbURL.Path)
		DbExists = false
	} else {
		DbExists = true
	}

	Database, err = sql.Open("sqlite3", conf.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}

	// a simple but effective way to not get "database is locked" from sqlite3
	Database.SetMaxOpenConns(1)

	if !DbExists {
		var sqlStmts []byte
		if sqlStmts, err = os.ReadFile(conf.CreateTablesFile); err != nil {
			log.Fatal(err)
		} else {
			if _, err = Database.Exec(string(sqlStmts)); err != nil {
				log.Fatalf("%q: %s\n", err, sqlStmts)
			}
		}
	}
}
