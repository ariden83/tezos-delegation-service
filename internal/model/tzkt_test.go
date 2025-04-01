package model

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTzktDelegation(t *testing.T) {
	now := time.Now()
	delegation := TzktDelegation{
		Type:      "delegation",
		ID:        12345,
		Level:     67890,
		Timestamp: now,
		Block:     "BLjrTQnYs6WKdJubK5b5U1oblX4tRFcJZ5iL8N8yNFSoyz92Gty",
		Hash:      "op2g1rSKTHKhNjk3x8EypCKFRJaBKFNKH3Nj1SbXbdKF9Nfw9EK",
		Counter:   123,
		Sender: TzktAddress{
			Address: "tz1abc",
			Alias:   "Alice",
		},
		GasLimit: 10000,
		GasUsed:  5000,
		BakerFee: 1000,
		Amount:   50000000,
		Delegate: TzktDelegate{
			Address: "tz1def",
			Alias:   "Bob's Bakery",
		},
		PrevDelegate: &TzktDelegate{
			Address: "tz1ghi",
			Alias:   "Carol's Cakes",
		},
		Status: "applied",
		Errors: []TzktError{
			{Type: "temporary"},
		},
		Originated: []TzktOriginated{
			{
				Address:  "KT1abc",
				TypeHash: 123456,
				CodeHash: 789012,
				Tzips:    []string{"FA1.2", "FA2"},
			},
		},
	}

	// Verify fields
	assert.Equal(t, "delegation", delegation.Type)
	assert.Equal(t, int64(12345), delegation.ID)
	assert.Equal(t, int64(67890), delegation.Level)
	assert.Equal(t, now, delegation.Timestamp)
	assert.Equal(t, "BLjrTQnYs6WKdJubK5b5U1oblX4tRFcJZ5iL8N8yNFSoyz92Gty", delegation.Block)
	assert.Equal(t, "op2g1rSKTHKhNjk3x8EypCKFRJaBKFNKH3Nj1SbXbdKF9Nfw9EK", delegation.Hash)
	assert.Equal(t, int64(123), delegation.Counter)
	assert.Equal(t, "tz1abc", delegation.Sender.Address)
	assert.Equal(t, "Alice", delegation.Sender.Alias)
	assert.Equal(t, int64(10000), delegation.GasLimit)
	assert.Equal(t, int64(5000), delegation.GasUsed)
	assert.Equal(t, int64(1000), delegation.BakerFee)
	assert.Equal(t, int64(50000000), delegation.Amount)
	assert.Equal(t, "tz1def", delegation.Delegate.Address)
	assert.Equal(t, "Bob's Bakery", delegation.Delegate.Alias)
	assert.NotNil(t, delegation.PrevDelegate)
	assert.Equal(t, "tz1ghi", delegation.PrevDelegate.Address)
	assert.Equal(t, "Carol's Cakes", delegation.PrevDelegate.Alias)
	assert.Equal(t, "applied", delegation.Status)
	assert.Len(t, delegation.Errors, 1)
	assert.Equal(t, "temporary", delegation.Errors[0].Type)
	assert.Len(t, delegation.Originated, 1)
	assert.Equal(t, "KT1abc", delegation.Originated[0].Address)
	assert.Equal(t, int64(123456), delegation.Originated[0].TypeHash)
	assert.Equal(t, int64(789012), delegation.Originated[0].CodeHash)
	assert.Equal(t, []string{"FA1.2", "FA2"}, delegation.Originated[0].Tzips)
}

func Test_TzktDelegationJSON(t *testing.T) {
	jsonData := `{
		"type": "delegation",
		"id": 12345,
		"level": 67890,
		"timestamp": "2023-01-01T12:00:00Z",
		"block": "BLjrTQnYs6WKdJubK5b5U1oblX4tRFcJZ5iL8N8yNFSoyz92Gty",
		"hash": "op2g1rSKTHKhNjk3x8EypCKFRJaBKFNKH3Nj1SbXbdKF9Nfw9EK",
		"counter": 123,
		"sender": {
			"address": "tz1abc",
			"alias": "Alice"
		},
		"gasLimit": 10000,
		"gasUsed": 5000,
		"bakerFee": 1000,
		"amount": 50000000,
		"newDelegate": {
			"address": "tz1def",
			"alias": "Bob's Bakery"
		},
		"prevDelegate": {
			"address": "tz1ghi",
			"alias": "Carol's Cakes"
		},
		"status": "applied",
		"errors": [
			{"type": "temporary"}
		],
		"originated": [
			{
				"address": "KT1abc",
				"typeHash": 123456,
				"codeHash": 789012,
				"tzips": ["FA1.2", "FA2"]
			}
		]
	}`

	var delegation TzktDelegation
	err := json.Unmarshal([]byte(jsonData), &delegation)
	assert.NoError(t, err)

	expectedTime, _ := time.Parse(time.RFC3339, "2023-01-01T12:00:00Z")
	assert.Equal(t, "delegation", delegation.Type)
	assert.Equal(t, int64(12345), delegation.ID)
	assert.Equal(t, int64(67890), delegation.Level)
	assert.Equal(t, expectedTime, delegation.Timestamp)
	assert.Equal(t, "BLjrTQnYs6WKdJubK5b5U1oblX4tRFcJZ5iL8N8yNFSoyz92Gty", delegation.Block)
	assert.Equal(t, "op2g1rSKTHKhNjk3x8EypCKFRJaBKFNKH3Nj1SbXbdKF9Nfw9EK", delegation.Hash)
	assert.Equal(t, int64(123), delegation.Counter)
	assert.Equal(t, "tz1abc", delegation.Sender.Address)
	assert.Equal(t, "Alice", delegation.Sender.Alias)
	assert.Equal(t, int64(10000), delegation.GasLimit)
	assert.Equal(t, int64(5000), delegation.GasUsed)
	assert.Equal(t, int64(1000), delegation.BakerFee)
	assert.Equal(t, int64(50000000), delegation.Amount)
	assert.Equal(t, "tz1def", delegation.Delegate.Address)
	assert.Equal(t, "Bob's Bakery", delegation.Delegate.Alias)
	assert.NotNil(t, delegation.PrevDelegate)
	assert.Equal(t, "tz1ghi", delegation.PrevDelegate.Address)
	assert.Equal(t, "Carol's Cakes", delegation.PrevDelegate.Alias)
	assert.Equal(t, "applied", delegation.Status)
	assert.Len(t, delegation.Errors, 1)
	assert.Equal(t, "temporary", delegation.Errors[0].Type)
	assert.Len(t, delegation.Originated, 1)
	assert.Equal(t, "KT1abc", delegation.Originated[0].Address)
	assert.Equal(t, int64(123456), delegation.Originated[0].TypeHash)
	assert.Equal(t, int64(789012), delegation.Originated[0].CodeHash)
	assert.Equal(t, []string{"FA1.2", "FA2"}, delegation.Originated[0].Tzips)
}

func Test_TzktAddress(t *testing.T) {
	address := TzktAddress{
		Address: "tz1abc",
		Alias:   "Alice",
	}

	assert.Equal(t, "tz1abc", address.Address)
	assert.Equal(t, "Alice", address.Alias)

	jsonData, err := json.Marshal(address)
	assert.NoError(t, err)

	var parsedAddress TzktAddress
	err = json.Unmarshal(jsonData, &parsedAddress)
	assert.NoError(t, err)

	assert.Equal(t, address.Address, parsedAddress.Address)
	assert.Equal(t, address.Alias, parsedAddress.Alias)
}

func Test_TzktDelegate(t *testing.T) {
	delegate := TzktDelegate{
		Address: "tz1def",
		Alias:   "Bob's Bakery",
	}

	assert.Equal(t, "tz1def", delegate.Address)
	assert.Equal(t, "Bob's Bakery", delegate.Alias)

	jsonData, err := json.Marshal(delegate)
	assert.NoError(t, err)

	var parsedDelegate TzktDelegate
	err = json.Unmarshal(jsonData, &parsedDelegate)
	assert.NoError(t, err)

	assert.Equal(t, delegate.Address, parsedDelegate.Address)
	assert.Equal(t, delegate.Alias, parsedDelegate.Alias)
}

func Test_TzktError(t *testing.T) {
	errorObj := TzktError{
		Type: "temporary",
	}

	assert.Equal(t, "temporary", errorObj.Type)

	jsonData, err := json.Marshal(errorObj)
	assert.NoError(t, err)

	var parsedError TzktError
	err = json.Unmarshal(jsonData, &parsedError)
	assert.NoError(t, err)

	assert.Equal(t, errorObj.Type, parsedError.Type)
}

func Test_TzktOriginated(t *testing.T) {
	originated := TzktOriginated{
		Address:  "KT1abc",
		TypeHash: 123456,
		CodeHash: 789012,
		Tzips:    []string{"FA1.2", "FA2"},
	}

	assert.Equal(t, "KT1abc", originated.Address)
	assert.Equal(t, int64(123456), originated.TypeHash)
	assert.Equal(t, int64(789012), originated.CodeHash)
	assert.Equal(t, []string{"FA1.2", "FA2"}, originated.Tzips)

	jsonData, err := json.Marshal(originated)
	assert.NoError(t, err)

	var parsedOriginated TzktOriginated
	err = json.Unmarshal(jsonData, &parsedOriginated)
	assert.NoError(t, err)

	assert.Equal(t, originated.Address, parsedOriginated.Address)
	assert.Equal(t, originated.TypeHash, parsedOriginated.TypeHash)
	assert.Equal(t, originated.CodeHash, parsedOriginated.CodeHash)
	assert.Equal(t, originated.Tzips, parsedOriginated.Tzips)
}

func Test_TzktDelegationResponse(t *testing.T) {
	now := time.Now()
	response := TzktDelegationResponse{
		{
			Type:      "delegation",
			ID:        12345,
			Level:     67890,
			Timestamp: now,
			Status:    "applied",
			Amount:    50000000,
		},
		{
			Type:      "delegation",
			ID:        12346,
			Level:     67891,
			Timestamp: now.Add(time.Hour),
			Status:    "failed",
			Amount:    25000000,
		},
	}

	assert.Len(t, response, 2)
	assert.Equal(t, int64(12345), response[0].ID)
	assert.Equal(t, int64(12346), response[1].ID)
	assert.Equal(t, "applied", response[0].Status)
	assert.Equal(t, "failed", response[1].Status)

	jsonData, err := json.Marshal(response)
	assert.NoError(t, err)

	var parsedResponse TzktDelegationResponse
	err = json.Unmarshal(jsonData, &parsedResponse)
	assert.NoError(t, err)

	assert.Len(t, parsedResponse, 2)
	assert.Equal(t, response[0].ID, parsedResponse[0].ID)
	assert.Equal(t, response[1].ID, parsedResponse[1].ID)
}
