package main

import (
	"log"
	"net/http"

	"checkout-api/handlers"
	"checkout-api/store"
)

func main() {
	s := store.NewStore()
	h := handlers.NewHandler(s)

	http.HandleFunc("/user/cart",
		func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				h.GetUserCart(w, r)
			case http.MethodPost:
				h.CreateUserCartAndAddItems(w, r)
			default:
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
		})
	http.HandleFunc("/user/cart/items/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPatch:
			h.UpdateCartItemQuantity(w, r)
		case http.MethodDelete:
			h.DeleteItemFromCart(w, r)

		}
	})
	http.HandleFunc("/items", h.GetItems)
	http.HandleFunc("/items/", h.GetItemByID)
	http.HandleFunc("/orders", h.CreateOrder)
	http.HandleFunc("/user/orders", h.PlaceOrder)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
