package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"slices"

	"net/http"
	"os"

	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

// Dominios permitidos para el CORS

// Obtener origin desde variable de entorno
func getOrigins() map[string]bool {
	origins := os.Getenv("ALLOWED_ORIGINS")
	DEFAULT_ALLOWED_ORIGINS := "http://localhost:8080"

	allowedOrigins := make(map[string]bool)

	// Si no hay variables de entorno, usar el valor por defecto
	if origins == "" {
		allowedOrigins[DEFAULT_ALLOWED_ORIGINS] = true
		return allowedOrigins
	}

	// Dividir la cadena de orígenes en un slice y agregarlos al mapa
	for _, origin := range strings.Split(origins, ",") {
		allowedOrigins[strings.TrimSpace(origin)] = true
	}

	return allowedOrigins
}

var allowedOrigins = map[string]bool{}

// Models

type Stock struct {
	gorm.Model
	Ticker      string `gorm:"column:ticker;uniqueIndex"`
	TargetFrom  float64 `gorm:"column:target_from;not null"`
	TargetTo    float64 `gorm:"column:target_to;not null"`
	Company     string `gorm:"column:company;not null"`
	Action      string `gorm:"column:action;not null"`
	Brokerage   string `gorm:"column:brokerage;not null"`
	RatingFrom  string `gorm:"column:rating_from;not null"`
	RatingTo    string `gorm:"column:rating_to;not null"`
	Time        time.Time `gorm:"column:time;not null;type:timestamp"`

    // Relaciones
    PriceHistory []StockPriceHistory `gorm:"foreignKey:StockID;constraint:OnDelete:CASCADE" json:"price_history"`
}

type NewItemPriceHistory struct {
	TargetFrom  float64   `json:"target_from"`
	TargetTo    float64   `json:"target_to"`
	RatingFrom  string    `json:"rating_from"`
	RatingTo    string    `json:"rating_to"`
	Time        time.Time    `json:"time"`
}
type StockPartialRedis struct {
	Ticker      string    `json:"ticker"`
	TargetFrom  float64   `json:"target_from"`
	TargetTo    float64   `json:"target_to"`
	RatingFrom  string    `json:"rating_from"`
	RatingTo    string    `json:"rating_to"`
	NewItemPriceHistory NewItemPriceHistory `json:"new_item_price_history"`
	Time        time.Time    `json:"time"`
}

type StockPriceHistory struct {
    gorm.Model
    StockID     uint    `gorm:"index"`
    Ticker      string  `gorm:"index" json:"ticker"`
    TargetFrom  float64 `json:"target_from"`
    TargetTo    float64 `json:"target_to"`
    Time        time.Time  `gorm:"index" json:"time"`
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

// Estructura para manejar las conexiones WebSocket
type WebSocketHandler struct {
    clients        map[*websocket.Conn]string
    broadcast        chan []Stock
    register   chan *websocket.Conn
    unregister chan *websocket.Conn
    mutex         sync.Mutex
    redisClient    *redis.Client
}

// Configuración del upgrader de WebSocket
var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool {
        origin := r.Header.Get("Origin")
		_, exists := allowedOrigins[origin]

        return exists
    },
}

// Crear el handler de WebSocket
func NewWebSocketHandler(redisClient *redis.Client) *WebSocketHandler {
    return &WebSocketHandler{
        clients:        make(map[*websocket.Conn]string),
        broadcast:  make(chan []Stock),
        register:   make(chan *websocket.Conn),
        unregister: make(chan *websocket.Conn),
        redisClient:    redisClient,
    }
}

func (w *WebSocketHandler) handleConnections(c *gin.Context) {
	ticker := c.Request.URL.Query().Get("ticker")
	if ticker == "" {
		http.Error(c.Writer, "Ticker is required", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Error al actualizar la conexión a WebSocket:", err)
		return
	}
	defer conn.Close()

	w.mutex.Lock()
	w.clients[conn] = ticker
	w.mutex.Unlock()

    // Leer los mensajes de la conexión WebSocket
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			w.mutex.Lock()
			delete(w.clients, conn)
			w.mutex.Unlock()
			break
		}
	}
}

// Escuchar eventos de Redis y enviarlos a los clientes WS
func (w *WebSocketHandler) subscribeStockUpdates() {
	ctx := context.Background()
	pubsub := w.redisClient.Subscribe(ctx, "stock_updates")
	defer pubsub.Close()

	for {
		msg, err := pubsub.ReceiveMessage(ctx)
		if err != nil {
			log.Println("Error en Redis Pub/Sub:", err)
			continue
		}

        log.Println("🔹 Stock recibido:", msg.Payload)

		var stock StockPartialRedis
		err = json.Unmarshal([]byte(msg.Payload), &stock)
        if err != nil {
			log.Println("Error al parsear stock:", err)
			continue
		}

		// Guardar en base de datos antes de enviar a clientes
		w.saveStockUpdate(stock)

        priceChange := CalculatePriceChange(stock.TargetFrom, stock.TargetTo)

        response := map[string]interface{}{
			"ticker":      stock.Ticker,
			"target_from": roundToTwoDecimals(stock.TargetFrom),
			"target_to":   roundToTwoDecimals(stock.TargetTo),
			"rating_from": stock.RatingFrom,
			"rating_to":   stock.RatingTo,
			"new_item_price_history": NewItemPriceHistory {
                TargetFrom:  roundToTwoDecimals(stock.NewItemPriceHistory.TargetFrom),
                TargetTo:    roundToTwoDecimals(stock.NewItemPriceHistory.TargetTo),
                Time:        stock.NewItemPriceHistory.Time,
            },
            "analysis": priceChange,
		}

        // Convertir a JSON
		customPayload, err := json.Marshal(response)
		if err != nil {
			log.Println("Error al serializar respuesta personalizada:", err)
			continue
		}

		// Enviar datos solo a clientes suscritos a ese símbolo
		w.mutex.Lock()
		for client, subTicker := range w.clients {
			if subTicker == stock.Ticker {
				client.WriteMessage(websocket.TextMessage, customPayload)
			}
		}
		w.mutex.Unlock()
	}
}

// Guardar actualización en la base de datos CockroachDB
func (w *WebSocketHandler) saveStockUpdate(stockPartial StockPartialRedis) error {

    existingStock := Stock{}
    result := DB.Where("ticker = ?", stockPartial.Ticker).First(&existingStock)

    if result.Error != nil {
        if errors.Is(result.Error, gorm.ErrRecordNotFound) {
           log.Println("No se encontró ningún registro con ese ticker")
        } else {
            log.Println("Error en la consulta:", result.Error)
        }
    } else {
        // Actualizar el registro existente
        existingStock.TargetFrom  = stockPartial.TargetFrom
        existingStock.TargetTo    = stockPartial.TargetTo
        existingStock.RatingFrom  = stockPartial.RatingFrom
        existingStock.RatingTo    = stockPartial.RatingTo
        existingStock.Time        = stockPartial.NewItemPriceHistory.Time.Truncate(time.Microsecond)

        if err := DB.Save(&existingStock).Error; err != nil {
            return err
        }

        // Guardar historial de precios
        priceHistory := StockPriceHistory{
            StockID:     existingStock.ID,
            Ticker:      existingStock.Ticker,
            TargetFrom:  stockPartial.TargetFrom,
            TargetTo:    stockPartial.TargetTo,
            Time:        stockPartial.NewItemPriceHistory.Time.Truncate(time.Microsecond),
        }

        if err := DB.Create(&priceHistory).Error; err != nil {
            log.Printf("Error al guardar historial de precios: %v", err)
        }
    }

    return nil
}

// Convierte un string "$13.00" a float64
func parsePrice(priceStr string) (float64, error) {
	cleanStr := strings.Replace(priceStr, "$", "", 1) // Eliminar $
	cleanStr = strings.Replace(cleanStr, ",", "", 1) // Eliminar ,
	value, err := strconv.ParseFloat(cleanStr, 64)   // Convertir a float64
	if err != nil {
        fmt.Println("Error al convertir el precio:", err)
		return 0, err
	}

	return value, nil
}

// Services
func InitDB() {
	dsn := os.Getenv("DATABASE_URL")
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("No se pudo conectar a CockroachDB:", err)
	}

	DB = database
	fmt.Println("Conectado a CockroachDB con GORM")
}

// Obtener datos de la API externa - seeding
func FetchStockData() ([]Stock, error) {
	baseURL := os.Getenv("HOST_API")
	var allStocks []Stock
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

        var dataParsed []Stock = make([]Stock, len(response.Items))

        for i := range response.Items {
            formattedTime, _ := time.Parse(time.RFC3339, response.Items[i].Time)


            dataParsed[i].Action = response.Items[i].Action
            dataParsed[i].Brokerage = response.Items[i].Brokerage
            dataParsed[i].Company = response.Items[i].Company
            dataParsed[i].RatingFrom = response.Items[i].RatingFrom
            dataParsed[i].RatingTo = response.Items[i].RatingTo

            targetFrom, _ := parsePrice(response.Items[i].TargetFrom)
            targetTo, _ := parsePrice(response.Items[i].TargetTo)

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

// Guardar datos de la API externa en la base de datos
func SaveStockData(stocks []Stock) error {
	for _, stockData := range stocks {
		existingStock := Stock{}
		result := DB.Where("ticker = ?", stockData.Ticker).First(&existingStock)

        if result.Error != nil {
            if errors.Is(result.Error, gorm.ErrRecordNotFound) {
                // No se encontró ningún registro con ese ticker
                // Insertar uno nuevo
                newStock := Stock{
                    Ticker:     stockData.Ticker,
                    TargetFrom: stockData.TargetFrom,
                    TargetTo:   stockData.TargetTo,
                    Company:    stockData.Company,
                    Action:     stockData.Action,
                    Brokerage:  stockData.Brokerage,
                    RatingFrom: stockData.RatingFrom,
                    RatingTo:   stockData.RatingTo,
                    Time:       stockData.Time.Truncate(time.Microsecond),
                }
                if err := DB.Create(&newStock).Error; err != nil {
                    return err
                }
            } else {
                fmt.Println("Error en la consulta:", result.Error)
                return result.Error
            }
        } else {
            // Actualizar el registro existente
            existingStock.TargetFrom = stockData.TargetFrom
            existingStock.TargetTo = stockData.TargetTo
            existingStock.Company = stockData.Company
            existingStock.Action = stockData.Action
            existingStock.Brokerage = stockData.Brokerage
            existingStock.RatingFrom = stockData.RatingFrom
            existingStock.RatingTo = stockData.RatingTo
            existingStock.Time = stockData.Time.Truncate(time.Microsecond)

            if err := DB.Save(&existingStock).Error; err != nil {
                return err
            }
		}
	}

	return nil
}

// Seeding de datos en la base de datos
func GetStockAndStoreInDB() ([]Stock, error) {
	stocks, err := FetchStockData()
	if err != nil {
		return nil, err
	}

	err = SaveStockData(stocks)
	if err != nil {
		return nil, err
	}

	return stocks, nil
}

// Función para calcular la diferencia en valores y porcentaje
func CalculatePriceChange(targetFrom, targetTo float64) map[string]interface{} {
    if targetFrom != 0 && targetTo != 0 {
        changeValue := targetTo - targetFrom
        changePercentage := (changeValue / targetFrom) * 100

        return map[string]interface{}{
            "change_value":      roundToTwoDecimals(changeValue),
            "change_percentage": roundToTwoDecimals(changePercentage),
        }
    }

    return map[string]interface{}{
        "change_value":      0,
        "change_percentage": 0,
    }
}

func roundToTwoDecimals(value float64) float64 {
    return math.Round(value*100) / 100
}

// Handlers para las rutas

func GetStocks(c *gin.Context) {
    var stocks []Stock
    DB.Find(&stocks)

    var stockResponses []map[string]interface{}

    for _, stock := range stocks {
        var priceHistory []StockPriceHistory
        DB.Where("ticker = ?", stock.Ticker).Order("time DESC").Limit(10).Find(&priceHistory)

        slices.Reverse(priceHistory)

        // Formatear datos para la gráfica
        var priceHistoryMapped = make([]map[string]interface{}, 0)
        for _, h := range priceHistory {
            priceHistoryMapped = append(priceHistoryMapped, map[string]interface{}{
                "time": h.Time,
                "target_to": roundToTwoDecimals(h.TargetTo),
                "target_from": roundToTwoDecimals(h.TargetFrom),
            })
        }

        priceChange := CalculatePriceChange(stock.TargetFrom, stock.TargetTo)

        // formatear a 2 decimales
        stockResponse := map[string]interface{}{
            "id":           stock.ID,
            "ticker":       stock.Ticker,
            "company":      stock.Company,
            "action":       stock.Action,
            "brokerage":    stock.Brokerage,
            "rating_from":  stock.RatingFrom,
            "rating_to":    stock.RatingTo,
            "time":         stock.Time,
            "target_from":  roundToTwoDecimals(stock.TargetFrom),
            "target_to":    roundToTwoDecimals(stock.TargetTo),
            "price_history": priceHistoryMapped,
            "analysis":      priceChange,
        }

        stockResponses = append(stockResponses, stockResponse)
    }

    c.JSON(http.StatusOK, stockResponses)
}

func GetStockByTicker(c *gin.Context) {
    ticker := c.Param("ticker") // Obtiene el ticker de la URL

    // Buscar en la base de datos
    var stock Stock
    result := DB.Where("ticker = ?", ticker).First(&stock)
    if result.Error != nil {
        if result.Error == gorm.ErrRecordNotFound {
            c.JSON(http.StatusNotFound, gin.H{
                "error": "Stock no encontrado",
                "ticker": ticker,
            })
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Error al buscar el stock",
        })
        return
    }

    var priceHistory []StockPriceHistory
    result = DB.Where("ticker = ?", stock.Ticker).Order("time ASC").Limit(10).Find(&priceHistory)

    if result.Error != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Error al buscar el stock",
        })
        return
    }

    // Formatear datos para la gráfica
    var priceHistoryMapped = make([]map[string]interface{}, 0)
    for _, h := range priceHistory {
        priceHistoryMapped = append(priceHistoryMapped, map[string]interface{}{
            "time": h.Time,
            "target_to": roundToTwoDecimals(h.TargetTo),
            "target_from": roundToTwoDecimals(h.TargetFrom),
        })
    }

    priceChange := CalculatePriceChange(stock.TargetFrom, stock.TargetTo)


    stockMapped := map[string]interface{}{
        "id":           stock.ID,
        "ticker":       stock.Ticker,
        "company":      stock.Company,
        "action":       stock.Action,
        "brokerage":    stock.Brokerage,
        "rating_from":  stock.RatingFrom,
        "rating_to":    stock.RatingTo,
        "time":         stock.Time,
        "target_from":  roundToTwoDecimals(stock.TargetFrom),
        "target_to":    roundToTwoDecimals(stock.TargetTo),
        "price_history": priceHistoryMapped,
        "analysis":      priceChange,
    }

    c.JSON(http.StatusOK, stockMapped)
}

func GetStockPriceHistory(c *gin.Context) {
    ticker := c.Param("ticker")

    // Obtener el período de tiempo desde los query params (default último día)
    period := c.DefaultQuery("period", "1d")

    var startTime time.Time
    endTime := time.Now()

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

    var history []StockPriceHistory
    result := DB.Where("ticker = ? AND time BETWEEN ? AND ?",
        ticker, startTime, endTime).
        Order("time ASC").
        Find(&history)

    if result.Error != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener historial"})
        return
    }

    // Formatear datos para la gráfica
    var response []map[string]interface{}
    for _, h := range history {
        response = append(response, map[string]interface{}{
            "time": h.Time,
            "target_to": roundToTwoDecimals(h.TargetTo),
            "target_from": roundToTwoDecimals(h.TargetFrom),
        })
    }

    c.JSON(http.StatusOK, response)
}

func UpdateStocks(c *gin.Context) {
	_, err := GetStockAndStoreInDB()
	if err != nil {
		log.Println("Error actualizando base de datos manualmente:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al actualizar base de datos manualmente"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Base de datos actualizada correctamente"})
}

func GetRecommendedStock(c *gin.Context) {
	var stocks []Stock

    today := time.Now().Truncate(24 * time.Hour)
    log.Println("Fecha de hoy en UTC:", today.Format("2006-01-02"))

	DB.Where("rating_to = ? AND target_to > target_from AND (action LIKE ? OR action LIKE ?) AND DATE(time AT TIME ZONE 'UTC') = ?",
        "Buy", "%target raised by%", "%upgraded to%", today.Format("2006-01-02")).Find(&stocks)

	var stockResponses []map[string]interface{}

	for _, stock := range stocks {
		var priceHistory []StockPriceHistory
		DB.Where("ticker = ?", stock.Ticker).Order("time ASC").Limit(10).Find(&priceHistory)

		var priceHistoryMapped = make([]map[string]interface{}, 0)
		for _, h := range priceHistory {
			priceHistoryMapped = append(priceHistoryMapped, map[string]interface{}{
				"time":  h.Time,
				"target_to":  roundToTwoDecimals(h.TargetTo),
				"target_from": roundToTwoDecimals(h.TargetFrom),
			})
		}

		log.Println(stock.Ticker)
		log.Println(stock.TargetFrom, stock.TargetTo)
		priceChange := CalculatePriceChange(stock.TargetFrom, stock.TargetTo)
		log.Println(priceChange)
		stockResponse := map[string]interface{}{
			"id":            stock.ID,
			"ticker":        stock.Ticker,
			"company":       stock.Company,
			"action":        stock.Action,
			"brokerage":     stock.Brokerage,
			"rating_from":   stock.RatingFrom,
			"rating_to":     stock.RatingTo,
			"time":          stock.Time,
			"target_from":   roundToTwoDecimals(stock.TargetFrom),
			"target_to":     roundToTwoDecimals(stock.TargetTo),
			"price_history": priceHistoryMapped,
			"analysis":      priceChange,
		}

		stockResponses = append(stockResponses, stockResponse)
	}

	c.JSON(http.StatusOK, stockResponses)
}

func GetNoRecommendedStock(c *gin.Context) {
    var stocks []Stock

    today := time.Now().UTC().Truncate(24 * time.Hour)

    DB.Where("rating_to = ? AND target_to < target_from AND (action LIKE ? OR action LIKE ?) AND DATE(time AT TIME ZONE 'UTC') = ?", 
        "Sell", "%target lowered by%", "%downgraded to%", today.Format("2006-01-02")).Find(&stocks)

    var stockResponses []map[string]interface{}

    for _, stock := range stocks {
        var priceHistory []StockPriceHistory
        DB.Where("ticker = ?", stock.Ticker).Order("times ASC").Limit(10).Find(&priceHistory)

        var priceHistoryMapped = make([]map[string]interface{}, 0)
        for _, h := range priceHistory {
            priceHistoryMapped = append(priceHistoryMapped, map[string]interface{}{
                "time": h.Time,
                "target_to": roundToTwoDecimals(h.TargetTo),
                "target_from": roundToTwoDecimals(h.TargetFrom),
            })
        }

        stockResponse := map[string]interface{}{
            "id":           stock.ID,
            "ticker":       stock.Ticker,
            "company":      stock.Company,
            "action":       stock.Action,
            "brokerage":    stock.Brokerage,
            "rating_from":  stock.RatingFrom,
            "rating_to":    stock.RatingTo,
            "time":         stock.Time,
            "target_from":  roundToTwoDecimals(stock.TargetFrom),
            "target_to":    roundToTwoDecimals(stock.TargetTo),
            "price_history": priceHistoryMapped,
        }

        stockResponses = append(stockResponses, stockResponse)
    }

    c.JSON(http.StatusOK, stockResponses)
}

func SearchStocks(c *gin.Context) {
    query := c.Query("q")
    if query == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Parámetro de búsqueda requerido"})
        return
    }

    // Preparar el término de búsqueda para LIKE
    searchTerm := "%" + strings.ToLower(query) + "%"

    var stocks []Stock
    result := DB.Where(
        "LOWER(ticker) LIKE ? OR LOWER(company) LIKE ? OR LOWER(brokerage) LIKE ?",
        searchTerm, searchTerm, searchTerm,
    ).Find(&stocks)

    if result.Error != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error en la búsqueda"})
        return
    }

    // Si no hay resultados, devolver array vacío en lugar de null
    if len(stocks) == 0 {
        c.JSON(http.StatusOK, gin.H{"message": "No se encontraron stocks para la búsqueda"})
        return
    }

    var stockResponses []map[string]interface{}

    for _, stock := range stocks {
        var priceHistory []StockPriceHistory
        DB.Where("ticker = ?", stock.Ticker).Order("times ASC").Limit(10).Find(&priceHistory)

        var priceHistoryMapped = make([]map[string]interface{}, 0)
        for _, h := range priceHistory {
            priceHistoryMapped = append(priceHistoryMapped, map[string]interface{}{
                "time": h.Time,
                "target_to": roundToTwoDecimals(h.TargetTo),
                "target_from": roundToTwoDecimals(h.TargetFrom),
            })
        }

        stockResponse := map[string]interface{}{
            "id":           stock.ID,
            "ticker":       stock.Ticker,
            "company":      stock.Company,
            "action":       stock.Action,
            "brokerage":    stock.Brokerage,
            "rating_from":  stock.RatingFrom,
            "rating_to":    stock.RatingTo,
            "time":         stock.Time,
            "target_from":  roundToTwoDecimals(stock.TargetFrom),
            "target_to":    roundToTwoDecimals(stock.TargetTo),
            "price_history": priceHistoryMapped,
        }

        stockResponses = append(stockResponses, stockResponse)
    }

    c.JSON(http.StatusOK, stockResponses)
}


// Middlewares

func CORSMiddleware() gin.HandlerFunc {
    var corsConfig = cors.Config{
        AllowOrigins:     []string{"http://localhost:5173", "http://localhost:4173"},
        AllowMethods:     []string{"GET", "PUT", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
        ExposeHeaders:    []string{"Content-Length", "Content-Type"},
        AllowCredentials: true,
        MaxAge: 12 * time.Hour,
    }
    return cors.New(corsConfig)
}

func main() {
    err := godotenv.Load()
	if err != nil {
		log.Fatal("Error cargando el archivo .env")
	}

    allowedOrigins = getOrigins()

    log.Println("Allowed origins:", allowedOrigins)

	InitDB()

    // Conectar a Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	r := gin.Default()

    // Middlewares
    r.Use(CORSMiddleware())

    // Crear el handler de WebSocket (sin redis)
	// wsHandler := NewWebSocketHandler()
	// go wsHandler.handleMessages()

    // Crear el handler de WebSocket (sin redis)
    wsHandler := NewWebSocketHandler(redisClient)
	go wsHandler.subscribeStockUpdates()

	// Routes
	r.GET("/stocks", GetStocks)
	r.PUT("/stocks", UpdateStocks)
	r.GET("/stocks/:ticker", GetStockByTicker)
	r.GET("/stocks/:ticker/history", GetStockPriceHistory)
	r.GET("/stocks/recommendations", GetRecommendedStock)
	r.GET("/stocks/not-recommended", GetNoRecommendedStock)
	r.GET("/stocks/search", SearchStocks)

    // WebSocket routes
	r.GET("/stocks/ws", wsHandler.handleConnections)

	// Start massive stock simulation service
	// StartMassiveStockSimulationService(wsHandler)

	// Iniciar servidor
	port := os.Getenv("PORT")

	if port == "" {
		port = "3000"
	}

	r.Run(":" + port)
}
