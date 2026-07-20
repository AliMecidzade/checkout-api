package handlers

// Implement HTTP handlers:
// - NewHandler(s *store.Store) — constructor
// - GetItems: GET /items → return all items as JSON
// - CreateOrder: POST /orders → decode request, validate items, calculate total, mock payment, create order
