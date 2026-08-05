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

//---------------------------RESETTING THE DB----------------------------------
//psql "$DATABASE_URL" -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public

//migrate -database "$DATABASE_URL" -path migrations up 8

//----------------------------------------------------------------------------

//-------------------EXPLAIN ANALYZE---------------------------------------
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
func main() {
	if err := loadEnv(".env"); err != nil {
		log.Printf("failed to load .env: %v", err)
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, getenv("DATABASE_URL", "postgresql://checkout:secret@localhost:5432/checkout"))
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
