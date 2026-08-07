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

//-----------------------TEST MONGODB OPERATIONS----------------------
//$env:Path = [System.Environment]::GetEnvironmentVariable("Path","User")
//mongosh --quiet --eval "db.runCommand({ping:1})"
// go run ./cmd/mongo_demo

//-----------------TEST PATCH ITEM STOCK WITH COMMAND-----------------------
//curl.exe -X PATCH http://localhost:8090/items/3/stock -H "Content-Type: application/json" -d '{"quantity": 5}'

//---------------------------------------------------------------------------

//---------------------------RESETTING THE DB----------------------------------
//psql "$DATABASE_URL" -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public

//migrate -database "$DATABASE_URL" -path migrations up 8

//----------------------------------------------------------------------------

//-------------------EXPLAIN ANALYZE---------------------------------------

//migrate -database "postgresql://checkout:secret@localhost:5432/checkout" -path migrations up 8
// migrate -database "postgresql://checkout:secret@localhost:5432/checkout" -path migrations down -all

//INSERT INTO orders (user_id,total,  status)
//SELECT
//(i % 1000) + 1,
//(random() * 100000)::int,
//CASE WHEN random() < 0.6 THEN 'paid'
//WHEN random() < 0.9 THEN 'pending'
//ELSE 'failed' END
//
//FROM generate_series(1, 150000) AS i;

//EXPLAIN ANALYZE SELECT * FROM orders WHERE user_id = 1 AND status = 'paid';

//--------------------------------------------------------------------------

// -----------------Normalization Task Slide 12---------------------------
// customers - cust_id, cust_name,state, city
// products - itemNo, itemDesc, itemPrice
// orders - order_id, date, cust_id
// order_items - order_id, item_no, quantity

//racing demo

// go run cmd/race_demo/main.go
// go run cmd/race_demo/main.go -lock

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("failed to load .env: %v", err)
	}

	ctx := context.Background()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgresql://checkout:secret@localhost:5432/checkout"
	}
	conn, err := pgx.Connect(ctx, dsn)
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
	http.HandleFunc("/items/{id}/stock", h.UpdateItemStock)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Println("Server starting on :" + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
