package application

import (
	"time"

	"github.com/OdannysDeLaCruz/stock-pulse/internal/domain/stock"
	"github.com/OdannysDeLaCruz/stock-pulse/pkg"
)

// stockService implementa stock.StockService
type stockService struct {
	stockRepo stock.StockRepository
	stockExternalRepo stock.StockExternalRepository
}

// NewStockService devuelve una nueva instancia de StockService
func NewStockService(stockRepo stock.StockRepository, stockExternalRepo stock.StockExternalRepository) stock.StockService {
	return &stockService{
		stockRepo,
		stockExternalRepo,
	}
}

func (s *stockService) GetAllStocks(page, limit int) ([]map[string]interface{}, error) {
	// Validar límites (máximo 15 stocks por página)
	if limit <= 0 || limit > 15 {
		limit = 15
	}
	if page < 1 {
		page = 1
	}

	stocks, err := s.stockRepo.FindAll(page, limit)
	if err != nil {
		return nil, err
	}
	// Mapear los resultados a JSON
	var stockResponses []map[string]interface{}
	for _, stock := range stocks {
		var StockPriceHistoryMapped []map[string]interface{} = make([]map[string]interface{}, 0)

		historyLen := len(stock.PriceHistory)
		startIdx := 0
		if historyLen > 10 {
			startIdx = historyLen - 10
		}

		for _, h := range stock.PriceHistory[startIdx:] {
			StockPriceHistoryMapped = append(StockPriceHistoryMapped, map[string]interface{}{
				"time":        h.Time,
				"target_to":   pkg.RoundToTwoDecimals(h.TargetTo),
				"target_from": pkg.RoundToTwoDecimals(h.TargetFrom),
			})
		}

		stockResponse := map[string]interface{}{
			"id":            stock.ID,
			"ticker":        stock.Ticker,
			"company":       stock.Company,
			"action":        stock.Action,
			"brokerage":     stock.Brokerage,
			"rating_from":   stock.RatingFrom,
			"rating_to":     stock.RatingTo,
			"time":          stock.Time,
			"target_from":   pkg.RoundToTwoDecimals(stock.TargetFrom),
			"target_to":     pkg.RoundToTwoDecimals(stock.TargetTo),
			"price_history": StockPriceHistoryMapped,
			"analysis":      s.CalculatePriceChange(stock.TargetFrom, stock.TargetTo),
		}

		stockResponses = append(stockResponses, stockResponse)
	}

	return stockResponses, nil
}

func (s *stockService) GetStockByTicker(ticker string) (map[string]interface{}, error) {
	stock, err := s.stockRepo.FindByTicker(ticker)

	if err != nil {
		return nil, err
	}

	// Mapear los resultados a JSON
	var StockPriceHistoryMapped []map[string]interface{} = make([]map[string]interface{}, 0)

	historyLen := len(stock.PriceHistory)
		startIdx := 0
		if historyLen > 10 {
			startIdx = historyLen - 10
		}

		for _, h := range stock.PriceHistory[startIdx:] {
		StockPriceHistoryMapped = append(StockPriceHistoryMapped, map[string]interface{}{
			"time":        h.Time,
			"target_to":   pkg.RoundToTwoDecimals(h.TargetTo),
			"target_from": pkg.RoundToTwoDecimals(h.TargetFrom),
		})
	}

	stockResponse := map[string]interface{}{
		"id":            stock.ID,
		"ticker":        stock.Ticker,
		"company":       stock.Company,
		"action":        stock.Action,
		"brokerage":     stock.Brokerage,
		"rating_from":   stock.RatingFrom,
		"rating_to":     stock.RatingTo,
		"time":          stock.Time,
		"target_from":   pkg.RoundToTwoDecimals(stock.TargetFrom),
		"target_to":     pkg.RoundToTwoDecimals(stock.TargetTo),
		"price_history": StockPriceHistoryMapped,
		"analysis":      s.CalculatePriceChange(stock.TargetFrom, stock.TargetTo),
	}


	return stockResponse, nil

}

func (s *stockService) GetStockByTickerRaw(ticker string) (*stock.Stock, error) {
	stock, err := s.stockRepo.FindByTicker(ticker)
	if err != nil {
		return nil, err
	}

	return stock, nil
}


func (s *stockService) GetStockHistory(ticker string, startTime, endTime time.Time) ([]map[string]interface{}, error) {

	StockPriceHistory, err := s.stockRepo.FindStockHistory(ticker, startTime, endTime)

	if err != nil {
		return nil, err
	}

	var response []map[string]interface{}
    for _, h := range StockPriceHistory {
        response = append(response, map[string]interface{}{
            "time": h.Time,
            "target_to": pkg.RoundToTwoDecimals(h.TargetTo),
            "target_from": pkg.RoundToTwoDecimals(h.TargetFrom),
        })
    }

	return response, nil
}

func (s *stockService) GetRecommendedStocks() ([]map[string]interface{}, error) {
	stocks, err := s.stockRepo.FindRecommendedStocks()
	if err != nil {
		return nil, err
	}

	// Mapear los resultados a JSON
	var stockResponses []map[string]interface{}
	for _, stock := range stocks {
		var StockPriceHistoryMapped []map[string]interface{} = make([]map[string]interface{}, 0)

		historyLen := len(stock.PriceHistory)
		startIdx := 0
		if historyLen > 10 {
			startIdx = historyLen - 10
		}

		for _, h := range stock.PriceHistory[startIdx:] {
			StockPriceHistoryMapped = append(StockPriceHistoryMapped, map[string]interface{}{
				"time":        h.Time,
				"target_to":   pkg.RoundToTwoDecimals(h.TargetTo),
				"target_from": pkg.RoundToTwoDecimals(h.TargetFrom),
			})
		}

		stockResponse := map[string]interface{}{
			"id":            stock.ID,
			"ticker":        stock.Ticker,
			"company":       stock.Company,
			"action":        stock.Action,
			"brokerage":     stock.Brokerage,
			"rating_from":   stock.RatingFrom,
			"rating_to":     stock.RatingTo,
			"time":          stock.Time,
			"target_from":   pkg.RoundToTwoDecimals(stock.TargetFrom),
			"target_to":     pkg.RoundToTwoDecimals(stock.TargetTo),
			"price_history": StockPriceHistoryMapped,
			"analysis":      s.CalculatePriceChange(stock.TargetFrom, stock.TargetTo),
		}

		stockResponses = append(stockResponses, stockResponse)
	}

	return stockResponses, nil
}

func (s *stockService) GetNoRecommendedStocks() ([]map[string]interface{}, error) {
	stocks, err := s.stockRepo.FindNoRecommendedStocks()
	if err != nil {
		return nil, err
	}

	// Mapear los resultados a JSON
	var stockResponses []map[string]interface{}
	for _, stock := range stocks {
		var StockPriceHistoryMapped []map[string]interface{} = make([]map[string]interface{}, 0)

		historyLen := len(stock.PriceHistory)
		startIdx := 0
		if historyLen > 10 {
			startIdx = historyLen - 10
		}

		for _, h := range stock.PriceHistory[startIdx:] {
			StockPriceHistoryMapped = append(StockPriceHistoryMapped, map[string]interface{}{
				"time":        h.Time,
				"target_to":   pkg.RoundToTwoDecimals(h.TargetTo),
				"target_from": pkg.RoundToTwoDecimals(h.TargetFrom),
			})
		}

		stockResponse := map[string]interface{}{
			"id":            stock.ID,
			"ticker":        stock.Ticker,
			"company":       stock.Company,
			"action":        stock.Action,
			"brokerage":     stock.Brokerage,
			"rating_from":   stock.RatingFrom,
			"rating_to":     stock.RatingTo,
			"time":          stock.Time,
			"target_from":   pkg.RoundToTwoDecimals(stock.TargetFrom),
			"target_to":     pkg.RoundToTwoDecimals(stock.TargetTo),
			"price_history": StockPriceHistoryMapped,
			"analysis":      s.CalculatePriceChange(stock.TargetFrom, stock.TargetTo),
		}

		stockResponses = append(stockResponses, stockResponse)
	}

	return stockResponses, nil
}

func (s *stockService) SearchStocks(query string) ([]map[string]interface{}, error) {
	stocks, err := s.stockRepo.SearchStocks(query)

	if err != nil {
		return nil, err
	}

	// Mapear los resultados a JSON
	var stockResponses []map[string]interface{}
	for _, stock := range stocks {
		var StockPriceHistoryMapped []map[string]interface{} = make([]map[string]interface{}, 0)
	
		historyLen := len(stock.PriceHistory)
		startIdx := 0
		if historyLen > 10 {
			startIdx = historyLen - 10
		}

		for _, h := range stock.PriceHistory[startIdx:] {
			StockPriceHistoryMapped = append(StockPriceHistoryMapped, map[string]interface{}{
				"time":        h.Time,
				"target_to":   pkg.RoundToTwoDecimals(h.TargetTo),
				"target_from": pkg.RoundToTwoDecimals(h.TargetFrom),
			})
		}

		stockResponse := map[string]interface{}{
			"id":            stock.ID,
			"ticker":        stock.Ticker,
			"company":       stock.Company,
			"action":        stock.Action,
			"brokerage":     stock.Brokerage,
			"rating_from":   stock.RatingFrom,
			"rating_to":     stock.RatingTo,
			"time":          stock.Time,
			"target_from":   pkg.RoundToTwoDecimals(stock.TargetFrom),
			"target_to":     pkg.RoundToTwoDecimals(stock.TargetTo),
			"price_history": StockPriceHistoryMapped,
			"analysis":      s.CalculatePriceChange(stock.TargetFrom, stock.TargetTo),
		}

		stockResponses = append(stockResponses, stockResponse)
	}

	return stockResponses, nil
}

func (s *stockService) UpdateStockData() error {
	stocks, err := s.stockExternalRepo.FetchData()
	if err != nil {
		return err
	}

	err = s.stockRepo.Upsert(stocks)
	if err != nil {
		return err
	}

	return nil
}

func (s *stockService) Save(stock *stock.Stock) error {
	return s.stockRepo.Save(stock)
}

func (s *stockService) CalculatePriceChange(targetFrom, targetTo float64) map[string]float64 {
	if targetFrom != 0 && targetTo != 0 {
		changeValue := targetTo - targetFrom
		changePercentage := (changeValue / targetFrom) * 100

		return map[string]float64{
			"change_value":      pkg.RoundToTwoDecimals(changeValue),
			"change_percentage": pkg.RoundToTwoDecimals(changePercentage),
		}
	}

	return map[string]float64{
		"change_value":      0,
		"change_percentage": 0,
	}
}