package main

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"time"

	"github.com/redis/go-redis/v9"
)

type NewItemPriceHistory struct {
	TargetFrom  float64   `json:"target_from"`
	TargetTo    float64   `json:"target_to"`
	RatingFrom  string    `json:"rating_from"`
	RatingTo    string    `json:"rating_to"`
	Time        string    `json:"time"`
}
type StockPartialRedis struct {
	Ticker      string    `json:"ticker"`
	TargetFrom  float64   `json:"target_from"`
	TargetTo    float64   `json:"target_to"`
	RatingFrom  string    `json:"rating_from"`
	RatingTo    string    `json:"rating_to"`
	NewItemPriceHistory NewItemPriceHistory `json:"new_item_price_history"`
}

func main() {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	ctx := context.Background()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// Lista de acciones simuladas
	stocksTicker := []string{"AKBA"}
	ratingList := []string{"Buy", "Hold", "Sell", "Strong-Buy", "Strong-Sell", "Neutral", "Overweight", "Underweight", "Outperform", "Underperform"}

	for {
		<-ticker.C

		// Seleccionar una acción al azar y generar un nuevo precio
		ticker := stocksTicker[rand.Intn(len(stocksTicker))]

		targetFrom := 100 + 50 + rand.Float64()*50
		targetTo := 100 + rand.Float64()*50

		ratingFrom := ratingList[rand.Intn(len(ratingList))]
		ratingTo := ratingList[rand.Intn(len(ratingList))]

		dateFormat := "2006-01-02T15:04:05.000000000Z"
		dateUTC := time.Now().UTC()
		time := dateUTC.Format(dateFormat)

		newItemPriceHistory := NewItemPriceHistory{
			TargetFrom: targetFrom,
			TargetTo:   targetTo,
			RatingFrom: ratingFrom,
			RatingTo:   ratingTo,
			Time:       time,
		}

		stock := StockPartialRedis{
			Ticker: ticker,
			TargetFrom: targetFrom,
			TargetTo:   targetTo,
			RatingFrom: ratingFrom,
			RatingTo:  	ratingTo,
			NewItemPriceHistory: newItemPriceHistory,
		}

		// Publicar en Redis
		data, _ := json.Marshal(stock)
		redisClient.Publish(ctx, "stock_updates", data)

		log.Println("🔹 Stock enviado:", string(data))
	}
}
