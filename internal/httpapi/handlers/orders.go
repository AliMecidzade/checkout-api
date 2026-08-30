package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"checkout-api/internal/domain"
	"checkout-api/internal/httpapi"
	"checkout-api/internal/httpapi/dto"
	"checkout-api/internal/httpapi/middleware"
	"checkout-api/internal/service"
)

type OrdersHandler struct {
	orders *service.OrderService
}

func NewOrdersHandler(orders *service.OrderService) *OrdersHandler {
	return &OrdersHandler{orders: orders}
}

func (h *OrdersHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserID(r.Context())
	if !ok {
		httpapi.Error(w, service.ErrInvalidCredentials)
		return
	}

	key := r.Header.Get("Idempotency-Key")
	if key == "" {
		httpapi.BadRequest(w, "Idempotency-Key header is required")
		return
	}

	var req dto.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpapi.BadRequest(w, "invalid request body")
		return
	}

	items := make([]domain.LineItem, 0, len(req.LineItems))
	for _, li := range req.LineItems {
		items = append(items, domain.LineItem{ItemID: li.ItemID, Quantity: li.Quantity, Price: li.Price})
	}

	res, err := h.orders.CreateOrder(r.Context(), int64(userID), items, key)
	if err != nil {
		httpapi.Error(w, err)
		return
	}

	status := http.StatusCreated
	if !res.Payment.Success {
		status = http.StatusPaymentRequired
	}

	httpapi.JSON(w, status, dto.CreateOrderResponse{
		Order:   toOrderResponse(res.Order),
		Payment: dto.PaymentResult{Success: res.Payment.Success, TransactionID: res.Payment.TransactionID, Error: res.Payment.Error},
	})
}

func (h *OrdersHandler) GetUserOrders(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserID(r.Context())
	if !ok {
		httpapi.Error(w, service.ErrInvalidCredentials)
		return
	}

	page, ok := parsePage(r)
	if !ok {
		httpapi.BadRequest(w, "limit, offset and cursor must be valid")
		return
	}

	orders, hasMore, err := h.orders.ListOrdersByUser(r.Context(), int64(userID), page)
	if err != nil {
		httpapi.Error(w, err)
		return
	}

	out := make([]dto.OrderResponse, 0, len(orders))
	for _, o := range orders {
		out = append(out, toOrderResponse(o))
	}

	pagination := dto.Pagination{Limit: page.Limit, Offset: page.Offset, HasMore: hasMore}
	if hasMore && len(orders) > 0 {
		pagination.NextCursor = strconv.FormatInt(orders[len(orders)-1].ID, 10)
	}

	httpapi.JSON(w, http.StatusOK, dto.OrdersListResponse{Orders: out, Pagination: pagination})
}

func toOrderResponse(order domain.Order) dto.OrderResponse {
	items := make([]dto.LineItemResponse, 0, len(order.Items))
	for _, li := range order.Items {
		items = append(items, dto.LineItemResponse{
			ItemID:   li.ItemID,
			Quantity: li.Quantity,
			Price:    li.Price,
		})
	}

	return dto.OrderResponse{
		ID:        order.ID,
		UserID:    order.UserID,
		Items:     items,
		Total:     order.Total,
		Status:    order.Status,
		CreatedAt: order.CreatedAt,
	}
}
