package main

import (
	node "blockchain/network"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func main() {
	// Initialize the P2P network
	bootstrapNode := &node.Node{
		ID:            "BootstrapNode",
		Address:       "localhost",
		Port:          "9000",
		Bootstrap:     true,
		Joined:        true,
		BootstrapIP:   "",
		BootstrapPort: "",
	}

	node.InitializeNetwork(bootstrapNode)

	fmt.Println("Initial P2P Network State:")
	node.DisplayNetwork()

	var numOfNodes int
	fmt.Scan(&numOfNodes)
	// Create and connect nodes to the network
	basePort := 8000
	var nodes []*node.Node

	for i := 1; i <= numOfNodes; i++ {
		nodeID := "Node" + strconv.Itoa(i)
		nodeAddress := "localhost"
		nodePort := strconv.Itoa(basePort + i)

		newNode := &node.Node{
			ID:            nodeID,
			Address:       nodeAddress,
			Port:          nodePort,
			Bootstrap:     false,
			Joined:        false,
			BootstrapIP:   "",
			BootstrapPort: "",
			Transactions:  node.NewTransactionList(),
		}

		nodes = append(nodes, newNode)

		node.ConnectToNetwork(newNode)
		go newNode.StartServer()
		time.Sleep(time.Millisecond * 100)
	}

	fmt.Println("P2P Network State After Creating 5 Nodes:")
	node.DisplayNetwork()

	var joinNetwork string
	fmt.Print("Do you want to join a new node to the network? (yes/no): ")
	fmt.Scan(&joinNetwork)

	if strings.ToLower(joinNetwork) == "yes" {
		// Join a new node to the network
		newNode := node.CreateNewNode()
		node.ConnectToNetwork(newNode)
		go newNode.StartServer()
		newNode.JoinNetwork(bootstrapNode)
		numOfNodes += 1
	}

	fmt.Println("P2P Network State After Joining a New Node:")
	node.DisplayNetwork()
	node.DisplayJoinedNodes()

	// Simulate nodes sending transactions to each other
	for i := 1; i <= numOfNodes; i++ {
		senderNodeID := "Node" + strconv.Itoa(i)
		recipientNodeID := "Node" + strconv.Itoa((i%5)+1) // Sending to a random node

		senderNode := node.FindNodeByID(senderNodeID)
		recipientNode := node.FindNodeByID(recipientNodeID)

		if senderNode != nil && recipientNode != nil {
			message := fmt.Sprintf("Transaction from %s to %s!", senderNodeID, recipientNodeID)
			senderNode.Transactions.BroadcastTransaction(senderNode, recipientNode.Neighbors, message)
			time.Sleep(time.Millisecond * 100)
		}
	}

	// Display recent transactions for each node
	for _, node := range nodes {
		node.Transactions.DisplayTransactions(node.ID)
	}
}
