package main

import (
	"checkout-api/handlers"
	"checkout-api/store"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, os.Getenv("DATABASE_URL"))

	if err != nil {
		panic(err)
	}

	err = conn.Ping(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Println("successfully connected to db")

	// s := store.NewInMemStore()
	postgresStore := store.NewPostgresStore(conn)
	h := handlers.NewHandler(postgresStore)

	// cart
	http.HandleFunc("GET /user/cart", handlers.AuthMiddleware(h.GetUserCart))
	http.HandleFunc("PATCH /user/cart/items/{item_id}", handlers.AuthMiddleware(h.UpsertCartItem))
	http.HandleFunc("DELETE /user/cart/items/{item_id}", handlers.AuthMiddleware(h.RemoveCartItem))

	// orders
	http.HandleFunc("POST /orders", handlers.AuthMiddleware(h.CreateOrder))

	// items
	http.HandleFunc("GET /items", h.GetItems)
	http.HandleFunc("GET /items/{item_id}", h.GetItemByID)

	// users
	http.HandleFunc("POST /signup", h.CreateUser)
	http.HandleFunc("POST /login", h.LoginUser)
	http.HandleFunc("GET /token", h.IssueJWT)

	fmt.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
