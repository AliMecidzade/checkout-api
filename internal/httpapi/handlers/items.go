package handlers

import (
	"net/http"
	"strconv"

	"checkout-api/internal/httpapi"
	"checkout-api/internal/httpapi/dto"
	"checkout-api/internal/service"
)

type ItemsHandler struct {
	catalog *service.CatalogService
}

func NewItemsHandler(catalog *service.CatalogService) *ItemsHandler {
	return &ItemsHandler{catalog: catalog}
}

func (h *ItemsHandler) List(w http.ResponseWriter, r *http.Request) {
	page, ok := parsePage(r)
	if !ok {
		httpapi.BadRequest(w, "limit, offset and cursor must be valid")
		return
	}

	items, hasMore, err := h.catalog.ListItems(r.Context(), page)
	if err != nil {
		httpapi.Error(w, err)
		return
	}

	out := make([]dto.ItemResponse, 0, len(items))
	for _, it := range items {
		out = append(out, dto.ItemResponse{
			ID: it.ID, Name: it.Name, Description: it.Description,
			Price: it.Price, Stock: it.Stock, CreatedAt: it.CreatedAt,
		})
	}

	pagination := dto.Pagination{Limit: page.Limit, Offset: page.Offset, HasMore: hasMore}
	if hasMore && len(items) > 0 {
		pagination.NextCursor = strconv.FormatInt(items[len(items)-1].ID, 10)
	}

	httpapi.JSON(w, http.StatusOK, dto.ItemsListResponse{Items: out, Pagination: pagination})
}

func (h *ItemsHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("item_id"), 10, 64)
	if err != nil {
		httpapi.BadRequest(w, "item_id must be an integer")
		return
	}

	item, err := h.catalog.GetItem(r.Context(), id)
	if err != nil {
		httpapi.Error(w, err)
		return
	}

	httpapi.JSON(w, http.StatusOK, dto.ItemResponse{
		ID: item.ID, Name: item.Name, Description: item.Description,
		Price: item.Price, Stock: item.Stock, CreatedAt: item.CreatedAt,
	})
}
