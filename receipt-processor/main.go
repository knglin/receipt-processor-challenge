package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

var receiptCached = make(map[string]Receipt)

func processReceipt(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("receipt has not correct format")
		log.Println("error: ", err)
		return
	}

	var receipt Receipt
	err = json.Unmarshal(body, &receipt)
	if err != nil {
		log.Println("unable to unmarshal JSON")
		log.Println("error: ", err)
		return
	}

	receiptId := uuid.New().String()
	receipt.ID = receiptId

	receiptCached[receiptId] = receipt

	response := PostReceiptID{
		ID: receiptId,
	}

	ret, err := json.Marshal(response)
	if err != nil {
		log.Println("unable to marshal JSON")
		log.Println("error: ", err)
		return
	}

	w.Write(ret)
}

// func calcPoints(w http.ResponseWriter, r *http.Request) {

// }

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/receipts/process", processReceipt).Methods("POST")
	// r.HandleFunc("/receipts/{id}/points", calcPoints).Methods("GET")

	http.Handle("/", r)
	http.ListenAndServe(":8080", nil)
}
