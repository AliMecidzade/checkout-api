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
	http.HandleFunc("/items/", handler.GetItemByID)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
