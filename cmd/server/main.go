package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	api "checkout-api/internal/httpapi/handlers"
	"checkout-api/internal/httpapi/middleware"
	"checkout-api/internal/repository/postgres"
	"checkout-api/internal/service"
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

	pg := postgres.NewPostgresStore(pool)

	authSvc := service.NewAuthService(pg, pg, []byte(os.Getenv("SIGNING_SECRET")))
	cartSvc := service.NewCartService(pg, pg)
	orderSvc := service.NewOrderService(pg)
	catalogSvc := service.NewCatalogService(pg)

	authHandler := api.NewAuthHandler(authSvc)
	cartHandler := api.NewCartHandler(cartSvc)
	orderHandler := api.NewOrdersHandler(orderSvc)
	itemsHandler := api.NewItemsHandler(catalogSvc)

	// cart
	mux.HandleFunc("GET /user/cart", middleware.AuthMiddleware(authSvc, cartHandler.GetCart))
	mux.HandleFunc("PATCH /user/cart/items/{item_id}", middleware.AuthMiddleware(authSvc, cartHandler.UpsertItem))
	mux.HandleFunc("DELETE /user/cart/items/{item_id}", middleware.AuthMiddleware(authSvc, cartHandler.RemoveItem))

	// orders
	mux.HandleFunc("POST /orders", middleware.AuthMiddleware(authSvc, orderHandler.CreateOrder))
	mux.HandleFunc("GET /user/orders", middleware.AuthMiddleware(authSvc, orderHandler.GetUserOrders))
	// items
	mux.HandleFunc("GET /items", itemsHandler.List)
	mux.HandleFunc("GET /items/{item_id}", itemsHandler.GetByID)

	// auth
	mux.HandleFunc("POST /signup", authHandler.CreateUser)
	mux.HandleFunc("POST /login", authHandler.LoginUser)
	mux.HandleFunc("GET /token", authHandler.Refresh)
	mux.HandleFunc("POST /logout", authHandler.Logout)

	fmt.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", middleware.WithCORS(mux)))
}
