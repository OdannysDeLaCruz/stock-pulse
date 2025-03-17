package repository

import (
	"time"

	"github.com/OdannysDeLaCruz/stock-tracker/internal/domain/price_history"
	"gorm.io/gorm"
)

type StockPriceHistoryRepository struct {
	db *gorm.DB
}

func NewStockPriceHistoryRepository(db *gorm.DB) price_history.StockPriceHistoryRepository {
	return &StockPriceHistoryRepository{db}
}

func (r *StockPriceHistoryRepository) FindAll(ticker string, startTime, endTime time.Time) ([]price_history.StockPriceHistory, error) {
	var StockPriceHistory []price_history.StockPriceHistory

	err := r.db.Where("ticker = ? AND time BETWEEN ? AND ?",
	ticker, startTime, endTime).
	Order("time ASC").
	Find(&StockPriceHistory).Error

	return StockPriceHistory, err
}


func (r *StockPriceHistoryRepository) Create(StockPriceHistory *price_history.StockPriceHistory) error {
	return r.db.Create(StockPriceHistory).Error
}
