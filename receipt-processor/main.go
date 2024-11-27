package main

import (
	"encoding/json"
	"io"
	"log"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

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

// RULES
// * [done] One point for every alphanumeric character in the retailer name.
// * [done] 50 points if the total is a round dollar amount with no cents.
// * [done] 25 points if the total is a multiple of `0.25`.
// * [done] 5 points for every two items on the receipt.
// * If the trimmed length of the item description is a multiple of 3, multiply the price by `0.2` and round up to the nearest integer. The result is the number of points earned.
// * [done] 6 points if the day in the purchase date is odd.
// * [done] 10 points if the time of purchase is after 2:00pm and before 4:00pm.

func calculate(receipt Receipt) int64 {
	points := int64(0)

	// One point for every alphanumeric character in the retailer name.
	var nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9]+`)
	retailer := nonAlphanumericRegex.ReplaceAllString(receipt.Retailer, "")
	points += int64(len(retailer))

	// fmt.Println(points)

	total, err := strconv.ParseFloat(receipt.Total, 64)
	if err != nil {
		log.Println("unable to parse float the receipt's total")
		log.Println("error:", err)
		return -1
	}

	// 50 points if the total is a round dollar amount with no cents.
	if total == float64(int(total)) {
		points += 50
	}
	// fmt.Println(points)

	// 25 points if the total is a multiple of `0.25`
	if int(total*100)%25 == 0 {
		points += 25
	}
	// fmt.Println(points)

	// 5 points for every two items on the receipt.
	points += 5 * int64((len(receipt.Items))/2)
	// fmt.Println(points)

	// If the trimmed length of the item description is a multiple of 3, multiply the price by `0.2` and round up to the nearest integer. The result is the number of points earned.
	for _, item := range receipt.Items {
		description := strings.TrimSpace(item.ShortDescription)

		if int64(len(description))%3 == 0 {
			price, err := strconv.ParseFloat(item.Price, 64)
			if err != nil {
				log.Println("unable to parse float the item's price")
				log.Println("error:", err)
				return -1
			}
			points += int64(math.Ceil(price * 0.2))
			// fmt.Println(points)
		}
	}

	day, err := time.Parse("2006-01-02", receipt.PurchaseDate)
	if err != nil {
		log.Println("unable to parse date")
		log.Println("error:", err)
		return -1
	}
	// 6 points if the day in the purchase date is odd.
	if day.Day()%2 == 1 {
		points += 6
	}
	// fmt.Println(points)

	t, err := time.Parse("15:04", receipt.PurchaseTime)
	if err != nil {
		log.Println("unable to parse time")
		log.Println("error:", err)
		return -1
	}
	// 10 points if the time of purchase is after 2:00pm and before 4:00pm.
	atTwo := time.Date(0, 1, 1, 14, 0, 0, 0, time.UTC)
	atFour := time.Date(0, 1, 1, 16, 0, 0, 0, time.UTC)

	if t.After(atTwo) && t.Before(atFour) {
		points += 10
	}
	// fmt.Println(points)

	return points
}

func getPoints(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	id, exist := params["id"]
	if !exist {
		log.Println("id does not exist")
		log.Println("error: ", exist)
		return
	}

	uuid, err := uuid.Parse(id)
	if err != nil {
		log.Println("unable to parse uuid")
		log.Println("error: ", err)
		return
	}

	receipt, exist := receiptCached[uuid.String()]
	if !exist {
		log.Println("receipt does not exist")
		log.Println("error: ", exist)
		return
	}

	points := calculate(receipt)

	response := GetReceiptPoints{
		Points: points,
	}

	ret, err := json.Marshal(response)
	if err != nil {
		log.Println("unable to marshal JSON")
		log.Println("error: ", err)
		return
	}

	w.Write(ret)
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/receipts/process", processReceipt).Methods("POST")
	r.HandleFunc("/receipts/{id}/points", getPoints).Methods("GET")

	http.Handle("/", r)
	http.ListenAndServe(":8080", nil)
}
