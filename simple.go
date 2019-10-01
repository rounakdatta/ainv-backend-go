package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	fmt.Println("Connecting to database")

	databaseCredentials := fmt.Sprintf("%s:%s@/%s", os.Getenv("APP_USER"), os.Getenv("APP_PASSWORD"), os.Getenv("APP_DATABASE"))
	fmt.Println(databaseCredentials)
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
