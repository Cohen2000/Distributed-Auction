package domain

// State represents the current auction state.
type State struct {
	HighestBid    int64  `json:"highestBid"`
	HighestBidder string `json:"highestBidder"`
	Lamport       int64  `json:"lamport"`
}

// Bid represents an incoming bid from a client.
type Bid struct {
	Amount   int64
	BidderID string
}

// IsBetterBid returns true if the bid beats the current state.
func (s State) IsBetterBid(b Bid) bool {
	if b.Amount != s.HighestBid {
		return b.Amount > s.HighestBid
	}
	if s.HighestBidder == "" {
		return true
	}
	return b.BidderID < s.HighestBidder
}
