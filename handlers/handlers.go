package handlers

import (
	"checkout-api/models"
	"cmp"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"
)

// Implement HTTP handlers:
// - NewHandler(s *store.Store) — constructor
// - GetItems: GET /items → return all items as JSON
// - CreateOrder: POST /orders → decode request, validate items, calculate total, mock payment, create order

type ItemStore interface {
	GetItems() []*models.Item
	GetItem(id int) *models.Item
	CreateOrder(userID int, items []models.LineItem, total int, status string) *models.Order

	GetOrdersByUserId(userID int) []*models.Order
}

type Handler struct {
	itemStore ItemStore
}

type CreateOrderRequest struct {
	UserID int               `json:"user_id"`
	Items  []models.LineItem `json:"items"`
}

func NewHandler(itemStore ItemStore) *Handler {
	return &Handler{
		itemStore: itemStore,
	}
}

func (handler *Handler) GetItem(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		http.Error(writer, http.StatusText(http.StatusMethodNotAllowed),
			http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimPrefix(request.URL.Path, "/items/")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(writer, "invalid item id", http.StatusBadRequest)
		return
	}

	item := handler.itemStore.GetItem(id)

	if item == nil {
		http.Error(writer, "Item Not Found", http.StatusNotFound)
		return
	}

	writer.Header().Set("Cache-Control", "max-age=60")
	writeJson(writer, http.StatusOK, item)

}
func (handler *Handler) GetItems(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		http.Error(writer, http.StatusText(http.StatusMethodNotAllowed),
			http.StatusMethodNotAllowed)
		return
	}

	items := handler.itemStore.GetItems()

	queries := request.URL.Query()

	category := queries.Get("category")
	sortBy := queries.Get("sort")

	fmt.Printf("%v\n", queries)
	fmt.Printf("%s\n", category)
	fmt.Printf("%s\n", sortBy)

	switch sortBy {

	case "":

	case "price":
		slices.SortFunc(items, func(a, b *models.Item) int {
			return cmp.Compare(a.Price, b.Price)
		})

	case "name":
		slices.SortFunc(items, func(a, b *models.Item) int {
			return strings.Compare(a.Name, b.Name)
		})

	case "stock":
		slices.SortFunc(items, func(a, b *models.Item) int {
			return cmp.Compare(a.StockSize, b.StockSize)
		})

	default:
		http.Error(writer, "invalid sort field", http.StatusBadRequest)
		return

	}

	writer.Header().Set("Vary", "Accept-Encoding")
	writer.Header().Set("Cache-Control", "max-age=60")
	writeJson(writer, http.StatusOK, items)
}

func (handler *Handler) CreateOrder(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		http.Error(writer, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	var req CreateOrderRequest
	total := 0
	retryCount := 4
	err := json.NewDecoder(request.Body).Decode(&req)

	if err != nil {
		http.Error(writer, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.UserID <= 0 {
		http.Error(writer, "invalid user id", http.StatusBadRequest)
		return
	}
	if len(req.Items) == 0 {
		http.Error(writer, "There are no items", http.StatusBadRequest)
		return

	}

	for i := range req.Items {
		if req.Items[i].ItemID <= 0 {
			http.Error(writer, "invalid item id", http.StatusBadRequest)
			return
		}

		if req.Items[i].Quantity <= 0 {
			http.Error(writer, "invalid quantity", http.StatusBadRequest)
			return
		}

		item := handler.itemStore.GetItem(req.Items[i].ItemID)
		if item == nil {
			http.Error(writer, "Item not found", http.StatusBadRequest)
			return
		}

		req.Items[i].Price = item.Price

		total += req.Items[i].Price * req.Items[i].Quantity
	}

	waitTime := 100
	var paymentErr error

	for i := 0; i < retryCount; i++ {
		paymentErr = MockPayment()

		if paymentErr == nil {
			fmt.Println("Payment successful ")
			break
		}

		fmt.Printf("Attempt %d: payment failed\n", i+1)

		if i == retryCount-1 {
			http.Error(writer, paymentErr.Error(), http.StatusInternalServerError)
			return
		}

		time.Sleep(time.Duration(waitTime) * time.Millisecond)
		waitTime *= 2
	}

	order := handler.itemStore.CreateOrder(req.UserID, req.Items, total, "paid")
	writeJson(writer, http.StatusCreated, order)

}

func (handler *Handler) GetOrdersByUserId(writer http.ResponseWriter, request *http.Request) {

	if request.Method != http.MethodGet {
		http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimPrefix(request.URL.Path, "/users/")
	idStr = strings.TrimSuffix(idStr, "/orders")
	userId, err := strconv.Atoi(idStr)

	if err != nil {
		http.Error(writer, "invalid user id", http.StatusBadRequest)
		return
	}

	orders := handler.itemStore.GetOrdersByUserId(userId)

	//if len(orders) == 0 {
	//
	//	return
	//}
	//
	writeJson(writer, http.StatusOK, orders)

}

func MockPayment() error {
	n := rand.Intn(10) + 1
	if n <= 4 {
		return nil
	}

	return fmt.Errorf("payment failed")
}

func writeJson(writer http.ResponseWriter, statusCode int, data any) {
	writer.Header().Add("Content-Type", "application/json")
	writer.WriteHeader(statusCode)
	json.NewEncoder(writer).Encode(data)
}
