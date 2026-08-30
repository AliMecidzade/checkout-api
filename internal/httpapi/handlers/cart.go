package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"checkout-api/internal/httpapi"
	"checkout-api/internal/httpapi/dto"
	"checkout-api/internal/httpapi/middleware"
	"checkout-api/internal/service"
)

type CartHandler struct {
	cart *service.CartService
}

func NewCartHandler(cart *service.CartService) *CartHandler {
	return &CartHandler{cart: cart}
}

func (h *CartHandler) GetCart(w http.ResponseWriter, r *http.Request) {
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

	items, hasMore, err := h.cart.List(r.Context(), int64(userID), page)
	if err != nil {
		httpapi.Error(w, err)
		return
	}

	out := make([]dto.CartItemResponse, 0, len(items))
	for _, it := range items {
		out = append(out, dto.CartItemResponse{
			ID: it.ID, Name: it.Name, Description: it.Description,
			Price: it.Price, Stock: it.Stock, Quantity: it.Quantity,
			CreatedAt: it.CreatedAt,
		})
	}

	pagination := dto.Pagination{Limit: page.Limit, Offset: page.Offset, HasMore: hasMore}
	if hasMore && len(items) > 0 {
		pagination.NextCursor = strconv.FormatInt(items[len(items)-1].ID, 10)
	}

	httpapi.JSON(w, http.StatusOK, dto.CartListResponse{UserID: int64(userID), Items: out, Pagination: pagination})
}

func (h *CartHandler) UpsertItem(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserID(r.Context())
	if !ok {
		httpapi.Error(w, service.ErrInvalidCredentials)
		return
	}

	itemID, err := strconv.ParseInt(r.PathValue("item_id"), 10, 64)
	if err != nil {
		httpapi.BadRequest(w, "item_id must be an integer")
		return
	}

	var req dto.UpsertCartItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpapi.BadRequest(w, "invalid request body")
		return
	}

	if err := h.cart.UpsertItem(r.Context(), int64(userID), itemID, req.Quantity); err != nil {
		httpapi.Error(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *CartHandler) RemoveItem(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserID(r.Context())
	if !ok {
		httpapi.Error(w, service.ErrInvalidCredentials)
		return
	}

	itemID, err := strconv.ParseInt(r.PathValue("item_id"), 10, 64)
	if err != nil {
		httpapi.BadRequest(w, "item_id must be an integer")
		return
	}

	if err := h.cart.RemoveItem(r.Context(), int64(userID), itemID); err != nil {
		httpapi.Error(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
