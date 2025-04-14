package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Block2 represents each 'block' in the blockchain for Segment 2
type Block2 struct {
	Index     int    `json:"index"`
	Timestamp string `json:"timestamp"`
	Data      string `json:"data"`
	Hash      string `json:"hash"`
	PrevHash  string `json:"prev_hash"`
}

var (
	mongoClient2     *mongo.Client
	blockCollection2 *mongo.Collection
	Blockchain2      []Block2
	mongoURI2        = "mongodb://localhost:27017" // Same MongoDB URI as architecture.go
	dbName2          = "blockchainDB"              // Same database name as architecture.go
	collectionName2  = "blocks"                    // Same collection name as architecture.go
)

// Function to calculate SHA-256 hash for a block in Segment 2
func calculateHash2(block Block2) string {
	record := fmt.Sprintf("%d%s%s%s", block.Index, block.Timestamp, block.Data, block.PrevHash)
	hash := sha256.Sum256([]byte(record))
	return hex.EncodeToString(hash[:])
}

// Initialize MongoDB Connection
func connectMongoDB2() {
	clientOptions := options.Client().ApplyURI(mongoURI2)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatalf("MongoDB Connection Error: %v", err)
	}
	if err := client.Ping(context.TODO(), nil); err != nil {
		log.Fatalf("MongoDB Ping Failed: %v", err)
	}

	mongoClient2 = client
	blockCollection2 = client.Database(dbName2).Collection(collectionName2)
	fmt.Println("Connected to MongoDB for Segment 2!")
}

// SaveBlock2 stores a new block in MongoDB
func saveBlock2(block Block2) {
	_, err := blockCollection.InsertOne(context.TODO(), block)
	if err != nil {
		log.Printf("Failed to save block: %v", err)
	}
}

// LoadBlockchain2 loads the blockchain from MongoDB
func loadBlockchain2() {
	cursor, err := blockCollection.Find(context.TODO(), bson.D{})
	if err != nil {
		log.Fatalf("Failed to load blockchain: %v", err)
	}
	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		var block Block2
		if err := cursor.Decode(&block); err != nil {
			log.Printf("Failed to decode block: %v", err)
			continue
		}
		Blockchain2 = append(Blockchain2, block)
	}
	fmt.Println("Blockchain loaded from MongoDB!")
}

// HTTP Handler to Add a New Block to Segment 2
func addBlockHandler2Post(w http.ResponseWriter, r *http.Request) {
	// Ensure the request method is POST
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method. Only POST is allowed.", http.StatusMethodNotAllowed)
		return
	}

	// Parse the form data
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form data.", http.StatusBadRequest)
		return
	}

	// Validate input data
	data := r.FormValue("data")
	if data == "" {
		http.Error(w, "Data field is required.", http.StatusBadRequest)
		return
	}

	// Create a new block
	prevBlock := Blockchain2[len(Blockchain2)-1]
	newBlock := Block2{
		Index:     prevBlock.Index + 1,
		Timestamp: time.Now().String(),
		Data:      data,
		PrevHash:  prevBlock.Hash,
	}
	newBlock.Hash = calculateHash2(newBlock)

	// Append the new block to the Segment 2 blockchain
	Blockchain2 = append(Blockchain2, newBlock)
	saveBlock2(newBlock)

	// Respond with success
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Block added successfully to Segment 2: %+v\n", newBlock)
}

func Main2() {
	// Connect to MongoDB
	connectMongoDB2()

	// Load existing blockchain from MongoDB
	loadBlockchain2()

	// Initialize blockchain if empty
	if len(Blockchain2) == 0 {
		genesisBlock := Block2{
			Index:     0,
			Timestamp: time.Now().String(),
			Data:      "Genesis Block",
			Hash:      calculateHash2(Block2{Index: 0, Timestamp: time.Now().String(), Data: "Genesis Block", PrevHash: ""}),
			PrevHash:  "",
		}
		Blockchain2 = append(Blockchain2, genesisBlock)
		saveBlock2(genesisBlock)
		fmt.Println("Genesis block created for Segment 2!")
	}

	// HTTP Handlers
	http.HandleFunc("/addBlock", addBlockHandler2Post)
	http.HandleFunc("/blockchain", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(Blockchain2)
	})

	// Start the HTTP server
	log.Fatal(http.ListenAndServe(":8090", nil))
}
