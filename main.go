package main

import (
	"checkout-api/handlers"
	"checkout-api/store"
	"log"
	"net/http"
)

func main() {
	itemStore := store.NewStore()
	handler := handlers.NewHandler(itemStore)
	http.HandleFunc("/items", handler.GetItems)
	http.HandleFunc("/users/", handler.GetOrdersByUserId)
	http.HandleFunc("/items/", handler.GetItem)
	http.HandleFunc("/orders", handler.CreateOrder)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
