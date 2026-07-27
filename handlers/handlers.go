package handlers

import (
	"checkout-api/models"
	"encoding/json"
	"net/http"
)

// Implement HTTP handlers:
// - NewHandler(s *store.Store) — constructor
// - GetItems: GET /items → return all items as JSON
// - CreateOrder: POST /orders → decode request, validate items, calculate total, mock payment, create order

type ItemStore interface {
	GetItems() []*models.Item
	GetItem(id int) *models.Item
	CreateOrder(userID int, items []models.LineItem, total int, status string) *models.Order
}

type Handler struct {
	itemStore ItemStore
}

func NewHandler(itemStore ItemStore) *Handler {
	return &Handler{
		itemStore: itemStore,
	}
}
func (handler *Handler) GetItems(writer http.ResponseWriter, request *http.Request) {
	if request.Method != "GET" {
		http.Error(writer, http.StatusText(http.StatusMethodNotAllowed),
			http.StatusMethodNotAllowed)
		return
	}
	items := handler.itemStore.GetItems()
	writeJson(writer, http.StatusOK, items)
}

func writeJson(writer http.ResponseWriter, statusCode int, data any) {
	writer.Header().Add("Content-Type", "application/json")
	writer.WriteHeader(statusCode)
	json.NewEncoder(writer).Encode(data)
}
