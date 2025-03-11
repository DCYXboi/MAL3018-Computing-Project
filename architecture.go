package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDB Connection URI
const mongoURI = "mongodb://localhost:27017"

// MongoDB Database & Collection Names
const dbName = "blockchainDB"
const collectionName = "blocks"

var mongoClient *mongo.Client
var blockCollection *mongo.Collection

// Block represents each 'block' in the blockchain
type Block struct {
	Index     int    `bson:"index"`
	Timestamp string `bson:"timestamp"`
	Data      string `bson:"data"`
	Hash      string `bson:"hash"`
	PrevHash  string `bson:"prev_hash"`
}

// Initialize MongoDB Connection
func connectMongoDB() {
	clientOptions := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatalf("MongoDB Connection Error: %v", err)
	}
	if err := client.Ping(context.TODO(), nil); err != nil {
		log.Fatalf("MongoDB Ping Failed: %v", err)
	}

	mongoClient = client
	blockCollection = client.Database(dbName).Collection(collectionName)
	fmt.Println("Connected to MongoDB!")
}

// SaveBlock stores a new block in MongoDB
func SaveBlock(block Block) {
	_, err := blockCollection.InsertOne(context.TODO(), block)
	if err != nil {
		log.Printf("Failed to save block: %v", err)
	}
}

// LoadBlockchain loads blocks from MongoDB
func LoadBlockchain() []Block {
	var blocks []Block
	cursor, err := blockCollection.Find(context.TODO(), bson.D{})
	if err != nil {
		log.Printf("Failed to load blockchain: %v", err)
		return blocks
	}
	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		var block Block
		err := cursor.Decode(&block)
		if err != nil {
			log.Printf("Error decoding block: %v", err)
			continue
		}
		blocks = append(blocks, block)
	}

	return blocks
}

// HTTP Handler to Get Blockchain from MongoDB
func getBlockchain(w http.ResponseWriter, r *http.Request) {
	blocks := LoadBlockchain()
	for _, block := range blocks {
		fmt.Fprintf(w, "Index: %d\nTimestamp: %s\nData: %s\nHash: %s\nPrevHash: %s\n\n",
			block.Index, block.Timestamp, block.Data, block.Hash, block.PrevHash)
	}
}

// Main Function
func main() {
	// Connect to MongoDB
	connectMongoDB()

	// Load existing blockchain from database
	Blockchain := LoadBlockchain()
	if len(Blockchain) == 0 {
		genesisBlock := Block{
			Index:     0,
			Timestamp: time.Now().String(),
			Data:      "Genesis Block",
			Hash:      fmt.Sprintf("%x", time.Now().UnixNano()),
			PrevHash:  "",
		}
		Blockchain = append(Blockchain, genesisBlock)
		SaveBlock(genesisBlock)
	}

	// HTTP Handlers
	http.HandleFunc("/blockchain", getBlockchain)

	// Start HTTP Server
	port := ":8080"
	fmt.Printf("Server running on http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
