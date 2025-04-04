package model

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_Delegation(t *testing.T) {
	now := time.Now().Unix()
	delegation := Delegation{
		ID:        1,
		Delegator: "tz1abc",
		Timestamp: now,
		Amount:    100.0,
		Level:     12345,
		CreatedAt: time.Now(),
	}

	assert.Equal(t, int64(1), delegation.ID)
	assert.Equal(t, "tz1abc", delegation.Delegator)
	assert.Equal(t, now, delegation.Timestamp)
	assert.Equal(t, 100.0, delegation.Amount)
	assert.Equal(t, int64(12345), delegation.Level)
}

func Test_DelegationJSON(t *testing.T) {
	now := time.Now().Unix()
	delegation := Delegation{
		ID:        1,
		Delegator: "tz1abc",
		Delegate:  "tz1def",
		Timestamp: now,
		Amount:    100.0,
		Level:     12345,
		CreatedAt: time.Now(),
	}

	jsonData, err := json.Marshal(delegation)
	assert.NoError(t, err)

	var jsonMap map[string]interface{}
	err = json.Unmarshal(jsonData, &jsonMap)
	assert.NoError(t, err)

	assert.Equal(t, "tz1abc", jsonMap["delegator"])
	assert.Equal(t, "tz1def", jsonMap["delegate"])
	assert.Equal(t, 100.0, jsonMap["amount"])
	assert.Equal(t, float64(12345), jsonMap["level"])

	_, hasID := jsonMap["id"]
	assert.False(t, hasID)
	_, hasCreatedAt := jsonMap["created_at"]
	assert.False(t, hasCreatedAt)
}

func Test_DelegationResponse(t *testing.T) {
	now := time.Now().Unix()
	delegations := []Delegation{
		{
			ID:        1,
			Delegator: "tz1abc",
			Timestamp: now,
			Amount:    100.0,
			Level:     12345,
			CreatedAt: time.Now(),
		},
		{
			ID:        2,
			Delegator: "tz1def",
			Timestamp: now - 100,
			Amount:    200.0,
			Level:     12346,
			CreatedAt: time.Now(),
		},
	}

	pagination := PaginationInfo{
		CurrentPage: 1,
		PerPage:     10,
		HasPrevPage: false,
		HasNextPage: false,
	}

	response := DelegationResponse{
		Data:       delegations,
		Pagination: pagination,
	}

	assert.Equal(t, delegations, response.Data)
	assert.Equal(t, pagination, response.Pagination)

	jsonData, err := json.Marshal(response)
	assert.NoError(t, err)

	var parsedResponse DelegationResponse
	err = json.Unmarshal(jsonData, &parsedResponse)
	assert.NoError(t, err)

	assert.Equal(t, len(delegations), len(parsedResponse.Data))
	assert.Equal(t, delegations[0].Delegator, parsedResponse.Data[0].Delegator)
	assert.Equal(t, pagination.CurrentPage, parsedResponse.Pagination.CurrentPage)
}

func Test_PaginationInfo(t *testing.T) {
	pagination := PaginationInfo{
		CurrentPage: 2,
		PerPage:     10,
		HasPrevPage: true,
		HasNextPage: true,
		PrevPage:    1,
		NextPage:    3,
	}

	assert.Equal(t, 2, pagination.CurrentPage)
	assert.Equal(t, 10, pagination.PerPage)
	assert.True(t, pagination.HasPrevPage)
	assert.True(t, pagination.HasNextPage)
	assert.Equal(t, 1, pagination.PrevPage)
	assert.Equal(t, 3, pagination.NextPage)

	jsonData, err := json.Marshal(pagination)
	assert.NoError(t, err)

	var jsonMap map[string]interface{}
	err = json.Unmarshal(jsonData, &jsonMap)
	assert.NoError(t, err)

	assert.Equal(t, float64(2), jsonMap["current_page"])
	assert.Equal(t, float64(10), jsonMap["per_page"])
	assert.Equal(t, true, jsonMap["has_prev_page"])
	assert.Equal(t, true, jsonMap["has_next_page"])
	assert.Equal(t, float64(1), jsonMap["prev_page"])
	assert.Equal(t, float64(3), jsonMap["next_page"])
}
