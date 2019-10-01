package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	fmt.Println("Connecting to database")

	databaseCredentials := fmt.Sprintf("%s:%s/%s", os.Getenv("USER"), os.Getenv("PASSWORD"), os.Getenv("DATABASE"))
	db, err := sql.Open("mysql", databaseCredentials)

	if err != nil {
		panic(err.Error())
	}

	defer db.Close()

	results, err := db.Query("SHOW TABLES")
	if err != nil {
		panic(err.Error())
	}

	for results.Next() {
		var col string
		if err := results.Scan(&col); err != nil {
			panic(err.Error())
		}

		fmt.Println(col)
	}
}
