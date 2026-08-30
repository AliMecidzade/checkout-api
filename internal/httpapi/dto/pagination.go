package dto

type Pagination struct {
	Limit      int    `json:"limit"`
	Offset     int    `json:"offset,omitempty"`
	HasMore    bool   `json:"has_more"`
	NextCursor string `json:"next_cursor,omitempty"`
}

type ItemsListResponse struct {
	Items      []ItemResponse `json:"items"`
	Pagination Pagination     `json:"pagination"`
}

type OrdersListResponse struct {
	Orders     []OrderResponse `json:"orders"`
	Pagination Pagination      `json:"pagination"`
}

type CartListResponse struct {
	UserID     int64              `json:"user_id"`
	Items      []CartItemResponse `json:"items"`
	Pagination Pagination         `json:"pagination"`
}
