package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

type Rate struct {
	RawPerSmall    string `json:"rawPerSmall"`
	SmallPerBig    string `json:"smallPerBig"`
	CartonQuantity string `json:"cartonQuantity"`
}

type Warehouse struct {
	WarehouseId       []string `json:"warehouseId"`
	WarehouseLocation string   `json:"warehouseLocation"`
}

type Client struct {
	ClientId   string `json:"clientId"`
	ClientName string `json:"clientName"`
}

type Customer struct {
	CustomerId   string `json:"customerId"`
	CustomerName string `json:"customerName"`
}

type WarehouseEntity struct {
	WarehouseId   string `json:"warehouseId"`
	WarehouseName string `json:"warehouseName"`
}

type Item struct {
	Name        string   `json:"name"`
	Description []string `json:"description"`
	ItemId      []string `json:"itemId"`
}

type searchPayload struct {
	Ids       string
	Locations string
}

type ItemInventory struct {
	ItemName          string `json:"itemName"`
	ItemVariant       string `json:"itemVariant"`
	HsnCode           string `json:"hsnCode"`
	ItemQuantity      string `json:"itemQuantity"`
	UomRaw            string `json:"uomRaw"`
	SmallboxQuantity  string `json:"smallboxQuantity"`
	UomSmall          string `json:"uomSmall"`
	BigcartonQuantity string `json:"bigcartonQuantity"`
	UomBig            string `json:"uomBig"`
	WarehouseName     string `json:"warehouseName"`
	WarehouseLocation string `json:"warehouseLocation"`
	ClientName        string `json:"clientName"`
}

type SalesTransaction struct {
	TransactionId     string  `json:"transactionId"`
	TrackingNumber    string  `json:"trackingNumber"`
	EntryDate         string  `json:"entryDate"`
	ItemId            string  `json:"itemId"`
	ItemName          string  `json:"itemName"`
	ItemVariant       string  `json:"itemVariant"`
	WarehouseId       string  `json:"warehouseId"`
	WarehouseName     string  `json:"warehouseName"`
	WarehouseLocation string  `json:"warehouseLocation"`
	ClientId          string  `json:"clientId"`
	ClientName        string  `json:"clientName"`
	CustomerId        string  `json:"customerId"`
	CustomerName      string  `json:"customerName"`
	ChangeStock       string  `json:"changeStock"`
	FinalStock        string  `json:"finalStock"`
	TotalPcs          string  `json:"totalPcs"`
	MaterialValue     string  `json:"materialValue"`
	GstValue          string  `json:"gstValue"`
	TotalValue        string  `json:"totalValue"`
	ValuePerPiece     float64 `json:"valuePerPiece"`
	IsPaid            string  `json:"isPaid"`
	PaidAmount        string  `json:"paidAmount"`
	PaymentDate       string  `json:"paymentDate"`
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
	ainvRouter := router.PathPrefix("/ainv").Subrouter()

	ainvRouter.HandleFunc("/", GetRoot).Methods("GET")

	ainvRouter.HandleFunc("/api/get/warehouses/", GetWarehouses).Methods("GET")
	ainvRouter.HandleFunc("/api/get/all/warehouses/", GetAllWarehouses).Methods("GET")
	ainvRouter.HandleFunc("/api/get/all/clients/", GetAllClients).Methods("GET")
	ainvRouter.HandleFunc("/api/get/all/customers/", GetAllCustomers).Methods("GET")
	ainvRouter.HandleFunc("/api/get/items/", GetItems).Methods("GET")
	ainvRouter.HandleFunc("/api/get/rate/", GetRate).Methods("POST")

	ainvRouter.HandleFunc("/api/put/warehouse/", CreateWarehouse).Methods("POST")
	ainvRouter.HandleFunc("/api/put/itemmaster/", CreateItemMaster).Methods("POST")
	ainvRouter.HandleFunc("/api/put/transaction/", CreateTransaction).Methods("POST")
	ainvRouter.HandleFunc("/api/put/client/", CreateClient).Methods("POST")
	ainvRouter.HandleFunc("/api/put/customer/", CreateCustomer).Methods("POST")

	ainvRouter.HandleFunc("/api/update/paidamount/", UpdatePaidAmount).Methods("POST")
	ainvRouter.HandleFunc("/api/update/paymentdate/", UpdatePaymentDate).Methods("POST")

	ainvRouter.HandleFunc("/api/search/items/", SearchItems).Methods("POST")
	ainvRouter.HandleFunc("/api/search/sales/", SearchSales).Methods("POST")

	ainvRouter.HandleFunc("/api/register/", RegisterUser).Methods("POST")
	ainvRouter.HandleFunc("/api/login/", LoginUser).Methods("POST")

	http.Handle("/", router)

	log.Println("Server started on port 1234")
	log.Fatal(http.ListenAndServe(":1234", nil))
}

// GetMD5Hash returns the MD5-hashed representation of a string
func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

// GetRoot returns OK if server is alive
func GetRoot(w http.ResponseWriter, r *http.Request) {
	payload := []byte("OK")
	w.Write(payload)
}

// GetWarehouses returns all the locations with their warehouse IDs
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
			WarehouseId:       strings.Split(warehouseId, "$"),
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

// GetAllWarehouses returns all the warehouses with their ID
func GetAllWarehouses(w http.ResponseWriter, r *http.Request) {

	var payload []WarehouseEntity

	getWhNamesQuery := `SELECT 
		warehouseId, CONCAT(warehouseName, ", ", warehouseLocation) AS warehouseName
		FROM warehouse`

	allWh, err := db.Query(getWhNamesQuery)
	if err != nil {
		panic(err.Error())
	}

	for allWh.Next() {
		var warehouseId string
		var warehouseName string

		err := allWh.Scan(&warehouseId, &warehouseName)
		if err != nil {
			panic(err.Error())
		}

		singleObject := WarehouseEntity{
			WarehouseId:   warehouseId,
			WarehouseName: warehouseName,
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

// GetAllClients returns all the clients with their ID
func GetAllClients(w http.ResponseWriter, r *http.Request) {

	var payload []Client

	getClientNamesQuery := `SELECT 
		id, clientName
		FROM client`

	allClients, err := db.Query(getClientNamesQuery)
	if err != nil {
		panic(err.Error())
	}

	for allClients.Next() {
		var clientId string
		var clientName string

		err := allClients.Scan(&clientId, &clientName)
		if err != nil {
			panic(err.Error())
		}

		singleObject := Client{
			ClientId:   clientId,
			ClientName: clientName,
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

// GetAllCustomers returns all the clients with their ID
func GetAllCustomers(w http.ResponseWriter, r *http.Request) {

	var payload []Customer

	getCustomerNamesQuery := `SELECT 
		id, customerName
		FROM customer`

	allCustomers, err := db.Query(getCustomerNamesQuery)
	if err != nil {
		panic(err.Error())
	}

	for allCustomers.Next() {
		var customerId string
		var customerName string

		err := allCustomers.Scan(&customerId, &customerName)
		if err != nil {
			panic(err.Error())
		}

		singleObject := Customer{
			CustomerId:   customerId,
			CustomerName: customerName,
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

// GetRate returns the rate for a particular item
func GetRate(w http.ResponseWriter, r *http.Request) {

	requestedItemId := r.FormValue("itemId")
	requestedWarehouseId := r.FormValue("warehouseId")
	requestedClientId := r.FormValue("clientId")
	var payload []Rate

	getRatesQuery := fmt.Sprintf(`SELECT im.rawPerSmall, im.smallPerBig, IFNULL(ic.bigcartonQuantity, 0) AS cartonQuantity
		FROM itemMaster im
		LEFT JOIN inventoryContents ic
		ON (im.itemId = ic.itemId AND ic.warehouseId = '%s' AND ic.clientId = '%s')
		WHERE im.itemId = '%s'`, requestedWarehouseId, requestedClientId, requestedItemId)

	log.Println(getRatesQuery)
	rateDetails, err := db.Query(getRatesQuery)
	if err != nil {
		panic(err.Error())
	}

	for rateDetails.Next() {
		var rawPerSmall string
		var smallPerBig string
		var cartonQuantity string

		err := rateDetails.Scan(&rawPerSmall, &smallPerBig, &cartonQuantity)
		if err != nil {
			panic(err.Error())
		}

		singleObject := Rate{
			RawPerSmall:    rawPerSmall,
			SmallPerBig:    smallPerBig,
			CartonQuantity: cartonQuantity,
		}

		payload = append(payload, singleObject)
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		log.Println(err)
	}

	log.Println(payload)
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
			Name:        name,
			Description: strings.Split(description, "$"),
			ItemId:      strings.Split(itemId, "$"),
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
		result = map[string]bool{
			"success": false,
		}
	} else {
		result = map[string]bool{
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
	('%s', '%s', '%s', '%s', '%s', '%s', %s, %s)`, itemName, itemVariant, hsnCode, uomRaw, uomSmall, uomBig, rawPerSmall, smallPerBig)
	log.Println(itemInsertQuery)

	_, err := db.Query(itemInsertQuery)

	var result map[string]bool

	if err != nil {
		log.Println(err)
		result = map[string]bool{
			"success": false,
		}
	} else {
		result = map[string]bool{
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

// CreateClient creates a new client and returns the status
func CreateClient(w http.ResponseWriter, r *http.Request) {

	clientName := r.FormValue("clientName")

	clientInsertQuery := fmt.Sprintf(`INSERT INTO client
		(clientName)
		VALUES
		('%s')`, clientName)

	_, err := db.Query(clientInsertQuery)

	var result map[string]bool

	if err != nil {
		result = map[string]bool{
			"success": false,
		}
	} else {
		result = map[string]bool{
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

// CreateCustomer creates a new customer and returns the status
func CreateCustomer(w http.ResponseWriter, r *http.Request) {

	customerName := r.FormValue("customerName")

	customerInsertQuery := fmt.Sprintf(`INSERT INTO customer
		(customerName)
		VALUES
		('%s')`, customerName)

	_, err := db.Query(customerInsertQuery)

	var result map[string]bool

	if err != nil {
		result = map[string]bool{
			"success": false,
		}
	} else {
		result = map[string]bool{
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

// InventoryContentQualityCheck ensures sanity of the numbers and ensures the calculation is correct
func InventoryContentQualityCheck(direction string, currentInv string, changeInv string, finalInv string) bool {
	currentInvNum, _ := strconv.Atoi(currentInv)
	changeInvNum, _ := strconv.Atoi(changeInv)
	finalInvNum, _ := strconv.Atoi(finalInv)

	if (currentInvNum + changeInvNum) != finalInvNum {
		return false
	}
	if (currentInvNum < finalInvNum) && (direction == "out") {
		return false
	}
	if (currentInvNum > finalInvNum) && (direction == "in") {
		return false
	}

	log.Println("InventoryContentQualityCheck succeeded")
	return true
}

// InventoryQuantityQualityCheck ensures the total quantity calculation is correct
func InventoryQuantityQualityCheck(quantity string, rate1 string, rate2 string, totalPcs string) bool {
	quantityNum, _ := strconv.Atoi(quantity)
	rate1Num, _ := strconv.Atoi(rate1)
	rate2Num, _ := strconv.Atoi(rate2)
	totalPcsNum, _ := strconv.Atoi(totalPcs)

	if totalPcsNum == 0 {
		return false
	}
	if quantityNum*rate1Num*rate2Num != totalPcsNum {
		return false
	}

	log.Println("InventoryQuantityQualityCheck succeeded")
	return true
}

// InventoryValueQualityCheck ensures the transaction value calculations are correct
func InventoryValueQualityCheck(assdValue string, dutyValue string, gstValue string, totalValue string) bool {
	assdValueNum, _ := strconv.ParseFloat(assdValue, 64)
	dutyValueNum, _ := strconv.ParseFloat(dutyValue, 64)
	gstValueNum, _ := strconv.ParseFloat(gstValue, 64)
	totalValueNum, _ := strconv.ParseFloat(totalValue, 64)

	if (assdValueNum + dutyValueNum + gstValueNum) != totalValueNum {
		return false
	}

	log.Println("InventoryValueQualityCheck succeeded")
	return true
}

// DataSanityDriver is a driver function to trigger checks for inventoryContent, inventoryQuantity, inventoryValue
func DataSanityDriver(direction string, currentInv string, changeInv string, finalInv string, quantity string, rate1 string, rate2 string, totalPcs string, assdValue string, dutyValue string, gstValue string, totalValue string) bool {
	return InventoryContentQualityCheck(direction, currentInv, changeInv, finalInv) && InventoryQuantityQualityCheck(quantity, rate1, rate2, totalPcs) && InventoryValueQualityCheck(assdValue, dutyValue, gstValue, totalValue)
}

func checkCount(rows *sql.Row) (count int) {
	rows.Scan(&count)

	log.Println(count)
	return count
}

// CommitInventoryChanges commits the inventory changes to the inventory table
func CommitInventoryChanges(itemId string, warehouseId string, clientId string, direction string, currentValue string, changeValue string, finalValue string, bigQuantity string, secretRate1 string, secretRate2 string, totalPcs string) bool {
	bigQuantityNum, _ := strconv.Atoi(bigQuantity)
	secretRate1Num, _ := strconv.Atoi(secretRate1)
	secretRate2Num, _ := strconv.Atoi(secretRate2)

	if direction == "out" {
		bigQuantityNum = -bigQuantityNum
	}

	smallboxQuantityNum := bigQuantityNum * secretRate1Num
	itemQuantityNum := smallboxQuantityNum * secretRate2Num

	var executionQuery string

	updateQuery := fmt.Sprintf(`UPDATE inventoryContents
		SET bigcartonQuantity = bigcartonQuantity + %d, smallboxQuantity = smallboxQuantity + %d, itemQuantity = itemQuantity + %d
		WHERE itemId = '%s' AND warehouseId = '%s' AND clientId = '%s' AND bigcartonQuantity = '%s'`, bigQuantityNum, smallboxQuantityNum, itemQuantityNum, itemId, warehouseId, clientId, currentValue)

	insertQuery := fmt.Sprintf(`INSERT INTO inventoryContents
		(itemId, itemQuantity, smallboxQuantity, bigcartonQuantity, warehouseId, clientId)
		VALUES
		('%s', '%d', '%d', '%s', '%s', '%s')`, itemId, itemQuantityNum, smallboxQuantityNum, bigQuantity, warehouseId, clientId)

	checkerQuery := fmt.Sprintf(`SELECT COUNT(*) FROM inventoryContents WHERE itemId = '%s' AND warehouseId = '%s' AND clientId = '%s'`, itemId, warehouseId, clientId)
	countRow := db.QueryRow(checkerQuery)
	if checkCount(countRow) < 1 {
		executionQuery = insertQuery
	} else {
		executionQuery = updateQuery
	}

	_, err := db.Query(executionQuery)
	if err != nil {
		log.Println(err)
		return false
	}

	return true
}

// CreateTransaction creates a transaction
func CreateTransaction(w http.ResponseWriter, r *http.Request) {

	trackingNumber := r.FormValue("trackingNumber")
	entryDate := r.FormValue("entryDate")
	itemId := r.FormValue("itemId")
	warehouseId := r.FormValue("warehouseId")
	comeOrGo := r.FormValue("comeOrGo")
	clientId := r.FormValue("clientId")
	customerId := r.FormValue("customerId")
	bigQuantity := r.FormValue("bigQuantity")
	currentValue := r.FormValue("currentValue")
	changeValue := r.FormValue("changeValue")
	finalValue := r.FormValue("finalValue")
	secretRate1 := r.FormValue("secretRate1")
	secretRate2 := r.FormValue("secretRate2")
	totalPcs := r.FormValue("totalPcs")
	assdValue := r.FormValue("assdValue")
	dutyValue := r.FormValue("dutyValue")
	gstValue := r.FormValue("gstValue")
	totalValue := r.FormValue("totalValue")
	valuePerPiece := r.FormValue("valuePerPiece")
	totalPieces := r.FormValue("totalPieces")
	isPaid := r.FormValue("isPaid")
	paidAmount := r.FormValue("paidAmount")
	date := r.FormValue("date")

	changeValue = strings.TrimSpace(changeValue)
	if date == "Expected Date" {
		date = "NULL"
	}

	var result map[string]bool

	qualityStatus := DataSanityDriver(comeOrGo, currentValue, changeValue, finalValue, bigQuantity, secretRate1, secretRate2, totalPcs, assdValue, dutyValue, gstValue, totalValue)
	if !qualityStatus {
		result = map[string]bool{
			"success": false,
		}

		payloadJSON, err := json.Marshal(result)
		if err != nil {
			log.Println(err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(payloadJSON)

		return
	}

	transactionQuery := fmt.Sprintf(`INSERT INTO transaction
	(trackingNumber, entryDate, itemId, warehouseId, comeOrGo, clientId, customerId, bigQuantity, currentValue, changeValue, finalValue, secretRate1, secretRate2, totalPcs, assdValue, dutyValue, gstValue, totalValue, valuePerPiece, totalPieces, isPaid, paidAmount, date)
	VALUES
	('%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', %s, '%s', '%s')`, trackingNumber, entryDate, itemId, warehouseId, comeOrGo, clientId, customerId, bigQuantity, currentValue, changeValue, finalValue, secretRate1, secretRate2, totalPcs, assdValue, dutyValue, gstValue, totalValue, valuePerPiece, totalPieces, isPaid, paidAmount, date)

	_, err := db.Query(transactionQuery)

	if err != nil {
		log.Println(err)
		result = map[string]bool{
			"success": false,
		}
	} else {
		result = map[string]bool{
			"success": true,
		}
	}

	if err == nil {
		commitStatus := CommitInventoryChanges(itemId, warehouseId, clientId, comeOrGo, currentValue, changeValue, finalValue, bigQuantity, secretRate1, secretRate2, totalPcs)
		if !commitStatus {
			result = map[string]bool{
				"success": false,
			}
		}
	}

	payloadJSON, err := json.Marshal(result)
	if err != nil {
		log.Println(err)
	}

	log.Println(payloadJSON)

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
		itm.itemName, itm.itemVariant, itm.hsnCode, inv.itemQuantity, itm.uomRaw, inv.smallboxQuantity, itm.uomSmall, inv.bigcartonQuantity, itm.uomBig, wh.warehouseName, wh.warehouseLocation, cl.clientName
		FROM inventoryContents inv, itemMaster itm, warehouse wh, client cl
		WHERE inv.itemId IN (%s) AND
		inv.itemId = itm.itemId AND
		inv.warehouseId = wh.warehouseId AND
		inv.clientId = cl.id AND
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
		var clientName string

		err := allContents.Scan(&itemName, &itemVariant, &hsnCode, &itemQuantity, &uomRaw, &smallboxQuantity, &uomSmall, &bigcartonQuantity, &uomBig, &warehouseName, &warehouseLocation, &clientName)
		if err != nil {
			panic(err.Error())
		}

		singleObject := ItemInventory{
			ItemName:          itemName,
			ItemVariant:       itemVariant,
			HsnCode:           hsnCode,
			ItemQuantity:      itemQuantity,
			UomRaw:            uomRaw,
			SmallboxQuantity:  smallboxQuantity,
			UomSmall:          uomSmall,
			BigcartonQuantity: bigcartonQuantity,
			UomBig:            uomBig,
			WarehouseName:     warehouseName,
			WarehouseLocation: warehouseLocation,
			ClientName:        clientName,
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

// SearchSales searches for the sales transactions by filters
func SearchSales(w http.ResponseWriter, r *http.Request) {

	salesInvoiceNumber := r.FormValue("salesInvoiceNumber")
	clientId := r.FormValue("clientId")
	customerId := r.FormValue("customerId")

	var payload []SalesTransaction
	var searchQuery string

	searchQuerySubstring := fmt.Sprintf(`SELECT
		tr.id, tr.trackingNumber, tr.entryDate, tr.itemId, im.itemName, im.itemVariant, tr.warehouseId, wh.warehouseName, wh.warehouseLocation, tr.clientId, cl.clientName, tr.customerId, cu.customerName, tr.changeValue, tr.finalValue, tr.totalPcs, tr.dutyValue, tr.gstValue, tr.totalValue, tr.isPaid, tr.paidAmount, tr.date
		FROM transaction tr, itemMaster im, warehouse wh, client cl, customer cu
		WHERE tr.itemId = im.itemId AND
		tr.warehouseId = wh.warehouseId AND
		tr.clientId = cl.id AND
		tr.customerId = cu.id`)
	trackingNumberSubstring := fmt.Sprintf("tr.trackingNumber = '%s'", salesInvoiceNumber)
	clientIdSubstring := fmt.Sprintf("tr.clientId = '%s'", clientId)
	customerIdSubstring := fmt.Sprintf("tr.customerId = '%s'", customerId)

	if salesInvoiceNumber == "all" && clientId == "all" && customerId == "all" {
		searchQuery = fmt.Sprintf("%s", searchQuerySubstring)
	} else if salesInvoiceNumber == "all" && customerId == "all" {
		searchQuery = fmt.Sprintf("%s AND %s", searchQuerySubstring, clientIdSubstring)
	} else if clientId == "all" && customerId == "all" {
		searchQuery = fmt.Sprintf("%s AND %s", searchQuerySubstring, trackingNumberSubstring)
	} else if salesInvoiceNumber == "all" && clientId == "all" {
		searchQuery = fmt.Sprintf("%s AND %s", searchQuerySubstring, customerIdSubstring)
	} else if salesInvoiceNumber == "all" {
		searchQuery = fmt.Sprintf("%s AND %s AND %s", searchQuerySubstring, clientIdSubstring, customerIdSubstring)
	} else if clientId == "all" {
		searchQuery = fmt.Sprintf("%s AND %s AND %s", searchQuerySubstring, trackingNumberSubstring, customerIdSubstring)
	} else if customerId == "all" {
		searchQuery = fmt.Sprintf("%s AND %s AND %s", searchQuerySubstring, trackingNumberSubstring, clientIdSubstring)
	} else {
		searchQuery = fmt.Sprintf("%s AND %s AND %s AND %s", searchQuerySubstring, trackingNumberSubstring, clientIdSubstring, customerIdSubstring)
	}

	allTransactions, err := db.Query(searchQuery)
	if err != nil {
		panic(err.Error())
	}

	for allTransactions.Next() {
		var transactionId string
		var trackingNumber string
		var entryDate string
		var itemId string
		var itemName string
		var itemVariant string
		var warehouseId string
		var warehouseName string
		var warehouseLocation string
		var clientId string
		var clientName string
		var customerId string
		var customerName string
		var changeValue string
		var finalValue string
		var totalPcs string
		var materialValue string
		var gstValue string
		var totalValue string
		var valuePerPiece float64
		var isPaid string
		var paidAmount string
		var paymentDate string

		err := allTransactions.Scan(&transactionId, &trackingNumber, &entryDate, &itemId, &itemName, &itemVariant, &warehouseId, &warehouseName, &warehouseLocation, &clientId, &clientName, &customerId, &customerName, &changeValue, &finalValue, &totalPcs, &materialValue, &gstValue, &totalValue, &isPaid, &paidAmount, &paymentDate)
		if err != nil {
			panic(err.Error())
		}

		totalValueFloat, _ := strconv.ParseFloat(totalValue, 64)
		totalPcsFloat, _ := strconv.ParseFloat(totalPcs, 64)
		valuePerPiece = totalValueFloat / totalPcsFloat

		singleObject := SalesTransaction{
			TransactionId:     transactionId,
			TrackingNumber:    trackingNumber,
			EntryDate:         entryDate,
			ItemId:            itemId,
			ItemName:          itemName,
			ItemVariant:       itemVariant,
			WarehouseId:       warehouseId,
			WarehouseName:     warehouseName,
			WarehouseLocation: warehouseLocation,
			ClientId:          clientId,
			ClientName:        clientName,
			CustomerId:        customerId,
			CustomerName:      customerName,
			ChangeStock:       changeValue,
			FinalStock:        finalValue,
			TotalPcs:          totalPcs,
			MaterialValue:     materialValue,
			GstValue:          gstValue,
			TotalValue:        totalValue,
			ValuePerPiece:     valuePerPiece,
			IsPaid:            isPaid,
			PaidAmount:        paidAmount,
			PaymentDate:       paymentDate,
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

// UpdatePaidAmount updates the expected payment date for a particular transaction and returns the status
func UpdatePaidAmount(w http.ResponseWriter, r *http.Request) {

	transactionId := r.FormValue("transactionId")
	paidAmount := r.FormValue("paidAmount")

	updateQuery := fmt.Sprintf(`UPDATE transaction
		SET paidAmount = '%s',
		isPaid = CASE WHEN totalValue = paidAmount THEN true ELSE false END
		WHERE id = '%s'`, paidAmount, transactionId)

	_, err := db.Query(updateQuery)

	var result map[string]bool

	if err != nil {
		result = map[string]bool{
			"success": false,
		}
	} else {
		result = map[string]bool{
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

// UpdatePaymentDate updates the expected payment date for a particular transaction and returns the status
func UpdatePaymentDate(w http.ResponseWriter, r *http.Request) {

	transactionId := r.FormValue("transactionId")
	paymentDate := r.FormValue("paymentDate")

	updateQuery := fmt.Sprintf(`UPDATE transaction
		SET date = '%s'
		WHERE id = '%s'`, paymentDate, transactionId)

	_, err := db.Query(updateQuery)

	var result map[string]bool

	if err != nil {
		result = map[string]bool{
			"success": false,
		}
	} else {
		result = map[string]bool{
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

// RegisterUser creates a new user and returns the status
func RegisterUser(w http.ResponseWriter, r *http.Request) {

	username := r.FormValue("username")
	passwordPlainText := r.FormValue("password")

	password := GetMD5Hash(passwordPlainText)

	userInsertQuery := fmt.Sprintf(`INSERT INTO user (username, password)
		SELECT '%s', '%s'
		WHERE NOT EXISTS (SELECT username FROM user WHERE username ='%s') LIMIT 1`, username, password, username)

	_, err := db.Query(userInsertQuery)

	var result map[string]bool

	if err != nil {
		result = map[string]bool{
			"success": false,
		}
	} else {
		result = map[string]bool{
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

// LoginUser creates a new user and returns the status
func LoginUser(w http.ResponseWriter, r *http.Request) {

	username := r.FormValue("username")
	passwordPlainText := r.FormValue("password")

	password := GetMD5Hash(passwordPlainText)

	userLoginQuery := fmt.Sprintf(`SELECT id, permission_createNew, permission_transactionIn, permission_transactionOut, permission_view FROM user WHERE username='%s' AND password='%s'`, username, password)
	rows, err := db.Query(userLoginQuery)

	var result map[string]bool

	if err != nil {
		result = map[string]bool{
			"success": false,
		}
	} else {
		var userId int
		var success bool
		var permission_createNew bool
		var permission_transactionIn bool
		var permission_transactionOut bool
		var permission_view bool

		for rows.Next() {
			success = true
			rows.Scan(&userId, &permission_createNew, &permission_transactionIn, &permission_transactionOut, &permission_view)
		}

		if success {
			result = map[string]bool{
				"success":                   success,
				"permission_createNew":      permission_createNew,
				"permission_transactionIn":  permission_transactionIn,
				"permission_transactionOut": permission_transactionOut,
				"permission_view":           permission_view,
			}
		} else {
			result = map[string]bool{
				"success": false,
			}
		}
	}

	payloadJSON, err := json.Marshal(result)
	if err != nil {
		log.Println(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(payloadJSON)
}
