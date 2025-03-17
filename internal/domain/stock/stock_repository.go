package stock

import (
	"time"
	"github.com/OdannysDeLaCruz/stock-tracker/internal/domain/price_history"
)

// StockRepository define las operaciones que debe soportar cualquier implementación de almacenamiento
type StockRepository interface {
	FindAll() ([]Stock, error)
	FindByTicker(ticker string) (*Stock, error)
	FindStockHistory(ticker string, startTime, endTime time.Time) ([]price_history.StockPriceHistory, error)
	FindRecommendedStocks() ([]Stock, error)
	FindNoRecommendedStocks() ([]Stock, error)
	SearchStocks(query string) ([]Stock, error)
	Save(stock *Stock) error
	Update(stock *Stock) error
}