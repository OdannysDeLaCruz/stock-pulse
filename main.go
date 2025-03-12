package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	// "math/rand"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/OdannysDeLaCruz/stock-tracker/migrations" // Ajusta según tu módulo

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
    "github.com/gin-contrib/cors"
)

var DB *gorm.DB

// Models

type Stock struct {
	gorm.Model
	Ticker      string `gorm:"uniqueIndex"`
	TargetFrom  string `gorm:"column:target_from;not null"`
	TargetTo    string `gorm:"column:target_to;not null"`
	Company     string `gorm:"not null"`
	Action      string `gorm:"not null"`
	Brokerage   string `gorm:"not null"`
	RatingFrom  string `gorm:"column:rating_from;not null"`
	RatingTo    string `gorm:"column:rating_to;not null"`
	Time        string `gorm:"not null"`
	ChangePct    float64
	CurrentPrice float64
}

type StockData struct {
	Ticker string `json:"ticker"`
	Target_from string `json:"target_from"`
	Target_to string `json:"target_to"`
	Company string `json:"company"`
	Action string `json:"action"`
	Brokerage string `json:"brokerage"`
	Rating_from string `json:"rating_from"`
	Rating_to string `json:"rating_to"`
	Time string `json:"time"`
}

type APIResponse struct {
    Items []StockData `json:"items"`
    NextPage string   `json:"next_page"`
}

// Estructura para mantener los datos simulados
type StockSimulator struct {
    stocks map[string]*Stock
    mu     sync.RWMutex
}

// Singleton para el simulador
var stockSimulator *StockSimulator

// Estructura para manejar las conexiones WebSocket
type WSHandler struct {
    clients    map[*websocket.Conn]bool
    broadcast  chan []Stock
    register   chan *websocket.Conn
    unregister chan *websocket.Conn
    mu         sync.Mutex
}

// Configuración del upgrader de WebSocket
var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool {
        return true // En producción, configura esto adecuadamente
    },
}

// Crear el manejador de WebSocket
func NewWSHandler() *WSHandler {
    return &WSHandler{
        clients:    make(map[*websocket.Conn]bool),
        broadcast:  make(chan []Stock),
        register:   make(chan *websocket.Conn),
        unregister: make(chan *websocket.Conn),
    }
}

// Nueva estructura para el historial de precios
type StockPriceHistory struct {
    gorm.Model
    StockID     uint      `gorm:"index"`
    Ticker      string    `gorm:"index"`
    Price       float64
    Timestamp   time.Time `gorm:"index"`
}

// Services
func InitDB() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error cargando el archivo .env")
	}

	dsn := os.Getenv("DATABASE_URL")
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("No se pudo conectar a CockroachDB:", err)
	}

	DB = database
	fmt.Println("Conectado a CockroachDB con GORM")
}

func initializeSimulator() {
    stockSimulator = &StockSimulator{
        stocks: make(map[string]*Stock),
    }

    // Cargar datos iniciales de la base de datos
    var dbStocks []Stock
    DB.Find(&dbStocks)

    for _, stock := range dbStocks {
        // Convertir el precio inicial desde TargetFrom
        basePrice, _ := strconv.ParseFloat(strings.Trim(strings.ReplaceAll(stock.TargetFrom, ",", ""), "$"), 64)
        stock.CurrentPrice = basePrice
        stockSimulator.stocks[stock.Ticker] = &stock
    }
}

// func fetchRealtimeStockData() ([]Stock, error) {
//     stockSimulator.mu.Lock()
//     defer stockSimulator.mu.Unlock()

//     for _, stock := range stockSimulator.stocks {
//         // Lógica existente de actualización de precios...
//         changePercent := (rand.Float64() * 4) - 2
//         stock.CurrentPrice = stock.CurrentPrice * (1 + changePercent/100)
        
//         // Guardar el historial de precios
//         priceHistory := StockPriceHistory{
//             StockID:   stock.ID,
//             Ticker:    stock.Ticker,
//             Price:     stock.CurrentPrice,
//             Timestamp: time.Now(),
//         }
        
//         if err := DB.Create(&priceHistory).Error; err != nil {
//             log.Printf("Error al guardar historial de precios: %v", err)
//         }
        
//         // Actualizar el tiempo
//         stock.Time = time.Now().Format(time.RFC3339)
        
//         // Actualizar los targets basados en el nuevo precio
//         newTarget := stock.CurrentPrice * (1 - (rand.Float64() * 10)/100)
//         stock.TargetFrom = fmt.Sprintf("$%.2f", stock.CurrentPrice)
//         stock.TargetTo = fmt.Sprintf("$%.2f", newTarget)
        
//         // Calcular y actualizar el ChangePct
//         targetFrom := stock.CurrentPrice
//         targetTo := newTarget
//         stock.ChangePct = ((targetTo - targetFrom) / targetFrom) * 100
        
//         // Actualizar el rating ocasionalmente
//         if rand.Float64() < 0.1 { // 10% de probabilidad de cambio
//             ratings := []string{"Buy", "Hold", "Sell", "Strong-Buy", "Strong-Sell", "Neutral", "Overweight", "Underweight", "Outperform", "Underperform"}
//             stock.RatingTo = ratings[rand.Intn(len(ratings))]
//         }
//     }

//     // Convertir el mapa a slice para retornar
//     result := make([]Stock, 0, len(stockSimulator.stocks))
//     for _, stock := range stockSimulator.stocks {
//         result = append(result, *stock)
//     }

//     return result, nil
// }

// func startStockUpdateService(wsHandler *WSHandler) {
//     if stockSimulator == nil {
//         initializeSimulator()
//     }

//     ticker := time.NewTicker(5 * time.Second)
//     go func() {
//         for {
//             select {
//             case <-ticker.C:
//                 stocks, err := fetchRealtimeStockData()
//                 if err != nil {
//                     log.Printf("Error actualizando stocks: %v", err)
//                     continue
//                 }
                
//                 // Actualizar la base de datos
//                 for _, stock := range stocks {
//                     DB.Save(&stock)
//                 }
                
//                 // Enviar actualización a todos los clientes conectados
//                 wsHandler.broadcast <- stocks
//             }
//         }
//     }()
// }

// Obtener datos de la API externa
func FetchStockData() ([]StockData, error) {
	baseURL := "https://8j5baasof2.execute-api.us-west-2.amazonaws.com/production/swechallenge/list"
	var allStocks []StockData
	nextPage := ""
	maxStocks := 40

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

		var response APIResponse
		err = json.Unmarshal(body, &response)
		if err != nil {
			return nil, fmt.Errorf("error al parsear JSON: %v", err)
		}

		// Agregar los items a la lista total
		allStocks = append(allStocks, response.Items...)

		// Si no hay siguiente página o ya tenemos suficientes stocks, terminamos
		if response.NextPage == "" || len(allStocks) >= maxStocks {
			break
		}

		nextPage = response.NextPage
	}

	// Si tenemos más de 40 stocks, truncamos la lista
	if len(allStocks) > maxStocks {
		allStocks = allStocks[:maxStocks]
	}

	return allStocks, nil
}

// Guardar datos de la API externa en la base de datos
func SaveStockData(stocks []StockData) error {
	log.Println(stocks)
	for _, stockData := range stocks {
		existingStock := Stock{}
		result := DB.Where("ticker = ?", stockData.Ticker).First(&existingStock)

		if result.RowsAffected == 0 {
			// Insertar si no existe
			newStock := Stock{
				Ticker:     stockData.Ticker,
				TargetFrom: stockData.Target_from,
				TargetTo:   stockData.Target_to,
				Company:    stockData.Company,
				Action:     stockData.Action,
				Brokerage:  stockData.Brokerage,
				RatingFrom: stockData.Rating_from,
				RatingTo:   stockData.Rating_to,
				Time:       stockData.Time,
			}
			if err := DB.Create(&newStock).Error; err != nil {
				return err
			}
		} else {
			// Actualizar si ya existe
			existingStock.TargetFrom = stockData.Target_from
			existingStock.TargetTo = stockData.Target_to
			existingStock.Company = stockData.Company
			existingStock.Action = stockData.Action
			existingStock.Brokerage = stockData.Brokerage
			existingStock.RatingFrom = stockData.Rating_from
			existingStock.RatingTo = stockData.Rating_to
			existingStock.Time = stockData.Time

			if err := DB.Save(&existingStock).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

// Handlers para las rutas

func GetStocks(c *gin.Context) {
	var stocks []Stock
	DB.Find(&stocks)
	c.JSON(http.StatusOK, stocks)
}

func GetStockAndStoreInDB() ([]StockData, error) {
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

func UpdateStocks(c *gin.Context) {
	_, err := GetStockAndStoreInDB()
	if err != nil {
		log.Println("Error actualizando base de datos manualmente:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al actualizar base de datos manualmente"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Base de datos actualizada correctamente"})
}

func recommendStocks(c *gin.Context) {
	var stocks []Stock
	DB.Where("rating_to = ?", "Buy").Find(&stocks)

	validStocks := []Stock{}
	now := time.Now()
	cutoffTime := now.AddDate(0, -1, 0)

	for _, stock := range stocks {
		stockTime, err := time.Parse(time.RFC3339, stock.Time)
		if err == nil && stockTime.After(cutoffTime) {
			// Convertir los valores de target eliminando el "$" y "," para calcular el potencial
			targetFromStr := strings.Trim(strings.ReplaceAll(stock.TargetFrom, ",", ""), "$")
			targetToStr := strings.Trim(strings.ReplaceAll(stock.TargetTo, ",", ""), "$")

			targetFrom, _ := strconv.ParseFloat(targetFromStr, 64)
			targetTo, _ := strconv.ParseFloat(targetToStr, 64)

			// Calcular el porcentaje de cambio potencial
			if targetFrom > 0 {
				stock.ChangePct = ((targetTo - targetFrom) / targetFrom) * 100
			}

			validStocks = append(validStocks, stock)
		}
	}

	// Ordenar por potencial de crecimiento (ChangePct) de mayor a menor
	sort.Slice(validStocks, func(i, j int) bool {
		return validStocks[i].ChangePct > validStocks[j].ChangePct
	})

	// Limitar a los 5 primeros stocks
	if len(validStocks) > 5 {
		validStocks = validStocks[:5]
	}

	c.JSON(http.StatusOK, validStocks)
}

func notRecommendedStocks(c *gin.Context) {
    var stocks []Stock
    DB.Where("rating_to = ?", "Sell").Find(&stocks)

    validStocks := []Stock{}
    now := time.Now()
    cutoffTime := now.AddDate(0, -1, 0) // Un mes atrás

    for _, stock := range stocks {
        stockTime, err := time.Parse(time.RFC3339, stock.Time)
        if err == nil && stockTime.After(cutoffTime) {
            // Convertir los valores de target eliminando el "$" y "," para calcular el potencial
            targetFromStr := strings.Trim(strings.ReplaceAll(stock.TargetFrom, ",", ""), "$")
            targetToStr := strings.Trim(strings.ReplaceAll(stock.TargetTo, ",", ""), "$")

            targetFrom, _ := strconv.ParseFloat(targetFromStr, 64)
            targetTo, _ := strconv.ParseFloat(targetToStr, 64)

            // Calcular el porcentaje de cambio potencial (negativo en este caso)
            if targetFrom > 0 {
                stock.ChangePct = ((targetTo - targetFrom) / targetFrom) * 100
            }

            validStocks = append(validStocks, stock)
        }
    }

    // Ordenar por potencial de pérdida (ChangePct) de menor a mayor
    sort.Slice(validStocks, func(i, j int) bool {
        return validStocks[i].ChangePct < validStocks[j].ChangePct
    })

    // Limitar a los 5 stocks con mayor potencial de pérdida
    if len(validStocks) > 5 {
        validStocks = validStocks[:5]
    }

    c.JSON(http.StatusOK, validStocks)
}

// Método para manejar las conexiones WebSocket
func (h *WSHandler) handleConnections(c *gin.Context) {
    // Actualizar la conexión HTTP a WebSocket
    ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        log.Printf("Error al actualizar a websocket: %v", err)
        return
    }

    // Registrar nuevo cliente
    h.register <- ws

    // Limpiar la conexión cuando la función termine
    defer func() {
        h.unregister <- ws
        ws.Close()
    }()

    // Mantener la conexión viva
    for {
        // Leer mensajes del cliente (opcional)
        _, _, err := ws.ReadMessage()
        if err != nil {
            break
        }
    }
}

// Método para ejecutar el bucle principal del WebSocket
func (h *WSHandler) run() {
    for {
        select {
        case client := <-h.register:
            h.mu.Lock()
            h.clients[client] = true
            h.mu.Unlock()

        case client := <-h.unregister:
            h.mu.Lock()
            if _, ok := h.clients[client]; ok {
                delete(h.clients, client)
                client.Close()
            }
            h.mu.Unlock()

        case stocks := <-h.broadcast:
            h.mu.Lock()
            for client := range h.clients {
                err := client.WriteJSON(stocks)
                if err != nil {
                    log.Printf("Error: %v", err)
                    client.Close()
                    delete(h.clients, client)
                }
            }
            h.mu.Unlock()
        }
    }
}

// Handler para obtener un stock específico
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

    c.JSON(http.StatusOK, stock)
}

// Nuevo endpoint para obtener el historial de precios
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
    result := DB.Where("ticker = ? AND timestamp BETWEEN ? AND ?", 
        ticker, startTime, endTime).
        Order("timestamp ASC").
        Find(&history)

    if result.Error != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener historial"})
        return
    }

    // Formatear datos para la gráfica
    var response []map[string]interface{}
    for _, h := range history {
        response = append(response, map[string]interface{}{
            "timestamp": h.Timestamp.Unix(),
            "price":     h.Price,
        })
    }

    c.JSON(http.StatusOK, response)
}

// Agregar esta función con los handlers
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

    c.JSON(http.StatusOK, stocks)
}

// Middlewares

func CORSMiddleware() gin.HandlerFunc {
    var corsConfig = cors.Config{
        AllowOrigins:     []string{"http://localhost:5173"},
        AllowMethods:     []string{"GET", "PUT", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
        ExposeHeaders:    []string{"Content-Length", "Content-Type"},
        AllowCredentials: true,
        MaxAge: 12 * time.Hour,
    }
    return cors.New(corsConfig)
}

func main() {
	InitDB()

	// Solo ejecutar migraciones si se especifica
	if os.Getenv("RUN_MIGRATIONS") == "true" {
		if err := migrations.RunMigrations(DB); err != nil {
			log.Fatal("Error en las migraciones:", err)
		}

		// Cargar datos iniciales solo si se resetea la DB
		if os.Getenv("RESET_DB") == "true" {
			_, err := GetStockAndStoreInDB()
			if err != nil {
				log.Fatal("Error cargando datos iniciales:", err)
			}
		}
	}

	r := gin.Default()

    r.Use(CORSMiddleware())

	// Crear y configurar WebSocket handler
	wsHandler := NewWSHandler()
	go wsHandler.run()

	// Rutas existentes
	r.GET("/stocks", GetStocks)
	r.GET("/stocks/:ticker", GetStockByTicker)
	r.PUT("/stocks", UpdateStocks)
	r.GET("/recommendations", recommendStocks)
	r.GET("/not-recommended", notRecommendedStocks)
	r.GET("/ws", wsHandler.handleConnections)
	r.GET("/stocks/:ticker/history", GetStockPriceHistory)
	r.GET("/stocks/search", SearchStocks)

	// Iniciar el servicio de actualización con WebSocket
	// startStockUpdateService(wsHandler)

	// Iniciar servidor
	port := os.Getenv("PORT")

	if port == "" {
		port = "3000"
	}

	r.Run(":" + port)
}
