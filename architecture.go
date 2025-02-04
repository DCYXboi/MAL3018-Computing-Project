package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"sync"
	"time"
)

// Block represents each 'block' in the blockchain
type Block struct {
	Index     int
	Timestamp string
	Data      string
	Hash      string
	PrevHash  string
}

// Blockchain is a series of validated Block
var Blockchain []Block
var mutex = &sync.Mutex{}

// Node represents each node in the blockchain network
type Node struct {
	ID   int
	Addr string
}

// CreateBlock generates a new block using previous block's hash
func CreateBlock(prevBlock Block, data string) Block {
	block := Block{
		Index:     prevBlock.Index + 1,
		Timestamp: time.Now().String(),
		Data:      data,
		PrevHash:  prevBlock.Hash,
		Hash:      fmt.Sprintf("%x", time.Now().UnixNano()),
	}
	return block
}

// Initialize Blockchain with Genesis Block
func GenesisBlock() Block {
	return Block{
		Index:     0,
		Timestamp: time.Now().String(),
		Data:      "Genesis Block",
		Hash:      fmt.Sprintf("%x", time.Now().UnixNano()),
		PrevHash:  "",
	}
}

// AddBlock adds a new block to the Blockchain
func AddBlock(data string) {
	mutex.Lock()
	defer mutex.Unlock()

	prevBlock := Blockchain[len(Blockchain)-1]
	newBlock := CreateBlock(prevBlock, data)
	Blockchain = append(Blockchain, newBlock)

	// Broadcast new block to other nodes
	broadcastNewBlock(newBlock)
}

// HTTP Handler to Display Blockchain
func getBlockchain(w http.ResponseWriter, r *http.Request) {
	for _, block := range Blockchain {
		fmt.Fprintf(w, "Index: %d\nTimestamp: %s\nData: %s\nHash: %s\nPrevHash: %s\n\n",
			block.Index, block.Timestamp, block.Data, block.Hash, block.PrevHash)
	}
}

// HTTP Handler to Add a Block
func addBlockHandler(w http.ResponseWriter, r *http.Request) {
	data := r.URL.Query().Get("data")
	if data == "" {
		http.Error(w, "Missing 'data' parameter", http.StatusBadRequest)
		return
	}
	AddBlock(data)
	fmt.Fprintf(w, "Block Added: %s", data)
}

// SyncBlockchain synchronizes blockchain data with another node
func syncBlockchain(node Node) ([]Block, error) {
	resp, err := http.Get(node.Addr + "/blockchain")
	if err != nil {
		return nil, fmt.Errorf("Error syncing blockchain with node %d: %s", node.ID, err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading response from node %d: %s", node.ID, err)
	}

	var blockchain []Block
	// Normally, you would decode the body into blockchain data here (JSON, for example)
	// But for simplicity, we assume the blockchain is text, and we'll just simulate it
	fmt.Printf("Synchronized blockchain from Node %d: %s\n", node.ID, string(body))
	return blockchain, nil
}

// Broadcast new block to other nodes
func broadcastNewBlock(block Block) {
	nodes := []Node{
		{ID: 2, Addr: "http://localhost:8081"},
		{ID: 3, Addr: "http://localhost:8082"},
		{ID: 4, Addr: "http://localhost:8083"},
	}

	for _, node := range nodes {
		// Send a POST request to add the block to the other node's blockchain
		_, err := http.PostForm(node.Addr+"/add", url.Values{"data": {block.Data}})
		if err != nil {
			fmt.Printf("Error broadcasting to node %d: %s\n", node.ID, err)
		} else {
			fmt.Printf("Successfully broadcasted new block to node %d\n", node.ID)
		}
	}
}

// Check if all nodes' blockchains are synchronized
func syncStatusHandler(w http.ResponseWriter, r *http.Request) {
	nodes := []Node{
		{ID: 2, Addr: "http://localhost:8081"},
		{ID: 3, Addr: "http://localhost:8082"},
		{ID: 4, Addr: "http://localhost:8083"},
	}

	// Fetch the blockchain data from all nodes
	var allBlockchains [][]Block
	for _, node := range nodes {
		blockchain, err := syncBlockchain(node)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		allBlockchains = append(allBlockchains, blockchain)
	}

	// Compare all blockchains to check if they are the same
	for _, blockchain := range allBlockchains {
		if !reflect.DeepEqual(blockchain, Blockchain) {
			http.Error(w, "Blockchains are not synchronized across all nodes.", http.StatusConflict)
			return
		}
	}

	fmt.Fprintf(w, "All nodes are synchronized.\n")
}

func main() {
	// Initialize Blockchain with Genesis Block
	Blockchain = append(Blockchain, GenesisBlock())

	// HTTP Handlers for each node
	http.HandleFunc("/blockchain", getBlockchain)
	http.HandleFunc("/add", addBlockHandler)
	http.HandleFunc("/sync-status", syncStatusHandler)

	// Start HTTP servers for nodes 1, 2, 3, 4
	go http.ListenAndServe(":8080", nil)
	go http.ListenAndServe(":8081", nil)
	go http.ListenAndServe(":8082", nil)
	go http.ListenAndServe(":8083", nil)

	// Simulate syncing process
	nodes := []Node{
		{ID: 2, Addr: "http://localhost:8081"},
		{ID: 3, Addr: "http://localhost:8082"},
		{ID: 4, Addr: "http://localhost:8083"},
	}

	for _, node := range nodes {
		go syncBlockchain(node)
	}

	// Start server for Node 1
	port := ":8080"
	fmt.Printf("Node 1 is running on http://localhost%s\n", port)
	http.ListenAndServe(port, nil)
}
