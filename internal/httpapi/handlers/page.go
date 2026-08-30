package handlers

import (
	"net/http"
	"strconv"

	"checkout-api/internal/domain"
)

func parsePage(r *http.Request) (domain.Page, bool) {
	page := domain.Page{Limit: 20, Offset: 0}

	if v := r.URL.Query().Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			return page, false
		}
		page.Limit = n
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			return page, false
		}
		page.Offset = n
	}
	if v := r.URL.Query().Get("cursor"); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil || id <= 0 {
			return page, false
		}
		page.Cursor = &domain.Cursor{ID: id}
	}
	return page, true
}
