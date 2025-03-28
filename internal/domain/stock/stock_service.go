package stock

import "time"

// StockService define los métodos de la capa de negocio
type StockService interface {
	GetAllStocks(page, limit int) (map[string]interface{}, error)
	GetStockByTicker(ticker string) (map[string]interface{}, error)
	GetStockByTickerRaw(ticker string) (*Stock, error)
	GetStockHistory(ticker string, startTime, endTime time.Time) ([]map[string]interface{}, error)
	GetRecommendedStocks() ([]map[string]interface{}, error)
	GetNoRecommendedStocks() ([]map[string]interface{}, error)
	SearchStocks(query string) ([]map[string]interface{}, error)
	UpdateStockData() error
	Save(stock *Stock) error
	CalculatePriceChange(targetFrom, targetTo float64) map[string]float64
}
