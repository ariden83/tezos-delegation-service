package model

import "time"

// StakingOperation represents a staking operation in the Tezos blockchain.
type StakingOperation struct {
	Hash       string        `json:"hash"`
	Type       OperationType `json:"type"`
	Entrypoint string        `json:"entrypoint,omitempty"`
	Amount     float64       `json:"amount"`
	Wallet     WalletAddress `json:"wallet"`
	Baker      WalletAddress `json:"baker,omitempty"`
	Timestamp  time.Time     `json:"timestamp"`
	Status     string        `json:"status"`
}
