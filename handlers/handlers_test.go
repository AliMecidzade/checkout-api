package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"checkout-api/models"
	"checkout-api/store"
)

func TestGetItems(t *testing.T) {
	s := store.NewStore()
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
		{
			name:       "POST not allowed",
			method:     http.MethodPost,
			wantStatus: http.StatusMethodNotAllowed,
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
	s := store.NewStore()
	h := NewHandler(s)

	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "valid item",
			method:     http.MethodGet,
			path:       "/items/1",
			wantStatus: http.StatusOK,
			wantBody:   "Laptop",
		},
		{
			name:       "item not found",
			method:     http.MethodGet,
			path:       "/items/999",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "invalid ID",
			method:     http.MethodGet,
			path:       "/items/abc",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "wrong method",
			method:     http.MethodPost,
			path:       "/items/1",
			wantStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
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

func TestCreateOrder(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		body       string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "valid order",
			method:     http.MethodPost,
			body:       `{"user_id":1,"items":[{"item_id":1,"quantity":1}]}`,
			wantStatus: http.StatusCreated,
			wantBody:   `"status":"paid"`,
		},
		{
			name:       "invalid item",
			method:     http.MethodPost,
			body:       `{"user_id":1,"items":[{"item_id":999,"quantity":1}]}`,
			wantStatus: http.StatusBadRequest,
			wantBody:   "not found",
		},
		{
			name:       "empty body",
			method:     http.MethodPost,
			body:       "",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "wrong method",
			method:     http.MethodGet,
			body:       "",
			wantStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := store.NewStore()
			h := NewHandler(s)

			req := httptest.NewRequest(tt.method, "/orders", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
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

func TestCreateUserCartAndAddItems(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		body       string
		setup      func(s *store.Store)
		wantStatus int
		wantBody   string
	}{
		{
			name:       "valid cart",
			method:     http.MethodPost,
			body:       `{"user_id":1,"items":[{"item_id":1,"quantity":2}]}`,
			wantStatus: http.StatusCreated,
			wantBody:   `"user_id":1`,
		},
		{
			name:   "cart already exists",
			method: http.MethodPost,
			body:   `{"user_id":1,"items":[{"item_id":1,"quantity":1}]}`,
			setup: func(s *store.Store) {
				s.CreateUserCart(&models.Cart{ID: "cart_existing", UserID: 1, Items: []models.LineItem{}})
			},
			wantStatus: http.StatusConflict,
			wantBody:   "cart already exists",
		},
		{
			name:       "empty items",
			method:     http.MethodPost,
			body:       `{"user_id":2,"items":[]}`,
			wantStatus: http.StatusBadRequest,
			wantBody:   "can't create empty cart",
		},
		{
			name:       "item not found",
			method:     http.MethodPost,
			body:       `{"user_id":3,"items":[{"item_id":999,"quantity":1}]}`,
			wantStatus: http.StatusBadRequest,
			wantBody:   "not found",
		},
		{
			name:       "invalid body",
			method:     http.MethodPost,
			body:       `{invalid`,
			wantStatus: http.StatusBadRequest,
			wantBody:   "invalid request body",
		},
		{
			name:       "wrong method",
			method:     http.MethodGet,
			body:       "",
			wantStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := store.NewStore()
			if tt.setup != nil {
				tt.setup(s)
			}
			h := NewHandler(s)

			req := httptest.NewRequest(tt.method, "/carts", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			h.CreateUserCartAndAddItems(w, req)

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
		method     string
		userID     string
		setUserID  bool
		setup      func(s *store.Store)
		wantStatus int
		wantBody   string
	}{
		{
			name:      "existing cart",
			method:    http.MethodGet,
			userID:    "1",
			setUserID: true,
			setup: func(s *store.Store) {
				s.CreateUserCart(&models.Cart{
					ID:     "cart_1",
					UserID: 1,
					Items:  []models.LineItem{{ItemID: 1, Quantity: 2, Price: 120000}},
				})
			},
			wantStatus: http.StatusOK,
			wantBody:   `"id":"cart_1"`,
		},
		{
			name:       "no cart returns empty cart",
			method:     http.MethodGet,
			userID:     "42",
			setUserID:  true,
			wantStatus: http.StatusOK,
			wantBody:   `"id":""`,
		},
		{
			name:       "missing header",
			method:     http.MethodGet,
			setUserID:  false,
			wantStatus: http.StatusBadRequest,
			wantBody:   "missing X-User-ID header",
		},
		{
			name:       "invalid header",
			method:     http.MethodGet,
			userID:     "abc",
			setUserID:  true,
			wantStatus: http.StatusBadRequest,
			wantBody:   "invalid X-User-ID header",
		},
		{
			name:       "wrong method",
			method:     http.MethodPost,
			userID:     "1",
			setUserID:  true,
			wantStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := store.NewStore()
			if tt.setup != nil {
				tt.setup(s)
			}
			h := NewHandler(s)

			req := httptest.NewRequest(tt.method, "/carts", nil)
			if tt.setUserID {
				req.Header.Set("X-User-ID", tt.userID)
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
