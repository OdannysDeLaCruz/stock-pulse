package stock

import (
	"time"

	"github.com/OdannysDeLaCruz/stock-pulse/internal/domain/price_history"
	"gorm.io/gorm"
)

// Stock representa la entidad de un stock
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

    // Relaciones
    PriceHistory []price_history.StockPriceHistory `gorm:"foreignKey:StockID;constraint:OnDelete:CASCADE"`
}