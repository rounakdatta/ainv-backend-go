package main

import (
	"strings"
	"encoding/json"
	"database/sql"
	"fmt"
	"os"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

type Warehouse struct {
	WarehouseId []string `json:"warehouseId"`
	WarehouseLocation string `json:"warehouseLocation"`
}

var db *sql.DB

func main() {
	// connect to MySQL database
	err := godotenv.Load()

	databaseCredentials := fmt.Sprintf("%s:%s@/%s", os.Getenv("APP_USER"), os.Getenv("APP_PASSWORD"), os.Getenv("APP_DATABASE"))
	db, err = sql.Open("mysql", databaseCredentials)

	if err != nil {
		panic(err.Error())
	}

	defer db.Close()

	// create the router and define the APIs
	router := mux.NewRouter()

	router.HandleFunc("/", GetRoot).Methods("GET")
	router.HandleFunc("/api/get/warehouses", GetWarehouses).Methods("GET")
	// router.HandleFunc("/api/get/items", GetItems).Methods("GET")
	// router.HandleFunc("/api/search/items", SearchItems).Methods("POST")
	// router.HandleFunc("/api/put/warehouse", CreateWarehouse).Methods("POST")
	// router.HandleFunc("/api/put/itemmaster", CreateItemMaster).Methods("POST")

	http.Handle("/", router)

	log.Println("Server started on port 3000")
	log.Fatal(http.ListenAndServe(":3000", nil))
}

// GetRoot returns OK if server is alive
func GetRoot(w http.ResponseWriter, r *http.Request) {
	payload := []byte("OK")
	w.Write(payload)
}

// GetWarehouses returns all the warehouses with their ID
func GetWarehouses(w http.ResponseWriter, r *http.Request) {

	var payload []Warehouse

	getWhNamesQuery := `SELECT 
		warehouseLocation,
		GROUP_CONCAT(warehouseId SEPARATOR '$') warehouseId
		FROM warehouse
		GROUP BY warehouseLocation`

	err := db.Ping()
	if err != nil {
		fmt.Println(err)
	}

	allWh, err := db.Query(getWhNamesQuery)
	if err != nil {
		panic(err.Error())
	}

	for allWh.Next() {
		var warehouseLocation string
		var warehouseId string

		err := allWh.Scan(&warehouseLocation, &warehouseId)
		if err != nil {
			panic(err.Error())
		}

		singleObject := Warehouse{
			WarehouseId: strings.Split(warehouseId, "$"),
			WarehouseLocation: warehouseLocation,
		}

		payload = append(payload, singleObject)
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		log.Println(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(payloadJSON)
}
