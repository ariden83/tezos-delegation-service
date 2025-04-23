package usecase

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/tezos-delegation-service/internal/adapter/database"
	"github.com/tezos-delegation-service/internal/adapter/metrics"
	"github.com/tezos-delegation-service/internal/model"
)

// getDelegations handles business logic for delegations.
type getDelegations struct {
	dbAdapter    database.Adapter
	defaultLimit uint16
}

// GetDelegationsFunc defines the function signature for fetching delegations.
type GetDelegationsFunc func(ctx context.Context, pageStr, limitStr, yearStr string, maxDelegationID int64) (*model.DelegationsResponse, error)

// NewGetDelegationsFunc creates a new instance of getDelegations.
func NewGetDelegationsFunc(defaultLimit uint16, adapter database.Adapter, metricsClient metrics.Adapter) GetDelegationsFunc {
	uc := &getDelegations{
		dbAdapter:    adapter,
		defaultLimit: defaultLimit,
	}
	return uc.withMonitorer(uc.GetDelegations, metricsClient)
}

// GetDelegations returns delegations with pagination and optional year filter.
func (uc *getDelegations) GetDelegations(ctx context.Context, pageStr, limitStr, yearStr string, maxDelegationID int64) (*model.DelegationsResponse, error) {
	page, err := uc.parsePage(pageStr)
	if err != nil {
		return nil, err
	}

	limit, err := uc.parseLimit(limitStr)
	if err != nil {
		return nil, err
	}

	year, err := uc.parseYear(yearStr)
	if err != nil {
		return nil, err
	}

	maxDelegationIDUint := uint64(0)
	if maxDelegationID > 0 {
		maxDelegationIDUint = uint64(maxDelegationID)
	}
	delegations, err := uc.dbAdapter.GetDelegations(ctx, page, limit, year, maxDelegationIDUint)
	if err != nil {
		return nil, err
	}

	var maxID int64 = 0
	for i, delegation := range delegations {
		if delegation.ID > maxID {
			maxID = delegation.ID
		}
		delegations[i].TimestampTime = time.Unix(delegation.Timestamp, 0).UTC().Format(time.RFC3339)
		delegations[i].Amount = delegation.Amount * 1000000.0 // Convert tez to mutez
	}

	pageInt := int(page)
	limitInt := int(limit)

	paginationInfo := model.PaginationInfo{
		CurrentPage: pageInt,
		PerPage:     limitInt,
		HasPrevPage: page > 1,
	}

	if page > 1 {
		paginationInfo.PrevPage = pageInt - 1
	}

	return &model.DelegationsResponse{
		Delegations:     delegations,
		Pagination:      paginationInfo,
		MaxDelegationID: maxID,
	}, nil
}

// parsePage parses the page from the string and returns it as an integer.
func (uc *getDelegations) parsePage(pageStr string) (uint32, error) {
	page := uint32(1)
	if pageStr != "" {
		p, err := strconv.Atoi(pageStr)
		if err != nil {
			return 0, err
		}
		if p <= 0 {
			return 0, errors.New("page must be a positive number")
		}
		if p > int(^uint32(0)) {
			return 0, errors.New("page number exceeds maximum allowed value of 4294967295")
		}
		page = uint32(p)
	}
	return page, nil
}

// parseLimit parses the limit from the string and returns it as an integer.
func (uc *getDelegations) parseLimit(limitStr string) (uint16, error) {
	limit := uc.defaultLimit
	if limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err != nil {
			return 0, err
		}
		if l <= 0 {
			return 0, errors.New("limit must be a positive number")
		}
		if l > 500 {
			return 0, errors.New("limit exceeds maximum allowed value of 500")
		}
		limit = uint16(l)
	}
	return limit, nil
}

// parseYear parses the year from the string and returns it as an integer.
func (uc *getDelegations) parseYear(yearStr string) (uint16, error) {
	var year uint16 = 0
	if yearStr != "" {
		y, err := strconv.Atoi(yearStr)
		if err != nil {
			return 0, err
		}
		if y <= 0 {
			return 0, errors.New("year must be a positive number")
		}

		currentYear := time.Now().Year()
		if y > currentYear {
			return 0, errors.New("year cannot exceed the current year")
		}

		year = uint16(y)
	}
	return year, nil
}

// withMonitorer wraps the GetDelegations function with telemetry monitoring.
func (uc *getDelegations) withMonitorer(getDelegations GetDelegationsFunc, metricsClient metrics.Adapter) GetDelegationsFunc {
	return func(ctx context.Context, pageStr, limitStr, yearStr string, maxDelegationID int64) (result *model.DelegationsResponse, err error) {
		startTime := time.Now()

		defer func() {
			if metricsClient != nil {
				duration := time.Since(startTime)

				metricsClient.RecordServiceOperation("GetDelegations", "UseCase", duration, err)
				if err == nil && result != nil {
					metricsClient.RecordDelegationsFetched(len(result.Delegations))
				}
			}
		}()

		return getDelegations(ctx, pageStr, limitStr, yearStr, maxDelegationID)
	}
}
