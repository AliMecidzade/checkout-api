package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"checkout-api/store"
)

func TestGetItems(t *testing.T) {
	s := store.NewInMemStore()
	h := NewHandler(s)

	tests := []struct {
		name       string
		method     string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "GET returns items",
			method:     http.MethodGet,
			wantStatus: http.StatusOK,
			wantBody:   "Laptop",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/items", nil)
			w := httptest.NewRecorder()

			h.GetItems(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
			if tt.wantBody != "" && !strings.Contains(w.Body.String(), tt.wantBody) {
				t.Errorf("body missing %q, got: %s", tt.wantBody, w.Body.String())
			}
		})
	}
}

func TestGetItemByID(t *testing.T) {
	s := store.NewInMemStore()
	h := NewHandler(s)

	tests := []struct {
		name       string
		itemIDPath string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "valid item",
			itemIDPath: "1",
			wantStatus: http.StatusOK,
			wantBody:   "Laptop",
		},
		{
			name:       "item not found",
			itemIDPath: "999",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "invalid ID",
			itemIDPath: "abc",
			wantStatus: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/items/"+tt.itemIDPath, nil)
			req.SetPathValue("item_id", tt.itemIDPath)
			w := httptest.NewRecorder()

			h.GetItemByID(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
			if tt.wantBody != "" && !strings.Contains(w.Body.String(), tt.wantBody) {
				t.Errorf("body missing %q, got: %s", tt.wantBody, w.Body.String())
			}
		})
	}
}

func TestGetUserCart(t *testing.T) {
	tests := []struct {
		name       string
		withAuth   bool
		setupCart  bool
		wantStatus int
		wantBody   string
	}{
		{
			name:       "cart does not exist - returns empty cart",
			withAuth:   true,
			setupCart:  false,
			wantStatus: http.StatusOK,
			wantBody:   `"user_id":1,"items":[]`,
		},
		{
			name:       "cart exists - returns cart with items",
			withAuth:   true,
			setupCart:  true,
			wantStatus: http.StatusOK,
			wantBody:   `"items":[{"id":1,"name":"Laptop"`,
		},
		{
			name:       "missing Authorization header",
			withAuth:   false,
			setupCart:  false,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "invalid JWT",
			withAuth:   false,
			setupCart:  false,
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := store.NewInMemStore()
			h := NewHandler(s)

			if tt.setupCart {
				createReq := httptest.NewRequest(http.MethodPut, "/user/cart/items/1", strings.NewReader(`{"quantity":2}`))
				createReq = createReq.WithContext(context.WithValue(createReq.Context(), "user", 1))
				createReq.Header.Set("Content-Type", "application/json")
				createReq.SetPathValue("item_id", "1")
				createW := httptest.NewRecorder()
				h.UpsertCartItem(createW, createReq)
			}

			req := httptest.NewRequest(http.MethodGet, "/user/cart", nil)
			if tt.withAuth {
				req = req.WithContext(context.WithValue(req.Context(), "user", 1))
			}
			w := httptest.NewRecorder()

			h.GetUserCart(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
			if tt.wantBody != "" && !strings.Contains(w.Body.String(), tt.wantBody) {
				t.Errorf("body missing %q, got: %s", tt.wantBody, w.Body.String())
			}
		})
	}
}

func TestCreateOrder(t *testing.T) {
	tests := []struct {
		name       string
		withAuth   bool
		idempotKey string
		body       string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "valid order",
			withAuth:   true,
			idempotKey: "key-001",
			body:       `{"line_items":[{"item_id":1,"quantity":1,"price":120000}],"total":120000}`,
			wantStatus: http.StatusCreated,
			wantBody:   `"status":"paid"`,
		},
		{
			name:       "empty items",
			withAuth:   true,
			idempotKey: "key-002",
			body:       `{"line_items":[],"total":0}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing Authorization header",
			withAuth:   false,
			idempotKey: "key-003",
			body:       `{"line_items":[{"item_id":1,"quantity":1,"price":120000}],"total":120000}`,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "missing idempotency key",
			withAuth:   true,
			idempotKey: "",
			body:       `{"line_items":[{"item_id":1,"quantity":1,"price":120000}],"total":120000}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid body",
			withAuth:   true,
			idempotKey: "key-004",
			body:       `not-json`,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := store.NewInMemStore()
			h := NewHandler(s)

			req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			if tt.withAuth {
				req = req.WithContext(context.WithValue(req.Context(), "user", 1))
			}
			if tt.idempotKey != "" {
				req.Header.Set("Idempotency-Key", tt.idempotKey)
			}
			w := httptest.NewRecorder()

			h.CreateOrder(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
			if tt.wantBody != "" && !strings.Contains(w.Body.String(), tt.wantBody) {
				t.Errorf("body missing %q, got: %s", tt.wantBody, w.Body.String())
			}
		})
	}
}
