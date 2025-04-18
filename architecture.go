package main

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
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
const encryptionKey = "examplekey123456" // 16-byte key for AES-128
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

// Function to encrypt data using AES
func encrypt(data, key string) (string, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	plaintext := []byte(data)
	ciphertext := make([]byte, len(plaintext))
	stream := cipher.NewCFBEncrypter(block, []byte(key)[:block.BlockSize()])
	stream.XORKeyStream(ciphertext, plaintext)

	return hex.EncodeToString(ciphertext), nil
}

// Function to decrypt data using AES
func decrypt(encryptedData, key string) (string, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		fmt.Println("Error creating cipher:", err)
		return "", err
	}

	ciphertext, err := hex.DecodeString(encryptedData)
	if err != nil {
		fmt.Println("Error decoding encrypted data:", err)
		return "", err
	}

	plaintext := make([]byte, len(ciphertext))
	stream := cipher.NewCFBDecrypter(block, []byte(key)[:block.BlockSize()])
	stream.XORKeyStream(plaintext, ciphertext)

	return string(plaintext), nil
}

// GetBlockByIndex retrieves a block by its index from MongoDB
func GetBlockByIndex(index int) (*Block, error) {
	var block Block
	err := blockCollection.FindOne(context.TODO(), bson.M{"index": index}).Decode(&block)
	if err != nil {
		return nil, err
	}
	return &block, nil
}

func getBlockByIndexHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the index from the query parameters
	indexParam := r.URL.Query().Get("index")
	if indexParam == "" {
		http.Error(w, "Index parameter is required", http.StatusBadRequest)
		return
	}

	// Convert index to integer
	index, err := strconv.Atoi(indexParam)
	if err != nil {
		http.Error(w, "Invalid index parameter", http.StatusBadRequest)
		return
	}

	// Retrieve the block
	block, err := GetBlockByIndex(index)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to retrieve block: %v", err), http.StatusNotFound)
		return
	}

	// Respond with the block
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(block)
}

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

	// Encrypt the data
	encryptionKey := "examplekey123456" // 16-byte key for AES-128
	encryptedData, err := encrypt(data, encryptionKey)
	if err != nil {
		http.Error(w, "Error encrypting data.", http.StatusInternalServerError)
		return
	}

	// Create a new block
	prevBlock := Blockchain[len(Blockchain)-1]
	newBlock := Block{
		Index:     prevBlock.Index + 1,
		Timestamp: time.Now().String(),
		Data:      encryptedData, // Store encrypted data
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
// SaveBlock stores a new block in MongoDB
func SaveBlock(block Block) {
	fmt.Println("Attempting to save block to MongoDB:", block)
	_, err := blockCollection.InsertOne(context.TODO(), block)
	if err != nil {
		log.Printf("Failed to save block: %v", err)
	} else {
		fmt.Println("Block successfully saved to MongoDB:", block)
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
	encryptionKey := "examplekey123456" // Same key used for encryption

	for _, block := range blocks {
		// Attempt to decrypt the data
		decryptedData, err := decrypt(block.Data, encryptionKey)
		if err != nil {
			fmt.Printf("Error decrypting block data for Index %d: %v\n", block.Index, err)
			decryptedData = "[Unable to decrypt data]"
		}

		// Print the block with decrypted data
		fmt.Fprintf(w, "Index: %d\nTimestamp: %s\nData: %s\nHash: %s\nPrevHash: %s\n\n",
			block.Index, block.Timestamp, decryptedData, block.Hash, block.PrevHash)
	}
}

func isNodeSynced() bool {
	// Count the number of blocks in the MongoDB collection
	count, err := blockCollection.CountDocuments(context.TODO(), bson.D{})
	if err != nil {
		log.Printf("Error checking sync status: %v", err)
		return false
	}

	// Compare the count with the local blockchain length
	return int(count) == len(Blockchain)
}

// HTTP Handler to Check Sync Status
func syncStatusHandler(w http.ResponseWriter, r *http.Request) {
	if isNodeSynced() {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "Node is synced")
	} else {
		w.WriteHeader(http.StatusConflict)
		fmt.Fprintln(w, "Node is out of sync")
	}
}

// HTTP Handler to Add a New Block
func addBlock(w http.ResponseWriter, r *http.Request) {
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

	// Save the block to MongoDB
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

func main() {
	// Connect to MongoDB
	connectMongoDB()

	// Load existing blockchain from database
	Blockchain = LoadBlockchain()

	// Check if the blockchain is empty and create a genesis block if necessary
	if len(Blockchain) == 0 {
		genesisBlock := Block{
			Index:     0,
			Timestamp: time.Now().String(),
			Data:      "Genesis Block",
			Hash:      calculateHash(Block{Index: 0, Timestamp: time.Now().String(), Data: "Genesis Block", PrevHash: ""}),
			PrevHash:  "",
		}
		Blockchain = append(Blockchain, genesisBlock)
		SaveBlock(genesisBlock)
		fmt.Println("Genesis block created!")
	}

	// Check and log the sync status
	if isNodeSynced() {
		fmt.Println("Node is synced")
	} else {
		fmt.Println("Node is out of sync")
	}

	http.HandleFunc("/getBlockByIndex", getBlockByIndexHandler)
	// HTTP Handlers
	http.HandleFunc("/blockchain", getBlockchain)
	http.HandleFunc("/addBlock", addBlock)
	http.HandleFunc("/sync-status", syncStatusHandler) // Register sync status endpoint

	// Start the HTTP server
	fmt.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
