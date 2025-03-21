package main

import (
	"log"

	"github.com/OdannysDeLaCruz/stock-pulse/config"
	"github.com/OdannysDeLaCruz/stock-pulse/internal/application"
	"github.com/OdannysDeLaCruz/stock-pulse/internal/delivery"
	"github.com/OdannysDeLaCruz/stock-pulse/internal/repository"
	"github.com/OdannysDeLaCruz/stock-pulse/pkg"
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
	stockExternalRepo := repository.NewStockExternalRepository(db)
	StockPriceHistoryRepo := repository.NewStockPriceHistoryRepository(db)
	stockService := application.NewStockService(stockRepo, stockExternalRepo)
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
