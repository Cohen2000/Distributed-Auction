package domain

import "testing"

func TestIsBetterBid_HigherAmountWins(t *testing.T) {
	state := State{HighestBid: 100, HighestBidder: "alice"}
	bid := Bid{Amount: 150, BidderID: "bob"}

	if !state.IsBetterBid(bid) {
		t.Error("higher bid should win")
	}
}

func TestIsBetterBid_LowerAmountLoses(t *testing.T) {
	state := State{HighestBid: 100, HighestBidder: "alice"}
	bid := Bid{Amount: 50, BidderID: "bob"}

	if state.IsBetterBid(bid) {
		t.Error("lower bid should lose")
	}
}

func TestIsBetterBid_EqualAmount_TieBreak(t *testing.T) {
	state := State{HighestBid: 100, HighestBidder: "bob"}

	bid1 := Bid{Amount: 100, BidderID: "alice"}
	if !state.IsBetterBid(bid1) {
		t.Error("equal amount with smaller bidderId should win")
	}

	bid2 := Bid{Amount: 100, BidderID: "carol"}
	if state.IsBetterBid(bid2) {
		t.Error("equal amount with larger bidderId should lose")
	}
}
