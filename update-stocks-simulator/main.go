package main

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"time"

	"github.com/redis/go-redis/v9"
)

type StockPartialRedis struct {
	Ticker      string    `json:"ticker"`
	TargetFrom  float64   `json:"target_from"`
	TargetTo    float64   `json:"target_to"`
	RatingFrom  string    `json:"rating_from"`
	RatingTo    string    `json:"rating_to"`
	Time        string    `json:"time"`
}

func main() {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	ctx := context.Background()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// Lista de acciones simuladas
	stocksTicker := []string{"MOMO"}
	ratingList := []string{"Buy", "Hold", "Sell", "Strong-Buy", "Strong-Sell", "Neutral", "Overweight", "Underweight", "Outperform", "Underperform"}

	for {
		<-ticker.C

		// Seleccionar una acción al azar y generar un nuevo precio
		ticker := stocksTicker[rand.Intn(len(stocksTicker))]
		stock := StockPartialRedis{
			Ticker: ticker,
			TargetFrom: 50 + rand.Float64()*50,
			TargetTo:  100 + rand.Float64()*50,
			RatingFrom: ratingList[rand.Intn(len(ratingList))],
			RatingTo:   ratingList[rand.Intn(len(ratingList))],
			Time:   time.Now().Format(time.RFC3339),
		}

		// Publicar en Redis
		data, _ := json.Marshal(stock)
		redisClient.Publish(ctx, "stock_updates", data)

		log.Println("🔹 Stock enviado:", string(data))
	}
}
