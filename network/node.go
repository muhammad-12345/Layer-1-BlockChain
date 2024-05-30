package node

import (
	"fmt"
	"net"
	"strings"
)

const (
	RequestGetExistingNodes = "GET_EXISTING_NODES"
	ResponseSeparator       = ","
)

// Node - Defines the node structure
type Node struct {
	ID            string // Unique identifier for each node
	Address       string // IP address or domain of the node
	Port          string // Port number for the node
	Bootstrap     bool   // Indicates whether the node is the bootstrap node
	Joined        bool   // Indicates whether the node has joined the network
	BootstrapIP   string // IP address of the bootstrap node
	BootstrapPort string // Port number of the bootstrap node
	Neighbors     []*Node
	Transactions  *TransactionList // Assuming you have a TransactionList type
}

// StartServer starts the server for the node
func (n *Node) StartServer() {
	listenAddr := n.Address + ":" + n.Port

	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		fmt.Printf("Error starting server for Node %s: %s\n", n.ID, err)
		return
	}
	defer listener.Close()

	fmt.Printf("Node %s listening for incoming connections on %s\n", n.ID, listenAddr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Error accepting connection for Node %s: %s\n", n.ID, err)
			continue
		}

		// Handle the incoming connection in a separate goroutine
		go n.handleIncomingConnection(conn)
	}
}

func CreateNewNode() *Node {
	var nodeID, nodeAddress, nodePort string

	fmt.Print("Enter Node ID: ")
	fmt.Scan(&nodeID)

	fmt.Print("Enter Node IP address: ")
	fmt.Scan(&nodeAddress)

	fmt.Print("Enter Node port: ")
	fmt.Scan(&nodePort)

	newNode := &Node{
		ID:           nodeID,
		Address:      nodeAddress,
		Port:         nodePort,
		Transactions: NewTransactionList(),
	}

	return newNode
}

// handleIncomingConnection handles an incoming connection for the node
func (node *Node) handleIncomingConnection(conn net.Conn) {
	defer conn.Close()

	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Printf("Error reading from connection for Node %s: %s\n", node.ID, err)
		return
	}

	message := string(buffer[:n])

	// Check if the received message is a request for existing nodes
	if message == "GET_EXISTING_NODES" {
		node.sendExistingNodes(conn)
		return
	}

	fmt.Printf("[%s] Received message: %s\n", node.ID, message)
}

func (node *Node) sendExistingNodes(conn net.Conn) {
	globalNetwork.Mutex.RLock()
	defer globalNetwork.Mutex.RUnlock()

	// Build a comma-separated list of nodes (ID:Address:Port)
	var existingNodesList []string
	for _, existingNode := range globalNetwork.Nodes {
		nodeInfo := fmt.Sprintf("%s:%s:%s", existingNode.ID, existingNode.Address, existingNode.Port)
		existingNodesList = append(existingNodesList, nodeInfo)
	}

	// Join the node information and send it as a response
	response := strings.Join(existingNodesList, ",")
	_, err := conn.Write([]byte(response))
	if err != nil {
		fmt.Printf("Error sending existing nodes list to Node %s: %s\n", node.ID, err)
	}
}

func (source *Node) SendMessage(destination *Node, message string) {
	destAddr := destination.Address + ":" + destination.Port

	// Establish a connection to the destination node
	conn, err := net.Dial("tcp", destAddr)
	if err != nil {
		fmt.Printf("Error connecting to Node %s: %s\n", destination.ID, err)
		return
	}
	defer conn.Close()

	// Send the message
	_, err = conn.Write([]byte(message))
	if err != nil {
		fmt.Printf("Error sending message to Node %s: %s\n", destination.ID, err)
		return
	}

	fmt.Printf("[%s] Sent message to [%s]: %s\n", source.ID, destination.ID, message)
}

func FindNodeByID(nodeID string) *Node {
	globalNetwork.Mutex.RLock()
	defer globalNetwork.Mutex.RUnlock()

	for _, node := range globalNetwork.Nodes {
		if node.ID == nodeID {
			return node
		}
	}

	return nil
}

// JoinNetwork - Joining the P2P network
func (n *Node) JoinNetwork(bootstrapNode *Node) {
	fmt.Printf("[%s] Joining the network...\n", n.ID)

	if n.Joined {
		fmt.Printf("[%s] Already joined the network.\n", n.ID)
		return
	}

	// Print bootstrap node information for debugging
	fmt.Printf("[%s] Attempting to connect to Bootstrap Node: %s:%s\n", n.ID, bootstrapNode.Address, bootstrapNode.Port)

	// Contact the bootstrap node to get the list of existing nodes
	existingNodes, err := n.getExistingNodesFromBootstrap(bootstrapNode)
	if err != nil {
		fmt.Printf("[%s] Error joining the network: %s\n", n.ID, err)
		return
	}

	// Establish connections with the existing nodes
	for _, existingNode := range existingNodes {
		err := n.connectToNode(existingNode)
		if err != nil {
			fmt.Printf("[%s] Error connecting to Node %s: %s\n", n.ID, existingNode.ID, err)
			// Handle error as needed (e.g., skip the node and continue)
		}
	}

	fmt.Printf("[%s] Successfully joined the network.\n", n.ID)
	n.Joined = true
}

// getExistingNodesFromBootstrap contacts the bootstrap node to get the list of existing nodes
func (n *Node) getExistingNodesFromBootstrap(bootstrapNode *Node) ([]*Node, error) {
	bootstrapAddr := bootstrapNode.Address + ":" + bootstrapNode.Port

	// Establish a connection to the bootstrap node
	conn, err := net.Dial("tcp", bootstrapAddr)
	if err != nil {
		return nil, fmt.Errorf("error connecting to bootstrap node (%s): %s", bootstrapAddr, err)
	}
	defer conn.Close()

	// Send a request to the bootstrap node to get existing nodes
	_, err = conn.Write([]byte(RequestGetExistingNodes))
	if err != nil {
		return nil, fmt.Errorf("error sending request to bootstrap node: %s", err)
	}

	// Read the response from the bootstrap node
	buffer := make([]byte, 1024)
	nBytes, err := conn.Read(buffer)
	if err != nil {
		return nil, fmt.Errorf("error reading response from bootstrap node: %s", err)
	}

	response := string(buffer[:nBytes])
	existingNodes := parseExistingNodesResponse(response)

	return existingNodes, nil
}

// parseExistingNodesResponse parses the response from the bootstrap node
func parseExistingNodesResponse(response string) []*Node {
	// Assuming the response format is a comma-separated list of nodes (ID:Address:Port)
	nodeStrings := strings.Split(response, ",")

	var existingNodes []*Node
	for _, nodeStr := range nodeStrings {
		nodeInfo := strings.Split(nodeStr, ":")
		if len(nodeInfo) == 3 {
			existingNodes = append(existingNodes, &Node{
				ID:      nodeInfo[0],
				Address: nodeInfo[1],
				Port:    nodeInfo[2],
			})
		}
	}

	return existingNodes
}

// connectToNode establishes a connection with another node
func (n *Node) connectToNode(targetNode *Node) error {
	targetAddr := targetNode.Address + ":" + targetNode.Port

	// Establish a connection to the target node
	conn, err := net.Dial("tcp", targetAddr)
	if err != nil {
		return fmt.Errorf("error connecting to Node %s: %s", targetNode.ID, err)
	}
	defer conn.Close()

	// Perform any additional setup or exchange of information if needed

	return nil
}

// Helper function to generate the response for existing nodes
func generateExistingNodesResponse() string {
	globalNetwork.Mutex.RLock()
	defer globalNetwork.Mutex.RUnlock()

	var nodes []string
	for _, node := range globalNetwork.Nodes {
		nodes = append(nodes, fmt.Sprintf("%s:%s:%s", node.ID, node.Address, node.Port))
	}

	return strings.Join(nodes, ResponseSeparator)
}
