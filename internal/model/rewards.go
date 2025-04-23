package model

// Reward represents a Tezos reward.
type Reward struct {
	ID            int64   `db:"id" json:"id"`
	RecipientID   int64   `db:"recipient_id" json:"recipient_id"`
	SourceID      int64   `db:"source_id" json:"source_id"`
	Cycle         int     `db:"cycle" json:"cycle"`
	Amount        float64 `db:"amount" json:"amount"`
	Timestamp     int64   `db:"timestamp" json:"-"`
	TimestampTime string  `db:"-" json:"timestamp"`
}

// RewardsResponse is the response format for the API.
type RewardsResponse struct {
	Rewards []Reward `json:"rewards"`
}
