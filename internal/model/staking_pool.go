package model

// StakingPool represents a staking pool in the Tezos blockchain.
type StakingPool struct {
	ID           int64         `json:"id"`
	Address      WalletAddress `json:"address"`
	Name         string        `json:"name"`
	StakingToken string        `json:"staking_token"`
	CreatedAt    string        `json:"created_at"`
	UpdatedAt    string        `json:"updated_at"`
}
