package main

import (
	"checkout-api/handlers"
	"checkout-api/store"
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5"
)

func main() {
	if err := loadEnv(".env"); err != nil {
		log.Printf("failed to load .env: %v", err)
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, getenv("DATABASE_URL", "postgresql://postgres:IRONMANisaHUMAN789@localhost:5432/postgres"))
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

	http.HandleFunc("/user/cart/items/",
		func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodPatch:
				h.UpdateCartItem(w, r)
			case http.MethodDelete:
				h.RemoveCartItem(w, r)
			default:
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
		})
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
	http.HandleFunc("/user/orders", h.CreateOrderFromCart)

	http.HandleFunc("/signup", h.SignUp)
	http.HandleFunc("/items", h.GetItems)
	http.HandleFunc("/items/", h.GetItemByID)

	port := getenv("PORT", "8080")
	fmt.Println("Server starting on :" + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
