package service

import "checkout-api/internal/domain"

func normalizePage(p domain.Page) domain.Page {
	if p.Limit <= 0 {
		p.Limit = 20
	}
	if p.Limit > 100 {
		p.Limit = 100
	}
	return p
}
