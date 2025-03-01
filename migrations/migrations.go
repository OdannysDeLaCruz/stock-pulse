package migrations

import (
	"log"
	"os"
	"time"

	"gorm.io/gorm"
)

type Stock struct {
	gorm.Model
	Ticker      string `gorm:"uniqueIndex"`
	TargetFrom  string `gorm:"column:target_from;not null"`
	TargetTo    string `gorm:"column:target_to;not null"`
	Company     string `gorm:"not null"`
	Action      string `gorm:"not null"`
	Brokerage   string `gorm:"not null"`
	RatingFrom  string `gorm:"column:rating_from;not null"`
	RatingTo    string `gorm:"column:rating_to;not null"`
	Time        string `gorm:"not null"`
	ChangePct    float64
	CurrentPrice float64
}

type StockPriceHistory struct {
	gorm.Model
	StockID     uint      `gorm:"index"`
	Ticker      string    `gorm:"index"`
	Price       float64
	Timestamp   time.Time `gorm:"index"`
}

func RunMigrations(db *gorm.DB) error {
	// Eliminar tabla Stocks solo si se especifica
	if os.Getenv("RESET_DB") == "true" {
		if err := db.Exec("DROP TABLE IF EXISTS stocks").Error; err != nil {
			return err
		}
		log.Println("Tabla stocks eliminada")
	}

	// Ejecutar migraciones
	if err := db.AutoMigrate(&Stock{}, &StockPriceHistory{}); err != nil {
		return err
	}

	log.Println("Migración completada")
	return nil
}