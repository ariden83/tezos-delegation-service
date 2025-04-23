package model

import "time"

// Delegation represents a Tezos delegation.
type Delegation struct {
	ID            int64     `db:"id" json:"-"`
	Delegator     string    `db:"delegator" json:"delegator"`
	Delegate      string    `db:"delegate" json:"delegate"`
	Timestamp     int64     `db:"timestamp" json:"-"`
	TimestampTime string    `db:"-" json:"timestamp"`
	Amount        float64   `db:"amount" json:"amount"`
	Level         int64     `db:"level" json:"level"`
	CreatedAt     time.Time `db:"created_at" json:"-"`
}

// PaginationInfo contains pagination metadata.
type PaginationInfo struct {
	CurrentPage int  `json:"current_page"`
	PerPage     int  `json:"per_page"`
	HasPrevPage bool `json:"has_prev_page"`
	HasNextPage bool `json:"has_next_page"`
	PrevPage    int  `json:"prev_page,omitempty"`
	NextPage    int  `json:"next_page,omitempty"`
}

// DelegationsResponse is the response format for the API.
type DelegationsResponse struct {
	Data            []Delegation   `json:"data"`
	Pagination      PaginationInfo `json:"-,omitempty"`
	MaxDelegationID int64          `json:"-"`
}
