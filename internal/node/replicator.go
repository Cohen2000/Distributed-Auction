package node

import (
	"context"
	"log"
	"time"

	"distributed-auction/internal/domain"
	pb "distributed-auction/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Replicator interface for dependency injection.
type Replicator interface {
	Broadcast(ctx context.Context, state domain.State)
}

// GRPCReplicator broadcasts state to peer nodes.
type GRPCReplicator struct {
	nodeID string
	peers  []string
}

func NewReplicator(nodeID string, peers []string) *GRPCReplicator {
	return &GRPCReplicator{nodeID: nodeID, peers: peers}
}

func (r *GRPCReplicator) Broadcast(ctx context.Context, state domain.State) {
	for _, peer := range r.peers {
		go r.sendToPeer(peer, state)
	}
}

func (r *GRPCReplicator) sendToPeer(addr string, state domain.State) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("[%s] connect to %s failed: %v", r.nodeID, addr, err)
		return
	}
	defer conn.Close()

	client := pb.NewNodeServiceClient(conn)
	_, err = client.Replicate(ctx, &pb.AuctionState{
		HighestBid:    state.HighestBid,
		HighestBidder: state.HighestBidder,
		Lamport:       state.Lamport,
	})
	if err != nil {
		log.Printf("[%s] replicate to %s failed: %v", r.nodeID, addr, err)
		return
	}
	log.Printf("[%s] replicated to %s", r.nodeID, addr)
}
