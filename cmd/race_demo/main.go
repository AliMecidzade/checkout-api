package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
)

const (
	itemID   = 1
	stockSQL = "SELECT stock FROM items WHERE id = $1"
)

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func buy(ctx context.Context, conn *pgx.Conn, withLock bool) (string, time.Duration) {
	start := time.Now()
	query := stockSQL
	if withLock {
		query += " FOR UPDATE"
	}

	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Sprintf("error: %v", err), time.Since(start)
	}
	defer tx.Rollback(ctx)

	var stock int
	if err := tx.QueryRow(ctx, query, itemID).Scan(&stock); err != nil {
		return fmt.Sprintf("error: %v", err), time.Since(start)
	}

	time.Sleep(100 * time.Millisecond)

	if stock < 1 {
		return "rejected (insufficient stock)", time.Since(start)
	}

	if _, err := tx.Exec(ctx, "UPDATE items SET stock = stock - 1 WHERE id = $1", itemID); err != nil {
		return fmt.Sprintf("error: %v", err), time.Since(start)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Sprintf("error: %v", err), time.Since(start)
	}

	return "sold", time.Since(start)
}

func main() {
	withLock := flag.Bool("lock", false, "use SELECT ... FOR UPDATE")
	flag.Parse()

	dsn := getenv("DATABASE_URL", "postgresql://checkout:secret@localhost:5432/checkout")
	ctx := context.Background()

	connA, err := pgx.Connect(ctx, dsn)
	if err != nil {
		log.Fatal("connect A:", err)
	}
	defer connA.Close(ctx)

	connB, err := pgx.Connect(ctx, dsn)
	if err != nil {
		log.Fatal("connect B:", err)
	}
	defer connB.Close(ctx)

	if _, err := connA.Exec(ctx, "UPDATE items SET stock = 1 WHERE id = $1", itemID); err != nil {
		log.Fatal("reset stock:", err)
	}

	mode := "no lock (racy)"
	if *withLock {
		mode = "SELECT ... FOR UPDATE"
	}
	fmt.Printf("--- mode: %s, item %d, starting stock = 1 ---\n", mode, itemID)

	start := make(chan struct{})
	var wg sync.WaitGroup
	results := make([]string, 2)
	durations := make([]time.Duration, 2)

	run := func(conn *pgx.Conn, i int) {
		defer wg.Done()
		<-start
		results[i], durations[i] = buy(ctx, conn, *withLock)
	}

	wg.Add(2)
	go run(connA, 0)
	go run(connB, 1)
	close(start)
	wg.Wait()

	for i, res := range results {
		fmt.Printf("goroutine %d: %-32s (%v)\n", i, res, durations[i])
	}

	var finalStock int
	if err := connA.QueryRow(ctx, stockSQL, itemID).Scan(&finalStock); err != nil {
		log.Fatal("final stock:", err)
	}
	fmt.Printf("final stock: %d\n", finalStock)
}
