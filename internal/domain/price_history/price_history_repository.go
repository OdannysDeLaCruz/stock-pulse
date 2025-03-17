package price_history

import "time"

// StockRepository define las operaciones que debe soportar cualquier implementación de almacenamiento
type StockPriceHistoryRepository interface {
	FindAll(ticker string, startTime, endTime time.Time) ([]StockPriceHistory, error)
	Create(stock *StockPriceHistory) error
}
