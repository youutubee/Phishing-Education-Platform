package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client
var DB *mongo.Database

func InitDB() (*mongo.Client, *mongo.Database, error) {
	mongoURL := os.Getenv("MONGODB_URL")
	if mongoURL == "" {
		mongoURL = "mongodb://localhost:27017"
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "seap"
	}

	// Use a longer timeout for initial connection (especially for Atlas)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Configure client options with proper settings for Atlas
	clientOptions := options.Client().ApplyURI(mongoURL)

	// Set server selection timeout
	clientOptions.SetServerSelectionTimeout(30 * time.Second)

	// Set connect timeout
	clientOptions.SetConnectTimeout(30 * time.Second)

	// For Atlas connections, ensure TLS is enabled (it should be in the connection string)
	// If the connection string starts with mongodb+srv://, TLS is automatically enabled

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping the database with a separate context
	pingCtx, pingCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer pingCancel()

	if err := client.Ping(pingCtx, nil); err != nil {
		client.Disconnect(context.Background())
		return nil, nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	log.Println("Database connected successfully")

	Client = client
	DB = client.Database(dbName)

	// Create indexes
	if err := createIndexes(DB); err != nil {
		log.Printf("Warning: Failed to create indexes: %v", err)
	}

	return client, DB, nil
}

func createIndexes(db *mongo.Database) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Users collection indexes
	usersCollection := db.Collection("users")
	_, err := usersCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    map[string]interface{}{"email": 1},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return fmt.Errorf("failed to create users email index: %w", err)
	}

	// Campaigns collection indexes
	campaignsCollection := db.Collection("campaigns")
	_, err = campaignsCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    map[string]interface{}{"tracking_token": 1},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return fmt.Errorf("failed to create campaigns tracking_token index: %w", err)
	}
	_, err = campaignsCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: map[string]interface{}{"user_id": 1},
	})
	if err != nil {
		return fmt.Errorf("failed to create campaigns user_id index: %w", err)
	}

	// Events collection indexes
	eventsCollection := db.Collection("events")
	_, err = eventsCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: map[string]interface{}{"campaign_id": 1},
	})
	if err != nil {
		return fmt.Errorf("failed to create events campaign_id index: %w", err)
	}
	_, err = eventsCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: map[string]interface{}{"created_at": -1},
	})
	if err != nil {
		return fmt.Errorf("failed to create events created_at index: %w", err)
	}

	// OTPs collection indexes
	otpsCollection := db.Collection("otps")
	_, err = otpsCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "email", Value: 1},
			{Key: "created_at", Value: -1},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create otps index: %w", err)
	}

	// Audit logs collection indexes
	auditLogsCollection := db.Collection("audit_logs")
	_, err = auditLogsCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: map[string]interface{}{"created_at": -1},
	})
	if err != nil {
		return fmt.Errorf("failed to create audit_logs created_at index: %w", err)
	}

	log.Println("Indexes created successfully")
	return nil
}
