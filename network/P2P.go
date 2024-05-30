package node

import (
	"fmt"
	"sync"
)

// Network - Represents the P2P network
type Network struct {
	Nodes []*Node       // List of nodes in the network
	Mutex *sync.RWMutex // Mutex for concurrent access to Nodes
}

// globalNetwork - A global instance of the network
var globalNetwork *Network

// Function in node.go
// InitializeNetwork - Initializes the P2P network
func InitializeNetwork(bootstrapNode *Node) {
	globalNetwork = &Network{
		Nodes: make([]*Node, 0),
		Mutex: new(sync.RWMutex),
	}

	// Add the bootstrap node to the network
	globalNetwork.Nodes = append(globalNetwork.Nodes, bootstrapNode)
	fmt.Printf("Bootstrap Node %s initialized\n", bootstrapNode.ID)
}

// ConnectToNetwork - Connects a node to the network
func ConnectToNetwork(node *Node) {
	globalNetwork.Mutex.Lock()
	defer globalNetwork.Mutex.Unlock()

	globalNetwork.Nodes = append(globalNetwork.Nodes, node)
	fmt.Printf("Node %s connected to the network\n", node.ID)

	// Update bootstrap node's IP and port if the connecting node is not the bootstrap node
	if !node.Bootstrap {
		globalNetwork.Nodes[0].BootstrapIP = node.Address
		globalNetwork.Nodes[0].BootstrapPort = node.Port
	}

	// Initialize neighbors
	node.Neighbors = make([]*Node, 0)
	for _, existingNode := range globalNetwork.Nodes {
		if existingNode != node {
			node.Neighbors = append(node.Neighbors, existingNode)
			existingNode.Neighbors = append(existingNode.Neighbors, node)
		}
	}
}

func DisplayNetwork() {
	globalNetwork.Mutex.RLock()
	defer globalNetwork.Mutex.RUnlock()

	fmt.Println("Current Nodes in the Network:")
	for _, node := range globalNetwork.Nodes {
		fmt.Printf("Node ID: %s, Address: %s:%s\n", node.ID, node.Address, node.Port)
	}
}

func DisplayJoinedNodes() {
	globalNetwork.Mutex.RLock()
	defer globalNetwork.Mutex.RUnlock()

	fmt.Println("Joined Nodes in the Network:")
	for _, node := range globalNetwork.Nodes {
		if node.Joined {
			fmt.Printf("Node ID: %s, Address: %s:%s\n", node.ID, node.Address, node.Port)
		}
	}
}
