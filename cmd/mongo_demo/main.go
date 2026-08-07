package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

type Item struct {
	Name  string `bson:"name"`
	Price int    `bson:"price"`
	Stock int    `bson:"stock"`
}

func main() {
	uri := getenv("MONGO_URL", "mongodb://localhost:27017")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal("connect:", err)
	}
	defer client.Disconnect(ctx)

	items := client.Database("checkout").Collection("items")

	// add a document
	res, err := items.InsertOne(ctx, Item{Name: "T-Shirt", Price: 2500, Stock: 120})
	if err != nil {
		log.Fatal("insert:", err)
	}
	fmt.Println("[INSERT] id:", res.InsertedID)

	//  read it back
	var found Item
	if err := items.FindOne(ctx, bson.M{"name": "T-Shirt"}).Decode(&found); err != nil {
		log.Fatal("find:", err)
	}
	fmt.Printf("[RETRIEVE] %+v\n", found)

	// 3. Update — change the stock
	upd, err := items.UpdateOne(ctx, bson.M{"name": "T-Shirt"}, bson.M{"$set": bson.M{"stock": 115}})
	if err != nil {
		log.Fatal("update:", err)
	}
	fmt.Println("[UPDATE] modified:", upd.ModifiedCount)

	if err := items.FindOne(ctx, bson.M{"name": "T-Shirt"}).Decode(&found); err != nil {
		log.Fatal("find after update:", err)
	}
	fmt.Printf("[UPDATE] after: %+v\n", found)

	// remove the document
	//del, err := items.DeleteOne(ctx, bson.M{"name": "T-Shirt"})
	//if err != nil {
	//	log.Fatal("delete:", err)
	//}
	//fmt.Println("[DELETE] deleted:", del.DeletedCount)
}
