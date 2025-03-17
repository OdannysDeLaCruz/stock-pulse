package main

import (
	"log"

	"github.com/OdannysDeLaCruz/stock-tracker/config"
	"github.com/OdannysDeLaCruz/stock-tracker/internal/application"
	"github.com/OdannysDeLaCruz/stock-tracker/internal/delivery"
	"github.com/OdannysDeLaCruz/stock-tracker/internal/repository"
	"github.com/OdannysDeLaCruz/stock-tracker/pkg"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg := config.LoadConfig()

	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Error al conectar a la BD:", err)
	}

	// Conectar a Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddress,
	})

	// Inyectar dependencias
	stockRepo := repository.NewStockRepository(db)
	StockPriceHistoryRepo := repository.NewStockPriceHistoryRepository(db)
	stockService := application.NewStockService(stockRepo)
	StockPriceHistoryService := application.NewStockPriceHistoryService(StockPriceHistoryRepo)

	// Iniciar servidor
	router := gin.Default()
	router.Use(pkg.CORSMiddleware())

	// Rutas HTTP
	delivery.NewStockHandler(router, stockService)

	// WebSockets
	delivery.NewWebSocketHandler(cfg, router, redisClient, stockService, StockPriceHistoryService)

	router.Run(":" + cfg.Port)
}
