package stock

import (
	"time"
)

type StockPartial struct {
	Ticker              string			`json:"ticker"`
	TargetFrom          float64			`json:"target_from"`
	TargetTo            float64			`json:"target_to"`
	RatingFrom          string			`json:"rating_from"`
	RatingTo            string			`json:"rating_to"`
	NewItemStockPriceHistory NewItemStockPriceHistory 	`json:"new_item_price_history"`
	Time                time.Time       `json:"time"`
}

type NewItemStockPriceHistory struct {
	TargetFrom  float64   `json:"target_from"`
	TargetTo    float64   `json:"target_to"`
	RatingFrom  string    `json:"rating_from"`
	RatingTo    string    `json:"rating_to"`
	Time        time.Time    `json:"time"`
}