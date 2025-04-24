package model

// Reward represents a Tezos reward.
type Reward struct {
	ID               int64         `db:"id" json:"id"`
	RecipientAddress WalletAddress `db:"recipient_address" json:"recipient_address"`
	SourceAddress    WalletAddress `db:"source_address" json:"source_address"`
	Cycle            int           `db:"cycle" json:"cycle"`
	Amount           float64       `db:"amount" json:"amount"`
	Timestamp        int64         `db:"timestamp" json:"-"`
	TimestampTime    string        `db:"-" json:"timestamp"`
}

// RewardsResponse is the response format for the API.
type RewardsResponse struct {
	Rewards []Reward `json:"rewards"`
}
