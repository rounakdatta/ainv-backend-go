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

type Item struct {
	Name string `json:"name"`
	Description []string `json:"description"`
	ItemId []string `json:"itemId"`
}

type searchPayload struct {
	Ids string
	Locations string
}

type ItemInventory struct {
	ItemName string `json:"itemName"`
	ItemVariant string `json:"itemVariant"`
	HsnCode string `json:"hsnCode"`
	ItemQuantity string `json:"itemQuantity"`
	UomRaw string `json:"uomRaw"`
	SmallboxQuantity string `json:"smallboxQuantity"`
	UomSmall string `json:"uomSmall"`
	BigcartonQuantity string `json:"bigcartonQuantity"`
	UomBig string `json:"uomBig"`
	WarehouseName string `json:"warehouseName"`
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
	router.HandleFunc("/api/get/items", GetItems).Methods("GET")
	router.HandleFunc("/api/search/items", SearchItems).Methods("POST")
	router.HandleFunc("/api/put/warehouse", CreateWarehouse).Methods("POST")
	router.HandleFunc("/api/put/itemmaster", CreateItemMaster).Methods("POST")

	http.Handle("/", router)

	log.Println("Server started on port 80")
	log.Fatal(http.ListenAndServe(":80", nil))
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

// GetItems returns all the items with description and ID
func GetItems(w http.ResponseWriter, r *http.Request) {

	requestedParameter, ok := r.URL.Query()["only"]

	// case of special parameter requested
	if ok && requestedParameter != nil {
		getSpecificQuery := fmt.Sprintf(`SELECT DISTINCT 
		%s FROM itemMaster`, requestedParameter[0])

		specificItems, err := db.Query(getSpecificQuery)
		if err != nil {
			panic(err.Error())
		}

		var payload []string

		for specificItems.Next() {
			var warehouseName string

			err := specificItems.Scan(&warehouseName)
			if err != nil {
				panic(err.Error())
			}

			payload = append(payload, warehouseName)
		}

		payloadJSON, err := json.Marshal(payload)
		if err != nil {
			log.Println(err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(payloadJSON)

		return
	}

	var payload []Item

	getItemDetailsQuery := `SELECT
		itemName,
		GROUP_CONCAT(itemVariant SEPARATOR '$') itemVariant,
		GROUP_CONCAT(itemId SEPARATOR '$') itemId
	FROM
		itemMaster
	GROUP BY
		itemName`

	allItems, err := db.Query(getItemDetailsQuery)
	if err != nil {
		panic(err.Error())
	}

	for allItems.Next() {
		var name string
		var description string
		var itemId string

		err := allItems.Scan(&name, &description, &itemId)
		if err != nil {
			panic(err.Error())
		}

		singleObject := Item{
			Name: name,
			Description: strings.Split(description, "$"),
			ItemId: strings.Split(itemId, "$"),
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

// CreateWarehouse creates a new warehouse and returns the status
func CreateWarehouse(w http.ResponseWriter, r *http.Request) {

	warehouseName := r.FormValue("warehouseName")
	warehouseLocation := r.FormValue("warehouseLocation")
	gstin := r.FormValue("gstin")
	contactName := r.FormValue("contactName")
	contactNumber := r.FormValue("contactNumber")

	warehouseInsertQuery := fmt.Sprintf(`INSERT INTO warehouse
		(warehouseName, warehouseLocation, gstin, contactName, contactNumber)
		VALUES
		('%s', '%s', '%s', '%s', '%s')`, warehouseName, warehouseLocation, gstin, contactName, contactNumber)

	_, err := db.Query(warehouseInsertQuery)

	var result map[string]bool

	if err != nil {
		result = map[string]bool {
			"success": false,
		}
	} else {
		result = map[string]bool {
			"success": true,
		}
	}

	payloadJSON, err := json.Marshal(result)
	if err != nil {
		log.Println(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(payloadJSON)
}

// CreateItemMaster creates a new item and returns the status
func CreateItemMaster(w http.ResponseWriter, r *http.Request) {

	itemName := r.FormValue("itemName")
	itemVariant := r.FormValue("itemVariant")
	hsnCode := r.FormValue("hsnCode")
	uomRaw := r.FormValue("uomRaw")
	uomSmall := r.FormValue("uomSmall")
	uomBig := r.FormValue("uomBig")
	rawPerSmall := r.FormValue("rawPerSmall")
	smallPerBig := r.FormValue("smallPerBig")

	itemInsertQuery := fmt.Sprintf(`INSERT INTO itemMaster
	(itemName, itemVariant, hsnCode, uomRaw, uomSmall, uomBig, rawPerSmall, smallPerBig)
	VALUES
	('%s', '%s', '%s, '%s', '%s', '%s', %s, %s)`, itemName, itemVariant, hsnCode, uomRaw, uomSmall, uomBig, rawPerSmall, smallPerBig)

	_, err := db.Query(itemInsertQuery)

	var result map[string]bool

	if err != nil {
		result = map[string]bool {
			"success": false,
		}
	} else {
		result = map[string]bool {
			"success": true,
		}
	}

	payloadJSON, err := json.Marshal(result)
	if err != nil {
		log.Println(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(payloadJSON)
}

// SearchItems searches for an item by id and location
func SearchItems(w http.ResponseWriter, r *http.Request) {

	requestedItemIdRaw := r.FormValue("itemId")
	requestedLocationsRaw := r.FormValue("locations")

	requestedItemId := strings.Split(strings.TrimSpace(requestedItemIdRaw), " ")
	requestedLocations := strings.Split(strings.TrimSpace(requestedLocationsRaw), " ")

	items := strings.Join(requestedItemId, ", ")
	locations := strings.Join(requestedLocations, ", ")

	var payload []ItemInventory

	searchQuery := fmt.Sprintf(`SELECT 
		itm.itemName, itm.itemVariant, itm.hsnCode, inv.itemQuantity, itm.uomRaw, inv.smallboxQuantity, itm.uomSmall, inv.bigcartonQuantity, itm.uomBig, wh.warehouseName, wh.warehouseLocation
		FROM inventoryContents inv, itemMaster itm, warehouse wh
		WHERE inv.itemId IN (%s) AND
		inv.itemId = itm.itemId AND
		inv.warehouseId = wh.warehouseId AND
		wh.warehouseId IN (%s)`, items, locations)

	allContents, err := db.Query(searchQuery)
	if err != nil {
		panic(err.Error())
	}

	for allContents.Next() {
		var itemName string
		var itemVariant string
		var hsnCode string
		var itemQuantity string
		var uomRaw string
		var smallboxQuantity string
		var uomSmall string
		var bigcartonQuantity string
		var uomBig string
		var warehouseName string
		var warehouseLocation string

		err := allContents.Scan(&itemName, &itemVariant, &hsnCode, &itemQuantity, &uomRaw, &smallboxQuantity, &uomSmall, &bigcartonQuantity, &uomBig, &warehouseName, &warehouseLocation)
		if err != nil {
			panic(err.Error())
		}

		singleObject := ItemInventory{
			ItemName: itemName,
			ItemVariant: itemVariant,
			HsnCode: hsnCode,
			ItemQuantity: itemQuantity,
			UomRaw: uomRaw,
			SmallboxQuantity: smallboxQuantity,
			UomSmall: uomSmall,
			BigcartonQuantity: bigcartonQuantity,
			UomBig: uomBig,
			WarehouseName: warehouseName,
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