package delivery

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/OdannysDeLaCruz/stock-tracker/config"
	"github.com/OdannysDeLaCruz/stock-tracker/internal/domain/stock"
	"github.com/OdannysDeLaCruz/stock-tracker/internal/domain/price_history"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

// WebSocketHandler maneja conexiones WebSocket
type WebSocketHandler struct {
	clients     map[*websocket.Conn]string
	broadcast   chan stock.Stock
	register    chan *websocket.Conn
	unregister  chan *websocket.Conn
	mutex       sync.Mutex
	redisClient *redis.Client
	stockService stock.StockService
	StockPriceHistoryService price_history.StockPriceHistoryService
	allowedOrigins map[string]bool
}

// NewWebSocketHandler crea un nuevo handler WS
func NewWebSocketHandler(config *config.Config, router *gin.Engine, redisClient *redis.Client, stockService stock.StockService, StockPriceHistoryService price_history.StockPriceHistoryService) *WebSocketHandler {
	handler := &WebSocketHandler{
		clients:      make(map[*websocket.Conn]string),
		broadcast:    make(chan stock.Stock),
		register:     make(chan *websocket.Conn),
		unregister:   make(chan *websocket.Conn),
		redisClient:  redisClient,
		stockService: stockService,
		StockPriceHistoryService: StockPriceHistoryService,
		allowedOrigins: config.AllowedOrigins,
	}

	handler.setupUpgrader()

	router.GET("/stocks/ws", handler.handleConnections)
	go handler.subscribeStockUpdates()

	return handler
}

// Configuración del upgrader de WebSocket
var upgrader websocket.Upgrader

func (w *WebSocketHandler) setupUpgrader() {
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			return w.allowedOrigins[origin]
		},
	}
}

// Maneja conexiones WebSocket
func (w *WebSocketHandler) handleConnections(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Error al actualizar la conexión WebSocket:", err)
		return
	}
	defer conn.Close()

	w.mutex.Lock()
	w.clients[conn] = c.Query("ticker") // Suscribir a un ticker específico
	w.mutex.Unlock()

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

// Suscribe a eventos de Redis y los envía a clientes WS
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

		var stock stock.StockPartial
		if err := json.Unmarshal([]byte(msg.Payload), &stock); err != nil {
			log.Println("Error al parsear stock:", err)
			continue
		}

		// Guardar en base de datos antes de enviar a clientes WS
		err = w.saveStockUpdate(stock)
		if err != nil {
			log.Println("Error al guardar stock en la base de datos:", err)
			continue
		}

		// Enviar solo a los clientes suscritos al ticker del stock
		w.mutex.Lock()
		for client, subTicker := range w.clients {
			if subTicker == stock.Ticker {
				client.WriteJSON(stock)
			}
		}
		w.mutex.Unlock()
	}
}

func (w *WebSocketHandler) saveStockUpdate(stockPartial stock.StockPartial) error {

    existingStock, err := w.stockService.GetStockByTickerRaw(stockPartial.Ticker)
	if err != nil {
		return err
	}

	if existingStock == nil {
		log.Println("No se encontró ningún registro con ese ticker")
	} else {
		// Actualizar el registro existente
        existingStock.TargetFrom  = stockPartial.TargetFrom
        existingStock.TargetTo    = stockPartial.TargetTo
        existingStock.RatingFrom  = stockPartial.RatingFrom
        existingStock.RatingTo    = stockPartial.RatingTo
        existingStock.Time        = stockPartial.NewItemStockPriceHistory.Time.Truncate(time.Microsecond)


        err = w.stockService.Save(existingStock)
		if err != nil {
			log.Println("Error actualizando stock:", err)
			return err
		}
		log.Println("✅ Stock actualizado:", stockPartial.Ticker)

		// Guardar historial de precios
		StockPriceHistory := price_history.StockPriceHistory{
			StockID:     existingStock.ID,
			Ticker:      existingStock.Ticker,
			TargetFrom:  stockPartial.TargetFrom,
			TargetTo:    stockPartial.TargetTo,
			Time:        stockPartial.NewItemStockPriceHistory.Time.Truncate(time.Microsecond),
		}

		err = w.StockPriceHistoryService.Create(&StockPriceHistory)
		if err != nil {
			log.Println("Error creando historial", err)
		}
	}

    return nil
}
