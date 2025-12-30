package node

import (
	"context"
	"testing"

	"distributed-auction/internal/domain"
	pb "distributed-auction/proto"
)

// Mock implementations for testing

type mockLock struct{}

func (m *mockLock) Acquire(ctx context.Context, nodeID string) error { return nil }
func (m *mockLock) Release(ctx context.Context, nodeID string) error { return nil }

type mockReplicator struct {
	called bool
	state  domain.State
}

func (m *mockReplicator) Broadcast(ctx context.Context, state domain.State) {
	m.called = true
	m.state = state
}

// Tests

func TestSubmitBid_Accepted(t *testing.T) {
	repl := &mockReplicator{}
	node := NewNode("test", &mockLock{}, repl)

	state, accepted, err := node.SubmitBid(100, "alice")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !accepted {
		t.Error("bid should be accepted")
	}
	if state.HighestBid != 100 {
		t.Errorf("expected bid 100, got %d", state.HighestBid)
	}
	if state.HighestBidder != "alice" {
		t.Errorf("expected bidder alice, got %s", state.HighestBidder)
	}
	if !repl.called {
		t.Error("replicator should be called")
	}
}

func TestSubmitBid_Rejected(t *testing.T) {
	repl := &mockReplicator{}
	node := NewNode("test", &mockLock{}, repl)

	// First bid
	node.SubmitBid(100, "alice")

	// Lower bid should be rejected
	_, accepted, _ := node.SubmitBid(50, "bob")

	if accepted {
		t.Error("lower bid should be rejected")
	}
}

func TestSubmitBid_LamportIncrements(t *testing.T) {
	node := NewNode("test", &mockLock{}, &mockReplicator{})

	state1, _, _ := node.SubmitBid(100, "alice")
	state2, _, _ := node.SubmitBid(200, "bob")

	if state2.Lamport <= state1.Lamport {
		t.Error("lamport should increment")
	}
}

func TestLamportClock(t *testing.T) {
	node := NewNode("test", &mockLock{}, &mockReplicator{})

	// tick increments
	ts1 := node.tick()
	ts2 := node.tick()
	if ts2 != ts1+1 {
		t.Error("tick should increment by 1")
	}

	// observe with higher value
	node.observe(100)
	ts3 := node.tick()
	if ts3 <= 100 {
		t.Error("observe should sync to higher value")
	}
}

func TestReplicate_EqualLamport_TieBreak(t *testing.T) {
	node := NewNode("test", &mockLock{}, &mockReplicator{})

	node.state = domain.State{
		HighestBid:    300,
		HighestBidder: "bob",
		Lamport:       5,
	}
	node.lamport = 5

	_, err := node.Replicate(context.Background(), &pb.AuctionState{
		HighestBid:    310,
		HighestBidder: "alice",
		Lamport:       5,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := node.GetState()
	if got.HighestBid != 310 || got.HighestBidder != "alice" || got.Lamport != 5 {
		t.Fatalf("expected (310, alice, 5), got (%d, %s, %d)", got.HighestBid, got.HighestBidder, got.Lamport)
	}
}
