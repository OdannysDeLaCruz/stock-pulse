package repository

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/OdannysDeLaCruz/stock-pulse/internal/domain/stock"
	"github.com/OdannysDeLaCruz/stock-pulse/pkg"
	"gorm.io/gorm"
)

type stockExternalRepository struct {
	db *gorm.DB
}

type StockResponseAPI struct {
	Ticker      string `json:"ticker"`
	TargetFrom  string `json:"target_from"`
	TargetTo    string `json:"target_to"`
	Company     string `json:"company"`
	Action      string `json:"action"`
	Brokerage   string `json:"brokerage"`
	RatingFrom  string `json:"rating_from"`
	RatingTo    string `json:"rating_to"`
	Time        string `json:"time"`
}

type APIResponse struct {
	Items []StockResponseAPI `json:"items"`
	NextPage string   `json:"next_page"`
}

func NewStockExternalRepository(db *gorm.DB) stock.StockExternalRepository {
	return &stockExternalRepository{db}
}

func (r *stockExternalRepository) FetchData() ([]stock.Stock, error) {
	baseURL := os.Getenv("HOST_API")
	var allStocks []stock.Stock
	nextPage := ""
	maxStocks := 50

	for len(allStocks) < maxStocks {
		// Construir URL con el parámetro next_page si existe
		url := baseURL
		if nextPage != "" {
			url += "?next_page=" + nextPage
		}

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("error en la solicitud HTTP: %v", err)
		}

		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Authorization", "Bearer " + os.Getenv("API_KEY"))

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("error en la solicitud HTTP: %v", err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)

        fmt.Println(string(body))

		var response APIResponse
		err = json.Unmarshal(body, &response)
		if err != nil {
			return nil, fmt.Errorf("error al parsear JSON: %v", err)
		}

        var dataParsed []stock.Stock = make([]stock.Stock, len(response.Items))

        for i := range response.Items {
            formattedTime, _ := time.Parse(time.RFC3339, response.Items[i].Time)


            dataParsed[i].Action = response.Items[i].Action
            dataParsed[i].Brokerage = response.Items[i].Brokerage
            dataParsed[i].Company = response.Items[i].Company
            dataParsed[i].RatingFrom = response.Items[i].RatingFrom
            dataParsed[i].RatingTo = response.Items[i].RatingTo

            targetFrom, _ := pkg.ParsePrice(response.Items[i].TargetFrom)
            targetTo, _ := pkg.ParsePrice(response.Items[i].TargetTo)

            dataParsed[i].TargetFrom = targetFrom
            dataParsed[i].TargetTo = targetTo
            dataParsed[i].Ticker = response.Items[i].Ticker
            dataParsed[i].Time = formattedTime
        }

		// Agregar los items a la lista total
		allStocks = append(allStocks, dataParsed...)
        // fmt.Println(dataParsed)
		// Si no hay siguiente página o ya tenemos suficientes stocks, terminamos
		if response.NextPage == "" || len(allStocks) >= maxStocks {
			break
		}

		nextPage = response.NextPage
	}

	// Si tenemos más de 50 stocks, truncamos la lista
	if len(allStocks) > maxStocks {
		allStocks = allStocks[:maxStocks]
	}

	return allStocks, nil
}
