package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "postgres"
	dbname   = "postgres"
)

var psqlInfo = fmt.Sprintf("host=%s port=%d user=%s "+
	"password=%s dbname=%s sslmode=disable",
	host, port, user, password, dbname)

type Database struct {
	Client *sql.DB
}

var single *Database

//singleton de cliente base de datos
func (d Database) GetClient() *Database {
	if single == nil {
		db, err := sql.Open("postgres", psqlInfo)
		if err != nil {
			panic(err)
		}
		db.SetMaxOpenConns(30)
		db.SetMaxIdleConns(30)
		db.SetConnMaxLifetime(0)
		single = &Database{Client: db}
	}

	return single
}
