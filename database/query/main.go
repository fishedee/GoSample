package main

import (
	"database/sql"
	_ "github.com/Go-SQL-Driver/MySQL" 
	"fmt"
	"log"
)

func main() {
	db, err := sql.Open("mysql", "root:1@/bakeweb")
	if err != nil {
		log.Fatal("Open database error: %s\n", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	rows, err := db.Query("select clientId, name from t_client")
	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()

	var clientId int
	var name string
	for rows.Next() {
		err := rows.Scan(&clientId, &name)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(clientId, name)
	}

	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
}