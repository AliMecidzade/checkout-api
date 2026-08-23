package main

import (
	"checkout-api/handlers"
	"checkout-api/store"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	mux := http.NewServeMux()
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))

	if err != nil {
		panic(err)
	}

	err = pool.Ping(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Println("successfully connected to db")

	// s := store.NewInMemStore()
	postgresStore := store.NewPostgresStore(pool)
	h := handlers.NewHandler(postgresStore)

	// cart
	mux.HandleFunc("GET /user/cart", handlers.AuthMiddleware(h.GetUserCart))
	mux.HandleFunc("PATCH /user/cart/items/{item_id}", handlers.AuthMiddleware(h.UpsertCartItem))
	mux.HandleFunc("DELETE /user/cart/items/{item_id}", handlers.AuthMiddleware(h.RemoveCartItem))

	// orders
	mux.HandleFunc("POST /orders", handlers.AuthMiddleware(h.CreateOrder))

	// items
	mux.HandleFunc("GET /items", h.GetItems)
	mux.HandleFunc("GET /items/{item_id}", h.GetItemByID)

	// users
	mux.HandleFunc("POST /signup", h.CreateUser)
	mux.HandleFunc("POST /login", h.LoginUser)
	mux.HandleFunc("GET /token", h.IssueJWT)
	mux.HandleFunc("POST /logout", h.Logout)

	handlers := handlers.WithCORS(mux)

	fmt.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", handlers))
}
