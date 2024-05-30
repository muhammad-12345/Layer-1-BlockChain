package node

import (
	"fmt"
	"sync"
	"time"
)

const (
	MaxTransactionListSize = 10
)

// Transaction - Structure for a transaction
type Transaction struct {
	// Transaction details
	ID        string // Unique identifier for the transaction
	Sender    string // ID of the sender node
	Recipient string // ID of the recipient node
	Message   string // Transaction message
	Timestamp time.Time
}

// TransactionList - Manages the list of recent transactions for a node
type TransactionList struct {
	Transactions []*Transaction
	Lock         *sync.RWMutex
}

// NewTransactionList - Creates a new transaction list
func NewTransactionList() *TransactionList {
	return &TransactionList{
		Transactions: make([]*Transaction, 0),
		Lock:         new(sync.RWMutex),
	}
}

// AddTransaction - Adds a transaction to the list, avoiding duplicates
func (tl *TransactionList) AddTransaction(transaction *Transaction) {
	tl.Lock.Lock()
	defer tl.Lock.Unlock()

	// Check for duplicate transaction ID
	for _, existingTransaction := range tl.Transactions {
		if existingTransaction.ID == transaction.ID {
			return // Transaction already in the list
		}
	}

	// Add the transaction to the list
	tl.Transactions = append(tl.Transactions, transaction)

	// Prune the list if it exceeds the maximum size
	if len(tl.Transactions) > MaxTransactionListSize {
		tl.Transactions = tl.Transactions[len(tl.Transactions)-MaxTransactionListSize:]
	}
}

// DisplayTransactions - Displays the list of recent transactions for a node
func (tl *TransactionList) DisplayTransactions(nodeID string) {
	tl.Lock.RLock()
	defer tl.Lock.RUnlock()

	fmt.Printf("[%s] Recent Transactions:\n", nodeID)
	for _, transaction := range tl.Transactions {
		fmt.Printf("[%s] - %s\n", nodeID, transaction.Message)
	}
}

// BroadcastTransaction - Broadcasts a transaction to all neighbors using flooding
func (tl *TransactionList) BroadcastTransaction(sender *Node, neighbors []*Node, message string) {
	// Create a new transaction
	transaction := &Transaction{
		ID:        fmt.Sprintf("%s-%d", sender.ID, time.Now().UnixNano()),
		Sender:    sender.ID,
		Recipient: "", // Modify based on your actual logic
		Message:   message,
		Timestamp: time.Now(),
	}

	// Add the transaction to the sender's list
	tl.AddTransaction(transaction)

	// Broadcast the transaction to all neighbors
	for _, neighbor := range neighbors {
		// Assuming SendMessage is a method in Node to send a message
		sender.SendMessage(neighbor, fmt.Sprintf("TX|%s|%s", transaction.ID, transaction.Message))
	}
}

func GetNeighbors(node *Node) []*Node {
	return node.Neighbors
}
