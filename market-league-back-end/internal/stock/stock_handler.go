package stock

import (
	"encoding/json"
	"fmt"
	"time"

	ws "github.com/market-league/api/websocket"
	"github.com/market-league/internal/models"
)

// StockHandler Interface
type StockHandlerInterface interface {
	CreateStock(conn *ws.Connection, rawData json.RawMessage) error
	CreateMultipleStocks(conn *ws.Connection, rawData json.RawMessage) error
	UpdatePrice(conn *ws.Connection, rawData json.RawMessage) error
	GetAllStocks(conn *ws.Connection, rawData json.RawMessage) error
	GetStockInfo(conn *ws.Connection, rawData json.RawMessage) error
}

// Compile-time check
var _ StockHandlerInterface = (*StockHandler)(nil)

// StockHandler defines the HTTP handler for stock-related operations.
type StockHandler struct {
	StockService *StockService
}

// NewStockHandler creates a new instance of StockHandler.
func NewStockHandler(service *StockService) *StockHandler {
	return &StockHandler{StockService: service}
}

// * Implementation of Interface

// CreateStock handles the creation of a new stock
func (h *StockHandler) CreateStock(conn *ws.Connection, rawData json.RawMessage) error {
	// Step 1: Parse the WebSocket message
	var request struct {
		TickerSymbol string `json:"ticker_symbol" binding:"required"`
		CompanyName  string `json:"company_name" binding:"required"`
	}

	// Step 2: Parse data from WebSocket JSON payload
	if err := json.Unmarshal(rawData, &request); err != nil {
		ws.SendError(conn, ws.MessageType_Stock_CreateStock, "Invalid input: "+err.Error())
		return fmt.Errorf("invalid input: %v", err)
	}

	// Step 3: Process business logic (reuse the service layer)
	stock, err := h.StockService.CreateStock(request.TickerSymbol, request.CompanyName)
	if err != nil {
		ws.SendError(conn, ws.MessageType_Stock_CreateStock, err.Error())
		return fmt.Errorf("failed to create portfolio: %v", err)
	}

	// Step 4: Marshal the portfolio into JSON
	stockJSON, err := json.Marshal(stock)
	if err != nil {
		ws.SendError(conn, ws.MessageType_Stock_CreateStock, "Failed to serialize portfolio")
		return fmt.Errorf("serialization error: %v", err)
	}

	// Step 5: Send success response back via WebSocket
	response := ws.WebsocketMessage{
		Type: ws.MessageType_Stock_CreateStock,
		Data: json.RawMessage(stockJSON), // Use marshaled JSON bytes
	}
	if err := conn.Ws.WriteJSON(response); err != nil {
		return fmt.Errorf("failed to send response: %v", err)
	}

	return nil
}

// CreateMultipleStocks handles the creation of multiple stocks
func (h *StockHandler) CreateMultipleStocks(conn *ws.Connection, rawData json.RawMessage) error {
	// Step 1: Parse the WebSocket message
	var request []struct {
		TickerSymbol string `json:"ticker_symbol" binding:"required"`
		CompanyName  string `json:"company_name" binding:"required"`
	}

	// Step 2: Parse data from WebSocket JSON payload
	if err := json.Unmarshal(rawData, &request); err != nil {
		ws.SendError(conn, ws.MessageType_Stock_CreateMultipleStocks, "Invalid input: "+err.Error())
		return fmt.Errorf("invalid input: %v", err)
	}

	// Step 3: Convert requests to models.Stock
	var stocksPointer []*models.Stock
	for _, stockReq := range request {
		stock := &models.Stock{
			TickerSymbol: stockReq.TickerSymbol,
			CompanyName:  stockReq.CompanyName,
		}
		stocksPointer = append(stocksPointer, stock)
	}

	// Step 4: Process business logic (reuse the service layer)
	err := h.StockService.CreateMultipleStocks(stocksPointer)
	if err != nil {
		ws.SendError(conn, ws.MessageType_Stock_CreateMultipleStocks, err.Error())
		return fmt.Errorf("failed to create portfolio: %v", err)
	}

	// Step 5: Pointer to model conversion
	var stocks []models.Stock = extractStocksData(stocksPointer)

	// Step 6: Marshal the portfolio into JSON
	stockJSON, err := json.Marshal(stocks)
	if err != nil {
		ws.SendError(conn, ws.MessageType_Stock_CreateMultipleStocks, "Failed to serialize portfolio")
		return fmt.Errorf("serialization error: %v", err)
	}

	// Step 7: Send success response back via WebSocket
	response := ws.WebsocketMessage{
		Type: ws.MessageType_Stock_CreateMultipleStocks,
		Data: json.RawMessage(stockJSON), // Use marshaled JSON bytes
	}
	if err := conn.Ws.WriteJSON(response); err != nil {
		return fmt.Errorf("failed to send response: %v", err)
	}

	return nil
}

// UpdatePrice handles the request to update a stock's current price
func (h *StockHandler) UpdatePrice(conn *ws.Connection, rawData json.RawMessage) error {
	// Step 1: Parse the WebSocket message
	var request struct {
		StockID   uint       `json:"stock_id" binding:"required"`
		NewPrice  float64    `json:"new_price" binding:"required"`
		Timestamp *time.Time `json:"timestamp,omitempty"`
	}

	// Step 2: Parse data from WebSocket JSON payload
	if err := json.Unmarshal(rawData, &request); err != nil {
		ws.SendError(conn, ws.MessageType_Stock_UpdateCurrentStockPrice, "Invalid input: "+err.Error())
		return fmt.Errorf("invalid input: %v", err)
	}

	// Step 3: Process business logic (reuse the service layer)
	if err := h.StockService.UpdateStockPrice(request.StockID, request.NewPrice, request.Timestamp); err != nil {
		ws.SendError(conn, ws.MessageType_Stock_UpdateCurrentStockPrice, err.Error())
		return fmt.Errorf("failed to create portfolio: %v", err)
	}

	// Step 4: Send success response (no data, just confirmation)
	response := ws.WebsocketMessage{
		Type: ws.MessageType_Stock_UpdateCurrentStockPrice,
		Data: json.RawMessage(`{"message": "Stock price updated successfully"}`), // Simple JSON message
	}
	if err := conn.Ws.WriteJSON(response); err != nil {
		return fmt.Errorf("failed to send response: %v", err)
	}

	return nil
}

func (h *StockHandler) GetAllStocks(conn *ws.Connection, rawData json.RawMessage) error {
	// Step 1: Fetch all stocks using the service layer
	stocks, err := h.StockService.GetAllStocks()
	if err != nil {
		// Send an error response if something goes wrong
		ws.SendError(conn, ws.MessageType_Stock_GetAllStocks, "Failed to retrieve stocks: "+err.Error())
		return fmt.Errorf("failed to retrieve stocks: %v", err)
	}

	// Step 2: Prepare the response
	responseData, err := json.Marshal(stocks)
	if err != nil {
		ws.SendError(conn, ws.MessageType_Stock_GetAllStocks, "Failed to serialize stock data: "+err.Error())
		return fmt.Errorf("failed to serialize stock data: %v", err)
	}

	// Step 3: Send the list of stocks back to the client
	response := ws.WebsocketMessage{
		Type: ws.MessageType_Stock_GetAllStocks,
		Data: json.RawMessage(responseData),
	}
	if err := conn.Ws.WriteJSON(response); err != nil {
		return fmt.Errorf("failed to send response: %v", err)
	}

	return nil
}

// GetStockInfo handles retrieving stock information with stock price history
func (h *StockHandler) GetStockInfo(conn *ws.Connection, rawData json.RawMessage) error {
	// Step 1: Parse the WebSocket message
	var request struct {
		StockID uint `json:"stock_id" binding:"required"`
	}

	// Step 2: Parse data from WebSocket JSON payload
	if err := json.Unmarshal(rawData, &request); err != nil {
		ws.SendError(conn, ws.MessageType_Stock_GetStockInformation, "Invalid input: "+err.Error())
		return fmt.Errorf("invalid input: %v", err)
	}

	// Step 3: Process business logic (reuse the service layer)
	stock, err := h.StockService.GetStockInfo(request.StockID)
	if err != nil {
		ws.SendError(conn, ws.MessageType_Stock_GetStockInformation, err.Error())
		return fmt.Errorf("failed to create portfolio: %v", err)
	}

	// Step 4: Convert Stock to StockInfo
	stockInfo := models.StockInfo{
		ID:             stock.ID,
		TickerSymbol:   stock.TickerSymbol,
		CompanyName:    stock.CompanyName,
		CurrentPrice:   stock.CurrentPrice,
		PriceHistories: convertPriceHistories(stock.PriceHistories),
	}

	// Step 5: Map models.Stock to StockInfo DTO
	stockJSON, err := json.Marshal(stockInfo)
	if err != nil {
		ws.SendError(conn, ws.MessageType_Stock_GetStockInformation, "Failed to serialize portfolio")
		return fmt.Errorf("serialization error: %v", err)
	}

	// Step 6: Send success response back via WebSocket
	response := ws.WebsocketMessage{
		Type: ws.MessageType_Stock_GetStockInformation,
		Data: json.RawMessage(stockJSON), // Use marshaled JSON bytes
	}
	if err := conn.Ws.WriteJSON(response); err != nil {
		return fmt.Errorf("failed to send response: %v", err)
	}

	return nil
}

// * Helper functions

// Helper function to detect unique constraint errors
func isUniqueConstraintError(err error, field string) bool {
	// TODO: This function needs to be implemented based on database driver and error handling
	// For PostgreSQL with lib/pq, we can check for pq.Error and the specific constraint name
	// Generic placeholder:
	return false
}

// Helper function to extract necessary data from stocks
func extractStocksData(stocks []*models.Stock) []models.Stock {
	var result []models.Stock
	for _, stock := range stocks {
		result = append(result, *stock)
	}
	return result
}

func convertPriceHistories(histories []models.PriceHistory) []models.PriceHistoryDTO {
	// Create a slice with the same length as the input
	dto := make([]models.PriceHistoryDTO, len(histories))
	for i, history := range histories {
		dto[i] = models.PriceHistoryDTO{
			ID:        history.ID,
			Price:     history.Price,
			Timestamp: history.Timestamp,
		}
	}
	return dto
}
