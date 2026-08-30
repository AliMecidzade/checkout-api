package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"checkout-api/handlers"
	api "checkout-api/internal/httpapi/handlers"
	"checkout-api/internal/httpapi/middleware"
	"checkout-api/internal/repository/postgres"
	"checkout-api/internal/service"
	"checkout-api/store"
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

	pg := postgres.NewPostgresStore(pool)
	authSvc := service.NewAuthService(pg, pg, []byte(os.Getenv("SIGNING_SECRET")))
	authHandler := api.NewAuthHandler(authSvc)
	// cart
	mux.HandleFunc("GET /user/cart", middleware.AuthMiddleware(authSvc, h.GetUserCart))
	mux.HandleFunc("PATCH /user/cart/items/{item_id}", middleware.AuthMiddleware(authSvc, h.UpsertCartItem))
	mux.HandleFunc("DELETE /user/cart/items/{item_id}", middleware.AuthMiddleware(authSvc, h.RemoveCartItem))

	// orders
	mux.HandleFunc("POST /orders", middleware.AuthMiddleware(authSvc, h.CreateOrder))
	mux.HandleFunc("GET /user/orders", middleware.AuthMiddleware(authSvc, h.GetUserOrders))
	// items
	mux.HandleFunc("GET /items", h.GetItems)
	mux.HandleFunc("GET /items/{item_id}", h.GetItemByID)

	// auth
	mux.HandleFunc("POST /signup", authHandler.CreateUser)
	mux.HandleFunc("POST /login", authHandler.LoginUser)
	mux.HandleFunc("GET /token", authHandler.Refresh)
	mux.HandleFunc("POST /logout", authHandler.Logout)

	handlers := handlers.WithCORS(mux)

	fmt.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", handlers))
}
