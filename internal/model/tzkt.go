package model

import (
	"time"
)

// TzktDelegationResponse represents the structure of the TzKT API response.
type TzktDelegationResponse []TzktDelegation

// TzktDelegation represents a delegation from the TzKT API.
type TzktDelegation struct {
	Type         string           `json:"type"`
	ID           int64            `json:"id"`
	Level        int64            `json:"level"`
	Timestamp    time.Time        `json:"timestamp"`
	Block        string           `json:"block"`
	Hash         string           `json:"hash"`
	Counter      int64            `json:"counter"`
	Sender       TzktAddress      `json:"sender"`
	GasLimit     int64            `json:"gasLimit"`
	GasUsed      int64            `json:"gasUsed"`
	BakerFee     int64            `json:"bakerFee"`
	Amount       int64            `json:"amount"`
	Delegate     TzktDelegate     `json:"newDelegate"`
	PrevDelegate *TzktDelegate    `json:"prevDelegate,omitempty"`
	Status       string           `json:"status"`
	Errors       []TzktError      `json:"errors,omitempty"`
	Originated   []TzktOriginated `json:"originated,omitempty"`
}

// TzktAddress represents an address in the TzKT API.
type TzktAddress struct {
	Address string `json:"address"`
	Alias   string `json:"alias,omitempty"`
}

// TzktDelegate represents a delegate in the TzKT API.
type TzktDelegate struct {
	Address string `json:"address"`
	Alias   string `json:"alias,omitempty"`
}

// TzktError represents an error in the TzKT API.
type TzktError struct {
	Type string `json:"type"`
}

// TzktOriginated represents an originated contract in the TzKT API.
type TzktOriginated struct {
	Address  string   `json:"address"`
	TypeHash int64    `json:"typeHash"`
	CodeHash int64    `json:"codeHash"`
	Tzips    []string `json:"tzips"`
}
