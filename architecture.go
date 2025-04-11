package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
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

// Blockchain is the in-memory representation of the blockchain
var Blockchain []Block

// Block represents each 'block' in the blockchain
type Block struct {
	Index     int    `bson:"index"`
	Timestamp string `bson:"timestamp"`
	Data      string `bson:"data"`
	Hash      string `bson:"hash"`
	PrevHash  string `bson:"prev_hash"`
}

// Function to calculate SHA-256 hash for a block
func calculateHash(block Block) string {
	record := fmt.Sprintf("%d%s%s%s", block.Index, block.Timestamp, block.Data, block.PrevHash)
	hash := sha256.Sum256([]byte(record))
	return hex.EncodeToString(hash[:])
}

// HTTP Handler to Add a New Block
func addBlock(w http.ResponseWriter, r *http.Request) {
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
	prevBlock := Blockchain[len(Blockchain)-1]
	newBlock := Block{
		Index:     prevBlock.Index + 1,
		Timestamp: time.Now().String(),
		Data:      data,
		PrevHash:  prevBlock.Hash,
	}
	newBlock.Hash = calculateHash(newBlock)

	// Append the new block to the blockchain
	Blockchain = append(Blockchain, newBlock)
	SaveBlock(newBlock)

	// Respond with success
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Block added successfully: %+v\n", newBlock)
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

// HTTP Handler to Add a New Block
func addBlock1(w http.ResponseWriter, r *http.Request) {
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
	prevBlock := Blockchain[len(Blockchain)-1]
	newBlock := Block{
		Index:     prevBlock.Index + 1,
		Timestamp: time.Now().String(),
		Data:      data,
		Hash:      fmt.Sprintf("%x", time.Now().UnixNano()),
		PrevHash:  prevBlock.Hash,
	}

	// Append the new block to the blockchain
	Blockchain = append(Blockchain, newBlock)
	SaveBlock(newBlock)

	// Respond with success
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Block added successfully: %+v\n", newBlock)
}

// Initialize the blockchain with a genesis block
func initBlockchain() {
	genesisBlock := Block{
		Index:     0,
		Timestamp: time.Now().String(),
		Data:      "Genesis Block",
		Hash:      calculateHash(Block{Index: 0, Timestamp: time.Now().String(), Data: "Genesis Block", PrevHash: ""}),
		PrevHash:  "",
	}
	Blockchain = append(Blockchain, genesisBlock)
	fmt.Println("Genesis block created!")
}

// Main Function
func main() {
	// Initialize the blockchain
	initBlockchain()

	// Connect to MongoDB
	Blockchain = LoadBlockchain()

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

	// Start the HTTP server
	http.HandleFunc("/addBlock", addBlock)
	http.HandleFunc("/blockchain", getBlockchain)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
