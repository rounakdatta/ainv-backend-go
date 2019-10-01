package main

import (
	"log"
	"net/http"

	"github.com/bitly/go-simplejson"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/", GetRoot).Methods("GET")
	router.HandleFunc("/api/get/warehouses", GetWarehouses).Methods("GET")
	// router.HandleFunc("/api/get/items", GetItems).Methods("GET")
	// router.HandleFunc("/api/search/items", SearchItems).Methods("POST")
	// router.HandleFunc("/api/put/warehouse", CreateWarehouse).Methods("POST")
	// router.HandleFunc("/api/put/itemmaster", CreateItemMaster).Methods("POST")

	http.Handle("/", router)

	log.Println("Server started on port 5000")
	log.Fatal(http.ListenAndServe(":5000", nil))
}

// GetRoot returns OK if server is alive
func GetRoot(w http.ResponseWriter, r *http.Request) {
	payload := []byte("OK")
	w.Write(payload)
}

// GetWarehouses returns all the warehouses with their ID
func GetWarehouses(w http.ResponseWriter, r *http.Request) {

	payloadJSON := simplejson.New()
	payloadJSON.Set("foo", "bar")

	payload, err := payloadJSON.MarshalJSON()
	if err != nil {
		log.Println(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(payload)
}
