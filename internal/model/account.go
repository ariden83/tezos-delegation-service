package model

import "time"

// WalletAddress represents a wallet address.
type WalletAddress string

// IsValid checks if the wallet address is valid.
func (w WalletAddress) IsValid() bool {
	if len(w) == 0 {
		return false
	}

	prefixes := []string{"tz1", "tz2", "tz3", "KT1"}
	validPrefix := false
	for _, prefix := range prefixes {
		if len(w) >= len(prefix) && w[:len(prefix)] == WalletAddress(prefix) {
			validPrefix = true
			break
		}
	}

	return validPrefix && (len(w) == 36)
}

// String returns the string representation of the wallet address.
func (w WalletAddress) String() string {
	return string(w)
}

// AccountType represents the type of account.
type AccountType string

const (
	AccountTypeDelegate AccountType = "delegate"
	AccountTypeUser     AccountType = "user"
)

// IsValid checks if the account type is valid.
func (a AccountType) IsValid() bool {
	return a == AccountTypeDelegate || a == AccountTypeUser
}

// String returns the string representation of the account type.
func (a AccountType) String() string {
	return string(a)
}

// Account represents a user account in the Tezos delegation service.
type Account struct {
	ID        int64         `db:"id" json:"id"`
	Address   WalletAddress `db:"address" json:"address"`
	Alias     string        `db:"alias" json:"alias"`
	Type      AccountType   `db:"type" json:"type"`
	CreatedAt time.Time     `db:"created_at" json:"-"`
}
