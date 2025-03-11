package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

// Block2 represents each 'block' in the blockchain for Segment 2
type Block2 struct {
	Index     int    `json:"index"`
	Timestamp string `json:"timestamp"`
	Data      string `json:"data"`
	Hash      string `json:"hash"`
	PrevHash  string `json:"prev_hash"`
}

// Blockchain for Segment 2
var Blockchain2 []Block2
var mutex2 = &sync.Mutex{}

var db *sql.DB

// Initialize database connection
func initDB() {
	var err error
	db, err = sql.Open("postgres", "host=localhost user=youruser password=yourpassword dbname=blockchain sslmode=disable")
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("Database ping failed:", err)
	}

	fmt.Println("✅ Database connected successfully.")
}

// Create a new block
func CreateBlock2(prevBlock Block2, data string) Block2 {
	block := Block2{
		Index:     prevBlock.Index + 1,
		Timestamp: time.Now().String(),
		Data:      data,
		PrevHash:  prevBlock.Hash,
		Hash:      fmt.Sprintf("%x", time.Now().UnixNano()),
	}
	return block
}

// Genesis Block for Segment 2
func GenesisBlock2() Block2 {
	return Block2{
		Index:     0,
		Timestamp: time.Now().String(),
		Data:      "Genesis Block (Segment 2)",
		Hash:      fmt.Sprintf("%x", time.Now().UnixNano()),
		PrevHash:  "",
	}
}

// Add a block to the blockchain
func AddBlock2(data string) {
	mutex2.Lock()
	defer mutex2.Unlock()

	prevBlock := Blockchain2[len(Blockchain2)-1]
	newBlock := CreateBlock2(prevBlock, data)
	Blockchain2 = append(Blockchain2, newBlock)

	// Store block in PostgreSQL
	_, err := db.Exec("INSERT INTO blocks (index, timestamp, data, hash, prev_hash) VALUES ($1, $2, $3, $4, $5)",
		newBlock.Index, newBlock.Timestamp, newBlock.Data, newBlock.Hash, newBlock.PrevHash)
	if err != nil {
		log.Println("⚠️ Error saving block to database:", err)
	}
}

// Retrieve blockchain from database
func loadBlockchain2() {
	rows, err := db.Query("SELECT index, timestamp, data, hash, prev_hash FROM blocks ORDER BY index")
	if err != nil {
		log.Println("⚠️ Error retrieving blockchain from database:", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var block Block2
		err := rows.Scan(&block.Index, &block.Timestamp, &block.Data, &block.Hash, &block.PrevHash)
		if err != nil {
			log.Println("⚠️ Error scanning blockchain data:", err)
			return
		}
		Blockchain2 = append(Blockchain2, block)
	}
}

// Display blockchain via HTTP
func getBlockchain2(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Blockchain2)
}

// Add a block via HTTP
func addBlockHandler2(w http.ResponseWriter, r *http.Request) {
	data := r.URL.Query().Get("data")
	if data == "" {
		http.Error(w, "Missing 'data' parameter", http.StatusBadRequest)
		return
	}
	AddBlock2(data)
	fmt.Fprintf(w, "✅ Block Added (Segment 2): %s", data)
}

// Sync blockchain with another Segment 2 node
func syncBlockchain2(nodeAddr string) ([]Block2, error) {
	resp, err := http.Get(nodeAddr + "/blockchain")
	if err != nil {
		return nil, fmt.Errorf("Error syncing blockchain from %s: %s", nodeAddr, err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading response from %s: %s", nodeAddr, err)
	}

	var blockchain []Block2
	err = json.Unmarshal(body, &blockchain)
	if err != nil {
		return nil, fmt.Errorf("Error decoding JSON from %s: %s", nodeAddr, err)
	}

	fmt.Printf("✅ Synchronized blockchain from %s\n", nodeAddr)
	return blockchain, nil
}

// Broadcast new block to other Segment 2 nodes
func broadcastNewBlock2(block Block2) {
	nodes := []string{
		"http://localhost:8091",
		"http://localhost:8092",
		"http://localhost:8093",
	}

	for _, node := range nodes {
		_, err := http.PostForm(node+"/add", url.Values{"data": {block.Data}})
		if err != nil {
			fmt.Printf("⚠️ Error broadcasting to %s: %s\n", node, err)
		} else {
			fmt.Printf("✅ Successfully broadcasted new block to %s\n", node)
		}
	}
}

// Check if all nodes in Segment 2 are synchronized
func syncStatusHandler2(w http.ResponseWriter, r *http.Request) {
	nodes := []string{
		"http://localhost:8091",
		"http://localhost:8092",
		"http://localhost:8093",
	}

	var allBlockchains [][]Block2
	for _, node := range nodes {
		blockchain, err := syncBlockchain2(node)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		allBlockchains = append(allBlockchains, blockchain)
	}

	// Compare all blockchains
	for _, blockchain := range allBlockchains {
		if !reflect.DeepEqual(blockchain, Blockchain2) {
			http.Error(w, "⚠️ Blockchains are NOT synchronized across Segment 2.", http.StatusConflict)
			return
		}
	}

	fmt.Fprintf(w, "✅ All nodes in Segment 2 are synchronized.\n")
}

// Start the Segment 2 blockchain network
func main2() {
	// Initialize database and blockchain
	initDB()
	Blockchain2 = append(Blockchain2, GenesisBlock2())
	loadBlockchain2()

	// HTTP Handlers
	http.HandleFunc("/blockchain", getBlockchain2)
	http.HandleFunc("/add", addBlockHandler2)
	http.HandleFunc("/sync-status", syncStatusHandler2)

	// Start HTTP servers for Segment 2 nodes
	go http.ListenAndServe(":8090", nil) // Main Node (Node 5)
	go http.ListenAndServe(":8091", nil) // Node 6
	go http.ListenAndServe(":8092", nil) // Node 7
	go http.ListenAndServe(":8093", nil) // Node 8

	// Start the main node server
	port := ":8090"
	fmt.Printf("🚀 Node 5 (Segment 2) running on http://localhost%s\n", port)
	http.ListenAndServe(port, nil)
}
