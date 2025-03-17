package price_history

import "time"

type StockPriceHistoryService interface {
	GetAll(ticker string, startTime, endTime time.Time) ([]StockPriceHistory, error)
	Create(StockPriceHistory *StockPriceHistory) error
}
