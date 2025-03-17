package price_history

import (
	"time"

	"gorm.io/gorm"
)

type StockPriceHistory struct {
	gorm.Model
	StockID     uint    	`gorm:"index"`
	Ticker      string  	`gorm:"index" json:"ticker"`
	TargetFrom  float64   	`json:"target_from"`
	TargetTo    float64   	`json:"target_to"`
	Time        time.Time 	`gorm:"index" json:"time"`
}