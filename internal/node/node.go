package node

import (
	"context"
	"log"
	"sync"
	"time"

	"distributed-auction/internal/coordinator"
	"distributed-auction/internal/domain"
	pb "distributed-auction/proto"
)

type Node struct {
	pb.UnimplementedNodeServiceServer

	nodeID string
	lock   coordinator.LockService
	repl   Replicator

	mu      sync.Mutex
	state   domain.State
	lamport int64
}

func NewNode(nodeID string, lock coordinator.LockService, repl Replicator) *Node {
	return &Node{
		nodeID: nodeID,
		lock:   lock,
		repl:   repl,
	}
}

// Lamport clock operations
func (n *Node) tick() int64 {
	n.lamport++
	return n.lamport
}

func (n *Node) observe(remote int64) {
	if remote > n.lamport {
		n.lamport = remote
	}
	n.lamport++
}

func (n *Node) GetState() domain.State {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.state
}

func (n *Node) SubmitBid(amount int64, bidderID string) (domain.State, bool, error) {
	bid := domain.Bid{Amount: amount, BidderID: bidderID}
	log.Printf("[%s] received bid: %d from %s", n.nodeID, amount, bidderID)

	// Acquire lock from coordinator
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := n.lock.Acquire(ctx, n.nodeID); err != nil {
		return domain.State{}, false, err
	}
	defer n.lock.Release(context.Background(), n.nodeID)

	n.mu.Lock()
	defer n.mu.Unlock()

	// Check if bid is better
	if !n.state.IsBetterBid(bid) {
		log.Printf("[%s] rejected: %d <= %d", n.nodeID, amount, n.state.HighestBid)
		return n.state, false, nil
	}

	// Accept: update state with new Lamport version
	ts := n.tick()
	n.state = domain.State{
		HighestBid:    amount,
		HighestBidder: bidderID,
		Lamport:       ts,
	}
	log.Printf("[%s] accepted: %d from %s (lamport=%d)", n.nodeID, amount, bidderID, ts)

	// Broadcast to peers
	n.repl.Broadcast(context.Background(), n.state)

	return n.state, true, nil
}

// Replicate handles incoming state from other nodes.
// Only applies if Lamport timestamp is higher (newer version).
func (n *Node) Replicate(ctx context.Context, req *pb.AuctionState) (*pb.Ack, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.observe(req.Lamport)

	apply := false
	if req.Lamport > n.state.Lamport {
		apply = true
	} else if req.Lamport == n.state.Lamport {
		if req.HighestBid > n.state.HighestBid {
			apply = true
		} else if req.HighestBid == n.state.HighestBid {
			if n.state.HighestBidder == "" || req.HighestBidder < n.state.HighestBidder {
				apply = true
			}
		}
	}

	if apply {
		n.state = domain.State{
			HighestBid:    req.HighestBid,
			HighestBidder: req.HighestBidder,
			Lamport:       req.Lamport,
		}
		log.Printf("[%s] applied: %d from %s (lamport=%d)",
			n.nodeID, req.HighestBid, req.HighestBidder, req.Lamport)
	} else {
		log.Printf("[%s] ignored old state (lamport %d <= %d)",
			n.nodeID, req.Lamport, n.state.Lamport)
	}

	return &pb.Ack{}, nil
}
