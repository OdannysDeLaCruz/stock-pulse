package repository

import (
	"errors"
	"time"

	"github.com/OdannysDeLaCruz/stock-tracker/internal/domain/stock"
	"github.com/OdannysDeLaCruz/stock-tracker/internal/domain/price_history"
	"gorm.io/gorm"
)

type stockRepository struct {
	db *gorm.DB
}

func NewStockRepository(db *gorm.DB) stock.StockRepository {
	return &stockRepository{db}
}

func (r *stockRepository) FindAll() ([]stock.Stock, error) {
	var stocks []stock.Stock
	err := r.db.Preload("PriceHistory").Find(&stocks).Error
	return stocks, err
}

func (r *stockRepository) FindByTicker(ticker string) (*stock.Stock, error) {
	var stock stock.Stock
	result := r.db.Where("ticker = ?", ticker).Preload("PriceHistory").First(&stock)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &stock, result.Error
}

func (r *stockRepository) FindStockHistory(ticker string, startTime, endTime time.Time) ([]price_history.StockPriceHistory, error) {
	var history []price_history.StockPriceHistory
	err := r.db.Where("ticker = ? AND time BETWEEN ? AND ?", ticker, startTime, endTime).
		Order("time ASC").Find(&history).Error
	return history, err
}

func (r *stockRepository) FindRecommendedStocks() ([]stock.Stock, error) {
	var stocks []stock.Stock
    today := time.Now().Truncate(24 * time.Hour)

	err := r.db.Where("rating_to = ? AND target_to > target_from AND (action LIKE ? OR action LIKE ?) AND DATE(time AT TIME ZONE 'UTC') = ?",
        "Buy", "%target raised by%", "%upgraded to%", today.Format("2006-01-02")).Find(&stocks).Error

	if err != nil {
		return nil, err
	}

	var stockResponses []stock.Stock
	for _, stock := range stocks {
		var StockPriceHistory []price_history.StockPriceHistory
		err := r.db.Where("ticker = ?", stock.Ticker).
			Order("time ASC").
			Limit(10).
			Find(&StockPriceHistory).Error

		if err != nil {
			return nil, err
		}

		// Asignar historial de precios al stock
		stock.PriceHistory = StockPriceHistory
		stockResponses = append(stockResponses, stock)
	}

	return stockResponses, nil
}

func (r *stockRepository) FindNoRecommendedStocks() ([]stock.Stock, error) {
	var stocks []stock.Stock
	today := time.Now().Truncate(24 * time.Hour)

	err := r.db.Where("rating_to = ? AND target_to < target_from AND (action LIKE ? OR action LIKE ?) AND DATE(time AT TIME ZONE 'UTC') = ?",
		"Sell", "%target lowered by%", "%downgraded to%", today.Format("2006-01-02")).
		Preload("PriceHistory").Find(&stocks).Error

	if err != nil {
		return nil, err
	}

	// Mapear la respuesta con historial de precios
	var stockResponses []stock.Stock
	for _, stock := range stocks {
		var StockPriceHistory []price_history.StockPriceHistory
		err := r.db.Where("ticker = ?", stock.Ticker).
			Order("time ASC").
			Limit(10).
			Find(&StockPriceHistory).Error

		if err != nil {
			return nil, err
		}

		// Asignar historial de precios al stock
		stock.PriceHistory = StockPriceHistory
		stockResponses = append(stockResponses, stock)
	}

	return stockResponses, nil
}

func (r *stockRepository) SearchStocks(query string) ([]stock.Stock, error) {
	searchTerm := "%" + query + "%"

	var stocks []stock.Stock
	err := r.db.Where(
		"LOWER(ticker) LIKE ? OR LOWER(company) LIKE ? OR LOWER(brokerage) LIKE ?",
		searchTerm, searchTerm, searchTerm,
	).Find(&stocks).Error

	return stocks, err
}

func (r *stockRepository) Save(stock *stock.Stock) error {
	return r.db.Save(stock).Error
}

func (r *stockRepository) Update(stock *stock.Stock) error {
	return r.db.Save(stock).Error
}
