package delivery

import (
	"log"
	"net/http"
	"time"

	"github.com/OdannysDeLaCruz/stock-tracker/internal/domain/stock"
	"github.com/gin-gonic/gin"
)

type StockHandler struct {
	stockService stock.StockService
}

func NewStockHandler(router *gin.Engine, stockService stock.StockService) {
	handler := &StockHandler{stockService}

	// Definir rutas
	router.GET("/stocks", handler.GetStocks)
	router.PUT("/stocks", handler.UpdateStocks)
	router.GET("/stocks/:ticker", handler.GetStockByTicker)
	router.GET("/stocks/:ticker/history", handler.GetStockPriceHistory)
	router.GET("/stocks/recommendations", handler.GetRecommendedStocks)
	router.GET("/stocks/not-recommended", handler.GetNoRecommendedStocks)
	router.GET("/stocks/search", handler.SearchStocks)
}

func (h *StockHandler) GetStocks(c *gin.Context) {
	stocks, err := h.stockService.GetAllStocks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error obteniendo stocks"})
		return
	}
	c.JSON(http.StatusOK, stocks)
}

func (h *StockHandler) GetStockByTicker(c *gin.Context) {
	ticker := c.Param("ticker")
	stock, err := h.stockService.GetStockByTicker(ticker)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error obteniendo stock"})
		return
	}
	if stock == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Stock no encontrado"})
		return
	}
	c.JSON(http.StatusOK, stock)
}

// Obtener historial de precios
func (h *StockHandler) GetStockPriceHistory(c *gin.Context) {
	ticker := c.Param("ticker")
	period := c.DefaultQuery("period", "1d")

	startTime, endTime := calculateTimeRange(period)

	history, err := h.stockService.GetStockHistory(ticker, startTime, endTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error obteniendo historial"})
		return
	}
	c.JSON(http.StatusOK, history)
}

// Obtener stocks recomendados
func (h *StockHandler) GetRecommendedStocks(c *gin.Context) {
	stocks, err := h.stockService.GetRecommendedStocks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error obteniendo recomendaciones"})
		return
	}

	c.JSON(http.StatusOK, stocks)
}

// Obtener stocks no recomendados
func (h *StockHandler) GetNoRecommendedStocks(c *gin.Context) {
	stocks, err := h.stockService.GetNoRecommendedStocks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error obteniendo stocks no recomendados"})
		return
	}

	

	c.JSON(http.StatusOK, stocks)
}

// Buscar stocks por query
func (h *StockHandler) SearchStocks(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Parámetro de búsqueda requerido"})
		return
	}

	stocks, err := h.stockService.SearchStocks(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error en la búsqueda"})
		return
	}

	if len(stocks) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "No se encontraron stocks"})
		return
	}

	c.JSON(http.StatusOK, stocks)
}

// Actualizar base de datos manualmente
func (h *StockHandler) UpdateStocks(c *gin.Context) {
	err := h.stockService.UpdateStockData()
	if err != nil {
		log.Println("Error actualizando base de datos manualmente:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al actualizar base de datos"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Base de datos actualizada correctamente"})
}

// Calcular rango de tiempo basado en el periodo
func calculateTimeRange(period string) (time.Time, time.Time) {
	endTime := time.Now()
	var startTime time.Time

	switch period {
	case "1d":
		startTime = endTime.AddDate(0, 0, -1)
	case "1w":
		startTime = endTime.AddDate(0, 0, -7)
	case "1m":
		startTime = endTime.AddDate(0, -1, 0)
	case "3m":
		startTime = endTime.AddDate(0, -3, 0)
	case "1y":
		startTime = endTime.AddDate(-1, 0, 0)
	default:
		startTime = endTime.AddDate(0, 0, -1)
	}

	return startTime, endTime
}