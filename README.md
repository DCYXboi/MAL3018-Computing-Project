# MAL3018-Computing-Project

CloudChain: Blockchain-based Data Storage System with AES Encryption
Welcome to CloudChain — a simple, secure, and efficient blockchain-based system built in Go (Golang) for storing encrypted data with AES encryption and SHA-256 hashing, using MongoDB as persistent storage.

📚 Features
✅ Block Creation & Hashing
✅ AES-128 Data Encryption & Decryption
✅ SHA-256 Secure Hash Generation
✅ Persistent Storage with MongoDB
✅ Blockchain Data Retrieval & Synchronization Check
✅ HTTP REST API Interface

📖 How It Works
Each block contains:
Index
Timestamp
Encrypted Data
Hash
Previous Block Hash

Data is encrypted using AES-128 before being stored.
Each block’s integrity is maintained by SHA-256 hashing.
The blockchain is stored both in-memory and in MongoDB.
Validators can retrieve and verify the chain using a simple HTTP API.
Synchronization status is checked by comparing local and MongoDB blockchain lengths.

⚙️ Setup & Run
Prerequisites
Go (Golang)
MongoDB

🙌 Credits
Developed by Danny Chan.
If you like this project — ⭐️ it!
