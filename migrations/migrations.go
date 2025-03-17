package migrations

import (
	"log"
	"os"
	"time"

	"gorm.io/gorm"
)

type Stock struct {
	gorm.Model
	Ticker      string `gorm:"column:ticker;uniqueIndex"`
	TargetFrom  float64 `gorm:"column:target_from;not null"`
	TargetTo    float64 `gorm:"column:target_to;not null"`
	Company     string `gorm:"column:company;not null"`
	Action      string `gorm:"column:action;not null"`
	Brokerage   string `gorm:"column:brokerage;not null"`
	RatingFrom  string `gorm:"column:rating_from;not null"`
	RatingTo    string `gorm:"column:rating_to;not null"`
	Time        time.Time `gorm:"column:time;not null;type:timestamp"`
}

type StockPriceHistory struct {
	gorm.Model
	StockID     uint      `gorm:"index"`
	Ticker      string    `gorm:"index"`
	TargetFrom  float64
	TargetTo    float64
	Time   		time.Time `gorm:"index;type:timestamp"`
}

func RunMigrations(db *gorm.DB) error {
	// Eliminar tabla Stocks solo si se especifica
	if os.Getenv("RESET_DB") == "true" {
		if err := db.Exec("DROP TABLE IF EXISTS stocks").Error; err != nil {
			return err
		}

		log.Println("Tabla stocks eliminada")

		if err := db.Exec("DROP TABLE IF EXISTS stock_price_histories").Error; err != nil {
			return err
		}

		log.Println("Tabla stock_price_histories eliminada")

		// Ejecutar migraciones
		if err := db.AutoMigrate(&Stock{}, &StockPriceHistory{}); err != nil {
			return err
		}

		log.Println("Migración completada")
	} else {
		log.Println("Migración no requerida")
	}


	return nil
}