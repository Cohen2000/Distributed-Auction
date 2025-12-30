package main

import (
	"encoding/json"
	"flag"
	"log"
	"net"
	"net/http"
	"strings"

	"distributed-auction/internal/coordinator"
	"distributed-auction/internal/node"
	pb "distributed-auction/proto"

	"github.com/gorilla/mux"
	"google.golang.org/grpc"
)

// Node process: runs a REST API for clients and a gRPC server for inter-node replication.
// State updates are locally serialized via the coordinator lock, then broadcast to peers.
func main() {
	nodeID := flag.String("id", "node-A", "node ID")
	coordAddr := flag.String("coord", "localhost:7000", "coordinator address")
	restAddr := flag.String("rest", ":8080", "REST address")
	grpcAddr := flag.String("grpc", ":9090", "gRPC address")
	peersFlag := flag.String("peers", "", "comma-separated peer addresses")
	flag.Parse()

	var peers []string
	if *peersFlag != "" {
		peers = strings.Split(*peersFlag, ",")
	}

	lockClient, err := coordinator.NewClient(*coordAddr)
	if err != nil {
		log.Fatalf("connect to coordinator: %v", err)
	}

	repl := node.NewReplicator(*nodeID, peers)
	n := node.NewNode(*nodeID, lockClient, repl)

	go startGRPC(*grpcAddr, n)
	startREST(*restAddr, *nodeID, n)
}

// startGRPC starts the gRPC server used by peers to replicate auction state.
func startGRPC(addr string, n *node.Node) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("grpc listen: %v", err)
	}
	server := grpc.NewServer()
	pb.RegisterNodeServiceServer(server, n)
	log.Printf("gRPC on %s", addr)
	server.Serve(lis)
}

// startREST starts the client-facing REST API.
// - GET  /state returns the current local view of the auction state.
// - POST /bid applies a bid locally (if it beats the current highest bid) and triggers replication.
// The "applied" flag indicates whether this node applied the bid locally; the final converged state
// can be observed by querying /state after replication.
func startREST(addr, nodeID string, n *node.Node) {
	r := mux.NewRouter()

	r.HandleFunc("/state", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(n.GetState())
	}).Methods("GET")

	r.HandleFunc("/bid", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Amount   int64  `json:"amount"`
			BidderID string `json:"bidderId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		if req.Amount <= 0 || req.BidderID == "" {
			http.Error(w, "invalid bid", http.StatusBadRequest)
			return
		}

		state, applied, err := n.SubmitBid(req.Amount, req.BidderID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"applied": applied,
			"state":   state,
		})
	}).Methods("POST")

	log.Printf("[%s] REST on %s", nodeID, addr)
	http.ListenAndServe(addr, r)
}
