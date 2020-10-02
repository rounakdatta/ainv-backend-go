package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math"
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
	SmallUnit      string `json:"smallUnit"`
	MediumUnit     string `json:"mediumUnit"`
	BigUnit        string `json:"bigUnit"`
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

type BillOfEntry struct {
	BillOfEntryNumber string `json:"billOfEntryNumber"`
	BillOfEntryId     string `json:"billOfEntryId"`
	BillOfEntryDate   string `json:"billOfEntryDate"`
}

type SalesInvoice struct {
	SalesInvoiceNumber string `json:"salesInvoiceNumber"`
	SalesInvoiceId     string `json:"salesInvoiceId"`
	SalesInvoiceDate   string `json:"salesInvoiceDate"`
	CustomerId         string `json:"customerId"`
	CustomerName       string `json:"customerName"`
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
	BillOfEntry       string  `json:"billOfEntry"`
	SalesInvoice      string  `json:"salesInvoice"`
	EntryDate         string  `json:"entryDate"`
	ItemId            string  `json:"itemId"`
	ItemName          string  `json:"itemName"`
	ItemVariant       string  `json:"itemVariant"`
	WarehouseName     string  `json:"warehouseName"`
	WarehouseLocation string  `json:"warehouseLocation"`
	ClientId          string  `json:"clientId"`
	ClientName        string  `json:"clientName"`
	CustomerId        string  `json:"customerId"`
	CustomerName      string  `json:"customerName"`
	ComeOrGo          string  `json:"comeOrGo"`
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
	Field1            string  `json:"field1"`
	Field2            string  `json:"field2"`
	Remarks           string  `json:"remarks"`
	RawUnit           string  `json:"rawUnit"`
}

type OverviewTransaction struct {
	BillOfEntryId  string `json:"billOfEntryId"`
	BillOfEntry    string `json:"billOfEntry"`
	SalesInvoiceId string `json:"salesInvoiceId"`
	SalesInvoice   string `json:"salesInvoice"`
	Direction      string `json:"direction"`
	EntryDate      string `json:"entryDate"`
	Item           string `json:"item"`
	Warehouse      string `json:"warehouse"`
	Client         string `json:"client"`
	Customer       string `json:"customer"`
	BigQuantity    string `json:"bigQuantity"`
	TotalValue     string `json:"totalValue"`
	IsPaid         string `json:"isPaid"`
	PaidAmount     string `json:"paidAmount"`
	Date           string `json:"date"`
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

	// obtain the cli arguments
	serviceName := os.Args[1]
	servicePort := os.Args[2]

	// create the router and define the APIs
	router := mux.NewRouter()
	ainvRouter := router.PathPrefix("/" + serviceName).Subrouter()

	ainvRouter.HandleFunc("/", GetRoot).Methods("GET")

	ainvRouter.HandleFunc("/api/get/warehouses/", GetWarehouses).Methods("GET")
	ainvRouter.HandleFunc("/api/get/all/warehouses/", GetAllWarehouses).Methods("GET")
	ainvRouter.HandleFunc("/api/get/all/clients/", GetAllClients).Methods("GET")
	ainvRouter.HandleFunc("/api/get/all/customers/", GetAllCustomers).Methods("GET")
	ainvRouter.HandleFunc("/api/get/items/", GetItems).Methods("GET")
	ainvRouter.HandleFunc("/api/get/all/bills/", GetAllBills).Methods("GET")
	ainvRouter.HandleFunc("/api/get/all/invoices/", GetAllInvoices).Methods("GET")
	ainvRouter.HandleFunc("/api/get/rate/", GetRate).Methods("POST")

	ainvRouter.HandleFunc("/api/put/warehouse/", CreateWarehouse).Methods("POST")
	ainvRouter.HandleFunc("/api/put/itemmaster/", CreateItemMaster).Methods("POST")
	ainvRouter.HandleFunc("/api/put/transaction/", CreateTransaction).Methods("POST")
	ainvRouter.HandleFunc("/api/put/client/", CreateClient).Methods("POST")
	ainvRouter.HandleFunc("/api/put/customer/", CreateCustomer).Methods("POST")

	ainvRouter.HandleFunc("/api/update/paidamount/", UpdatePaidAmount).Methods("POST")
	ainvRouter.HandleFunc("/api/update/paymentdate/", UpdatePaymentDate).Methods("POST")
	ainvRouter.HandleFunc("/api/update/field1/", UpdateField1).Methods("POST")
	ainvRouter.HandleFunc("/api/update/field2/", UpdateField2).Methods("POST")
	ainvRouter.HandleFunc("/api/update/remarks/", UpdateRemarks).Methods("POST")

	ainvRouter.HandleFunc("/api/search/items/", SearchItems).Methods("POST")
	ainvRouter.HandleFunc("/api/search/sales/", SearchSales).Methods("POST")
	ainvRouter.HandleFunc("/api/search/overview/", SearchOverview).Methods("POST")

	ainvRouter.HandleFunc("/api/register/", RegisterUser).Methods("POST")
	ainvRouter.HandleFunc("/api/login/", LoginUser).Methods("POST")

	http.Handle("/", router)

	log.Printf("Server started on port %s", servicePort)
	log.Fatal(http.ListenAndServe(":"+servicePort, nil))
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
		GROUP_CONCAT(id SEPARATOR '$') warehouseId
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
		id as warehouseId, CONCAT(warehouseName, ", ", warehouseLocation) AS warehouseName
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

// GetAllBills returns all the Bill of Entry numbers with their IDs
func GetAllBills(w http.ResponseWriter, r *http.Request) {

	var payload []BillOfEntry

	getBillsQuery := `SELECT 
		id, tracker, entryDate
		FROM billOfEntry`

	allBills, err := db.Query(getBillsQuery)
	if err != nil {
		panic(err.Error())
	}

	for allBills.Next() {
		var billId string
		var billNumber string
		var billDate string

		err := allBills.Scan(&billId, &billNumber, &billDate)
		if err != nil {
			panic(err.Error())
		}

		singleObject := BillOfEntry{
			BillOfEntryId:     billId,
			BillOfEntryNumber: billNumber,
			BillOfEntryDate:   billDate,
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

// GetAllInvoices returns all the Sales Invoice numbers with their IDs
func GetAllInvoices(w http.ResponseWriter, r *http.Request) {

	var payload []SalesInvoice

	getInvoicesQuery := `SELECT 
		id, tracker, entryDate, customerId, (select customerName from customer where id=customerId) as customerId
		FROM salesInvoice`

	allInvoices, err := db.Query(getInvoicesQuery)
	if err != nil {
		panic(err.Error())
	}

	for allInvoices.Next() {
		var invId string
		var invNumber string
		var invDate string
		var customerId string
		var customerName string

		err := allInvoices.Scan(&invId, &invNumber, &invDate, &customerId, &customerName)
		if err != nil {
			panic(err.Error())
		}

		singleObject := SalesInvoice{
			SalesInvoiceId:     invId,
			SalesInvoiceNumber: invNumber,
			SalesInvoiceDate:   invDate,
			CustomerId:         customerId,
			CustomerName:       customerName,
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

	getRatesQuery := fmt.Sprintf(`SELECT im.rawPerSmall, im.smallPerBig, IFNULL(ic.bigcartonQuantity, 0) AS cartonQuantity, im.uomRaw AS smallUnit, im.uomSmall AS mediumUnit, im.uomBig AS bigUnit
		FROM itemMaster im
		LEFT JOIN inventoryContents ic
		ON (im.id = ic.itemId AND ic.warehouseId = '%s' AND ic.clientId = '%s')
		WHERE im.id = '%s'`, requestedWarehouseId, requestedClientId, requestedItemId)

	log.Println(getRatesQuery)
	rateDetails, err := db.Query(getRatesQuery)
	if err != nil {
		panic(err.Error())
	}

	for rateDetails.Next() {
		var rawPerSmall string
		var smallPerBig string
		var cartonQuantity string
		var smallUnit string
		var mediumUnit string
		var bigUnit string

		err := rateDetails.Scan(&rawPerSmall, &smallPerBig, &cartonQuantity, &smallUnit, &mediumUnit, &bigUnit)
		if err != nil {
			panic(err.Error())
		}

		singleObject := Rate{
			RawPerSmall:    rawPerSmall,
			SmallPerBig:    smallPerBig,
			CartonQuantity: cartonQuantity,
			SmallUnit:      smallUnit,
			MediumUnit:     mediumUnit,
			BigUnit:        bigUnit,
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
		GROUP_CONCAT(id SEPARATOR '$') itemId
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
	totalPcsNum, _ := strconv.ParseFloat(totalPcs, 64)

	if totalPcsNum <= 0 {
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

	calculatedValueNum := assdValueNum + dutyValueNum + gstValueNum
	calculatedValueNum = math.Floor(calculatedValueNum*100) / 100

	assdValueNum = math.Floor(assdValueNum*100) / 100
	dutyValueNum = math.Floor(dutyValueNum*100) / 100
	gstValueNum = math.Floor(gstValueNum*100) / 100
	totalValueNum = math.Floor(totalValueNum*100) / 100

	if calculatedValueNum != totalValueNum {
		return false
	}

	log.Println("InventoryValueQualityCheck succeeded")
	return true
}

// DataSanityDriver is a driver function to trigger checks for inventoryContent, inventoryQuantity, inventoryValue
func DataSanityDriver(direction string, currentInv string, changeInv string, finalInv string, quantity string, rate1 string, rate2 string, totalPcs string, assdValue string, dutyValue string, gstValue string, totalValue string) bool {
	return InventoryContentQualityCheck(direction, currentInv, changeInv, finalInv) && InventoryQuantityQualityCheck(quantity, rate1, rate2, totalPcs)
}

func checkCount(rows *sql.Row) (count int) {
	rows.Scan(&count)

	log.Println(count)
	return count
}

// CommitInventoryChanges commits the inventory changes to the inventory table
func CommitInventoryChanges(itemId string, warehouseId string, clientId string, direction string, currentValue string, changeValue string, finalValue string, bigQuantity string, secretRate1 string, secretRate2 string, totalPcs string) bool {
	bigQuantityNum, _ := strconv.ParseFloat(bigQuantity, 64)
	secretRate1Num, _ := strconv.ParseFloat(secretRate1, 64)
	secretRate2Num, _ := strconv.ParseFloat(secretRate2, 64)

	if direction == "out" {
		bigQuantityNum = -bigQuantityNum
	}

	smallboxQuantityNum := bigQuantityNum * secretRate1Num
	itemQuantityNum := smallboxQuantityNum * secretRate2Num

	var executionQuery string

	updateQuery := fmt.Sprintf(`UPDATE inventoryContents
		SET bigcartonQuantity = bigcartonQuantity + %f, smallboxQuantity = smallboxQuantity + %f, itemQuantity = itemQuantity + %f
		WHERE itemId = '%s' AND warehouseId = '%s' AND clientId = '%s' AND bigcartonQuantity = '%s'`, bigQuantityNum, smallboxQuantityNum, itemQuantityNum, itemId, warehouseId, clientId, currentValue)

	insertQuery := fmt.Sprintf(`INSERT INTO inventoryContents
		(itemId, itemQuantity, smallboxQuantity, bigcartonQuantity, warehouseId, clientId)
		VALUES
		('%s', '%f', '%f', '%s', '%s', '%s')`, itemId, itemQuantityNum, smallboxQuantityNum, bigQuantity, warehouseId, clientId)

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

	oldOrNew := r.FormValue("oldOrNew")
	billRef := r.FormValue("billRef")
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
	secretRate1 := r.FormValue("secretRate2")
	secretRate2 := r.FormValue("secretRate1")
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
	field1 := r.FormValue("field1")
	field2 := r.FormValue("field2")
	remarks := r.FormValue("remarks")

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

	if comeOrGo == "in" {
		billRef = trackingNumber
		trackingNumber = "NULL"

		if oldOrNew == "New!" {

			beEntryQuery := fmt.Sprintf(`
				INSERT INTO billOfEntry (tracker, entryDate, customerId) VALUES ('%s', '%s', '%s')
			`, billRef, entryDate, clientId)
			_, err := db.Query(beEntryQuery)

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

			beIdSelectQuery := fmt.Sprintf(`
				SELECT id FROM billOfEntry WHERE tracker='%s'
			`, billRef)

			beData := db.QueryRow(beIdSelectQuery)
			beData.Scan(&billRef)

		} else {

			beIdSelectQuery := fmt.Sprintf(`
				SELECT id FROM billOfEntry WHERE tracker='%s'
			`, billRef)

			beData := db.QueryRow(beIdSelectQuery)
			beData.Scan(&billRef)

			billRef = fmt.Sprintf("'%s'", billRef)
		}

	} else {
		if oldOrNew == "New!" {
			siEntryQuery := fmt.Sprintf(`
				INSERT INTO salesInvoice (tracker, entryDate, customerId) VALUES ('%s', '%s', '%s')
			`, trackingNumber, entryDate, customerId)
			_, err := db.Query(siEntryQuery)

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

			siSelectQuery := fmt.Sprintf(`
				SELECT id FROM salesInvoice WHERE tracker='%s'
			`, trackingNumber)

			siData := db.QueryRow(siSelectQuery)
			siData.Scan(&trackingNumber)

			trackingNumber = fmt.Sprintf("'%s'", trackingNumber)
		} else {
			trackingNumber = oldOrNew
		}
	}

	transactionQuery := fmt.Sprintf(`INSERT INTO transaction
	(billOfEntry, salesInvoice, itemId, warehouseId, comeOrGo, clientId, customerId, bigQuantity, currentValue, changeValue, finalValue, secretRate1, secretRate2, totalPcs, assdValue, dutyValue, gstValue, totalValue, valuePerPiece, totalPieces, isPaid, paidAmount, date, delvDate1, delvDate2, remarks)
	VALUES
	(%s, %s, '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', %s, '%s', '%s', '%s', '%s', '%s')`, billRef, trackingNumber, itemId, warehouseId, comeOrGo, clientId, customerId, bigQuantity, currentValue, changeValue, finalValue, secretRate1, secretRate2, totalPcs, assdValue, dutyValue, gstValue, totalValue, valuePerPiece, totalPieces, isPaid, paidAmount, date, field1, field2, remarks)

	fmt.Println(transactionQuery)
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
	requestedClientsRaw := r.FormValue("clients")

	requestedItemId := strings.Split(strings.TrimSpace(requestedItemIdRaw), " ")
	requestedLocations := strings.Split(strings.TrimSpace(requestedLocationsRaw), " ")
	requestedClients := strings.Split(strings.TrimSpace(requestedClientsRaw), " ")

	items := strings.Join(requestedItemId, ", ")
	locations := strings.Join(requestedLocations, ", ")
	clients := strings.Join(requestedClients, ", ")

	var payload []ItemInventory

	searchQuery := fmt.Sprintf(`SELECT 
		itm.itemName, itm.itemVariant, itm.hsnCode, inv.itemQuantity, itm.uomRaw, inv.smallboxQuantity, itm.uomSmall, inv.bigcartonQuantity, itm.uomBig, wh.warehouseName, wh.warehouseLocation, cl.clientName
		FROM inventoryContents inv, itemMaster itm, warehouse wh, client cl
		WHERE inv.itemId IN (%s) AND
		inv.clientId IN (%s) AND
		inv.itemId = itm.id AND
		inv.warehouseId = wh.id AND
		inv.clientId = cl.id AND
		wh.id IN (%s)`, items, clients, locations)

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

	billOfEntry := r.FormValue("billOfEntry")
	clientId := r.FormValue("clientId")
	customerId := r.FormValue("customerId")
	searchFilter := r.FormValue("filter")

	var payload []SalesTransaction
	var searchQuery string

	var filterSubstring string
	var billOrSales string
	if searchFilter == "in" {
		filterSubstring = " AND tr.comeOrGo = 'in'"
		billOrSales = "billOfEntry"
	} else if searchFilter == "out" {
		filterSubstring = " AND tr.comeOrGo = 'out'"
		billOrSales = "salesInvoice"
	} else {
		filterSubstring = ""
		billOrSales = "salesInvoice"
	}

	searchQuerySubstring := fmt.Sprintf(`
		SELECT
		tr.id,
		IFNULL((select tracker from billOfEntry where id=tr.billOfEntry), 'N/A') as billOfEntry,
		IFNULL((select tracker from salesInvoice where id=tr.salesInvoice), 'N/A') as salesInvoice,
	IFNULL((
	SELECT
		entryDate
	FROM
		%s
	WHERE
		id = tr.%s
	), 'N/A') AS entryDate,
	tr.itemId,
	im.itemName,
	im.itemVariant,
	wh.warehouseName,
	wh.warehouseLocation,
	tr.clientId,
	cl.clientName,
	tr.customerId,
	cu.customerName,
	tr.comeOrGo,
	tr.changeValue,
	tr.finalValue,
	tr.totalPcs,
	tr.dutyValue,
	tr.gstValue,
	tr.totalValue,
	tr.isPaid,
	tr.paidAmount,
	tr.date,
	tr.delvDate1,
	tr.delvDate2,
	tr.remarks,
	im.uomRaw
	FROM transaction
		tr,
		itemMaster im,
		warehouse wh,
		client cl,
		customer cu
	WHERE
		tr.itemId = im.id AND tr.warehouseId = wh.id AND tr.clientId = cl.id AND tr.customerId = cu.id
	`, billOrSales, billOrSales)

	searchQuerySubstring = searchQuerySubstring + filterSubstring

	trackingNumberSubstring := fmt.Sprintf("tr.billOfEntry = '%s'", billOfEntry)
	clientIdSubstring := fmt.Sprintf("tr.clientId = '%s'", clientId)
	customerIdSubstring := fmt.Sprintf("tr.customerId = '%s'", customerId)

	if billOfEntry == "all" && clientId == "all" && customerId == "all" {
		searchQuery = fmt.Sprintf("%s", searchQuerySubstring)
	} else if billOfEntry == "all" && customerId == "all" {
		searchQuery = fmt.Sprintf("%s AND %s", searchQuerySubstring, clientIdSubstring)
	} else if clientId == "all" && customerId == "all" {
		searchQuery = fmt.Sprintf("%s AND %s", searchQuerySubstring, trackingNumberSubstring)
	} else if billOfEntry == "all" && clientId == "all" {
		searchQuery = fmt.Sprintf("%s AND %s", searchQuerySubstring, customerIdSubstring)
	} else if billOfEntry == "all" {
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
		var billOfEntry string
		var salesInvoice string
		var entryDate string
		var itemId string
		var itemName string
		var itemVariant string
		var warehouseName string
		var warehouseLocation string
		var clientId string
		var clientName string
		var customerId string
		var customerName string
		var comeOrGo string
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
		var field1 string
		var field2 string
		var remarks string
		var rawUnit string

		err := allTransactions.Scan(&transactionId, &billOfEntry, &salesInvoice, &entryDate, &itemId, &itemName, &itemVariant, &warehouseName, &warehouseLocation, &clientId, &clientName, &customerId, &customerName, &comeOrGo, &changeValue, &finalValue, &totalPcs, &materialValue, &gstValue, &totalValue, &isPaid, &paidAmount, &paymentDate, &field1, &field2, &remarks, &rawUnit)
		if err != nil {
			panic(err.Error())
		}

		totalValueFloat, _ := strconv.ParseFloat(totalValue, 64)
		totalPcsFloat, _ := strconv.ParseFloat(totalPcs, 64)
		valuePerPiece = totalValueFloat / totalPcsFloat

		singleObject := SalesTransaction{
			TransactionId:     transactionId,
			BillOfEntry:       billOfEntry,
			SalesInvoice:      salesInvoice,
			EntryDate:         entryDate,
			ItemId:            itemId,
			ItemName:          itemName,
			ItemVariant:       itemVariant,
			WarehouseName:     warehouseName,
			WarehouseLocation: warehouseLocation,
			ClientId:          clientId,
			ClientName:        clientName,
			CustomerId:        customerId,
			CustomerName:      customerName,
			ComeOrGo:          comeOrGo,
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
			Field1:            field1,
			Field2:            field2,
			Remarks:           remarks,
			RawUnit:           rawUnit,
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

// SearchOverview searches overview of transactions by filters
func SearchOverview(w http.ResponseWriter, r *http.Request) {

	salesInvoiceNumber := r.FormValue("salesInvoiceNumber")
	clientId := r.FormValue("clientId")
	customerId := r.FormValue("customerId")
	searchFilter := r.FormValue("filter")
	itemFilter := r.FormValue("itemName")

	var payload []OverviewTransaction
	var searchQuery string

	var filterSubstring string
	if searchFilter == "in" {
		filterSubstring = " AND temp.direction = 'in'"
	} else if searchFilter == "out" {
		filterSubstring = " AND temp.direction = 'out'"
	} else {
		filterSubstring = ""
	}

	var itemFilterSubstring string
	if itemFilter == "all" || itemFilter == "" {
		itemFilterSubstring = ""
	} else {
		itemFilterSubstring = fmt.Sprintf(" AND temp.item = '%s'", itemFilter)
	}

	searchQuerySubstring := fmt.Sprintf(`SELECT * FROM
	(SELECT 
		GROUP_CONCAT(DISTINCT(IFNULL(billOfEntryId, 'N/A'))) as billOfEntryId, 
		IFNULL(billOfEntry, 'N/A') as billOfEntry, 
		GROUP_CONCAT(DISTINCT(IFNULL(salesInvoiceId, 'N/A'))) as salesInvoiceId, 
		GROUP_CONCAT(DISTINCT(IFNULL(salesInvoice, 'N/A'))) as salesInvoice, 
		direction, 
		GROUP_CONCAT(DISTINCT(entryDate)) as entryDate, 
		GROUP_CONCAT(DISTINCT(item)) as item, 
		GROUP_CONCAT(DISTINCT(warehouse)) as warehouse,
        GROUP_CONCAT(DISTINCT(clientId)) as clientId,
		GROUP_CONCAT(DISTINCT(client)) as client, 
        GROUP_CONCAT(DISTINCT(customerId)) as customerId,
		GROUP_CONCAT(DISTINCT(customer)) as customer, 
		sum(bigQuantity) as bigQuantity, 
		sum(totalValue) as totalValue, 
		GROUP_CONCAT(DISTINCT(isPaid)) as isPaid, 
		sum(paidAmount) as paidAmount, 
		GROUP_CONCAT(DISTINCT(date)) as date 
	from 
		(
		SELECT 
			billOfEntry as billOfEntryId, 
			(
			SELECT 
				tracker 
			FROM 
				billOfEntry 
			WHERE 
				billOfEntry.id = billOfEntry
			) AS billOfEntry, 
			salesInvoice as salesInvoiceId, 
			(
			SELECT 
				tracker 
			FROM 
				salesInvoice 
			WHERE 
				salesInvoice.id = salesInvoice
			) AS salesInvoice, 
			min(comeOrGo) as direction, 
			(
			select 
				case when min(comeOrGo) like 'in' then GROUP_CONCAT(combo.bee) else GROUP_CONCAT(combo.sie) end 
			from 
				(
				select 
					id as be, 
					NULL as si, 
					entryDate as bee, 
					NULL as sie 
				from 
					billOfEntry 
				union all 
				select 
					NULL as be, 
					id as si, 
					NULL as bee, 
					entryDate as sie 
				from 
					salesInvoice
				) combo 
			where 
				billOfEntryId = combo.be 
				or salesInvoice = combo.si
			) as entryDate, 
			Group_concat(
			DISTINCT (
				SELECT 
				itemName 
				FROM 
				itemMaster 
				WHERE 
				itemMaster.id = itemId
			) SEPARATOR ' '
			) AS item, 
			Group_concat(
			DISTINCT (
				SELECT 
				Concat(
					warehouseName, ', ', warehouseLocation
				) 
				FROM 
				warehouse 
				WHERE 
				warehouse.id = warehouseId
			) SEPARATOR ' '
			) AS warehouse, 
			min(clientId) as clientId,
			Group_concat(
			DISTINCT (
				SELECT 
				clientName 
				FROM 
				client 
				WHERE 
				client.id = clientId
			) SEPARATOR ' '
			) AS client, 
			min(customerId) as customerId,
			Group_concat(
			DISTINCT (
				SELECT 
				customerName 
				FROM 
				customer 
				WHERE 
				customer.id = customerId
			) SEPARATOR ' '
			) AS customer, 
			Sum(bigQuantity) AS bigQuantity, 
			Sum(totalValue) AS totalValue, 
			'...' AS isPaid, 
			Sum(paidAmount) AS paidAmount, 
			'...' AS date 
		FROM 
			transaction 
		WHERE
			isError=0
		GROUP BY 
			billOfEntry, 
			salesInvoice
		) agg WHERE 1=1
	`)

	trackingNumberSubstring := fmt.Sprintf("(agg.billOfEntry = '%s' OR agg.salesInvoice = '%s')", salesInvoiceNumber, salesInvoiceNumber)
	clientIdSubstring := fmt.Sprintf("agg.clientId = '%s'", clientId)
	customerIdSubstring := fmt.Sprintf("agg.customerId = '%s'", customerId)

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

	searchQuery = searchQuery + " group by billOfEntry, direction) temp WHERE 1=1"
	searchQuery = searchQuery + filterSubstring
	searchQuery = searchQuery + itemFilterSubstring
	searchQuery = searchQuery + " ORDER BY temp.billOfEntryId DESC"

	allTransactions, err := db.Query(searchQuery)
	if err != nil {
		panic(err.Error())
	}

	for allTransactions.Next() {
		var billOfEntryId string
		var billOfEntry string
		var salesInvoiceId string
		var salesInvoice string
		var direction string
		var entryDate string
		var item string
		var warehouse string
		var clientId string
		var client string
		var customerId string
		var customer string
		var bigQuantity string
		var totalValue string
		var isPaid string
		var paidAmount string
		var date string

		err := allTransactions.Scan(&billOfEntryId, &billOfEntry, &salesInvoiceId, &salesInvoice, &direction, &entryDate, &item, &warehouse, &clientId, &client, &customerId, &customer, &bigQuantity, &totalValue, &isPaid, &paidAmount, &date)
		if err != nil {
			panic(err.Error())
		}

		if len(salesInvoice) > 30 {
			salesInvoice = salesInvoice[:30] + "..."
		}

		if len(customer) > 30 {
			customer = customer[:30] + "..."
		}

		singleObject := OverviewTransaction{
			BillOfEntryId:  billOfEntryId,
			BillOfEntry:    billOfEntry,
			SalesInvoiceId: salesInvoiceId,
			SalesInvoice:   salesInvoice,
			Direction:      direction,
			EntryDate:      entryDate,
			Item:           item,
			Warehouse:      warehouse,
			Client:         client,
			Customer:       customer,
			BigQuantity:    bigQuantity,
			TotalValue:     totalValue,
			IsPaid:         isPaid,
			PaidAmount:     paidAmount,
			Date:           date,
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
		isPaid = CASE WHEN cast(totalValue as unsigned) = cast(paidAmount as unsigned) THEN true ELSE false END
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
	paymentDate := r.FormValue("paymentdate")

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

// UpdateField1 updates the field 1 and returns the status
func UpdateField1(w http.ResponseWriter, r *http.Request) {

	transactionId := r.FormValue("transactionId")
	field := r.FormValue("field1")

	updateQuery := fmt.Sprintf(`UPDATE transaction
		SET delvDate1 = '%s'
		WHERE id = '%s'`, field, transactionId)

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

// UpdateField2 updates the field 2 and returns the status
func UpdateField2(w http.ResponseWriter, r *http.Request) {

	transactionId := r.FormValue("transactionId")
	field := r.FormValue("field2")

	updateQuery := fmt.Sprintf(`UPDATE transaction
		SET delvDate2 = '%s'
		WHERE id = '%s'`, field, transactionId)

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

// UpdateRemarks updates the remarks and returns the status
func UpdateRemarks(w http.ResponseWriter, r *http.Request) {

	transactionId := r.FormValue("transactionId")
	field := r.FormValue("remarks")

	updateQuery := fmt.Sprintf(`UPDATE transaction
		SET remarks = '%s'
		WHERE id = '%s'`, field, transactionId)

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
