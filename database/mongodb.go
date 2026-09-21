// Package database provides a MongoDB database connection and related functionality.
package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Database is the shared MongoDB database connection instance.
var Database *mongo.Database

// ConnectToDB opens a connection using the given URI and pings it
// to confirm the connection actually works before continuing.
func ConnectToDB(uri string) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	clientOpts := options.Client().ApplyURI(uri)
	clientOpts.SetServerSelectionTimeout(15 * time.Second)
	clientOpts.SetConnectTimeout(15 * time.Second)

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		log.Fatal("Failed to connect to MongoDB: ", err)
	}

	pingCtx, pingCancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer pingCancel()

	if err := client.Ping(pingCtx, nil); err != nil {
		log.Fatal("Failed to ping MongoDB: ", err)
	}

	fmt.Println("Connected to MongoDB")
	Database = client.Database("authgo") // The name of the database inside the cluster
}
