package application

import (
	"time"

	"github.com/OdannysDeLaCruz/stock-pulse/internal/domain/price_history"
)

// StockPriceHistoryService implementa price_history.StockService
type StockPriceHistoryService struct {
	StockPriceHistoryRepo price_history.StockPriceHistoryRepository
}

// NewStockService devuelve una nueva instancia de StockService
func NewStockPriceHistoryService(StockPriceHistoryRepo price_history.StockPriceHistoryRepository) price_history.StockPriceHistoryService {
	return &StockPriceHistoryService{StockPriceHistoryRepo}
}

func (s *StockPriceHistoryService) GetAll(ticker string, startTime, endTime time.Time) ([]price_history.StockPriceHistory, error) {
	StockPriceHistory, err := s.StockPriceHistoryRepo.FindAll(ticker, startTime, endTime)
	if err != nil {
		return nil, err
	}

	return StockPriceHistory, nil
}

func (s *StockPriceHistoryService) Create(StockPriceHistory *price_history.StockPriceHistory) error {
	return s.StockPriceHistoryRepo.Create(StockPriceHistory)
}