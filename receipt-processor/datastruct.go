package main

type Receipt struct {
	ID           string `json:"id"`
	Points       int64  `json:"points"`
	Retailer     string `json:"retailer"`
	PurchaseDate string `json:"purchaseDate"`
	PurchaseTime string `json:"purchaseTime"`
	Items        []Item `json:"items"`
	Total        string `json:"total"`
}

type Item struct {
	ShortDescription string `json:"shortDescription"`
	Price            string `json:"price"`
}

type PostReceiptID struct {
	ID string `json:"id"`
}

type GetReceiptPoints struct {
	Points int64 `json:"points"`
}
