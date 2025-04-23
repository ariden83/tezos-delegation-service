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

// getRewards handles business logic for delegations.
type getRewards struct {
	dbAdapter    database.Adapter
	defaultLimit uint16
}

// GetRewardsFunc defines the function signature for fetching delegations.
type GetRewardsFunc func(ctx context.Context, pageStr, limitStr string) (*model.RewardsResponse, error)

// NewGetRewardsFunc creates a new instance of getRewards.
func NewGetRewardsFunc(defaultLimit uint16, adapter database.Adapter, metricsClient metrics.Adapter) GetRewardsFunc {
	uc := &getRewards{
		dbAdapter:    adapter,
		defaultLimit: defaultLimit,
	}
	return uc.withMonitorer(uc.GetRewards, metricsClient)
}

// GetRewards returns delegations with pagination and optional year filter.
func (uc *getRewards) GetRewards(ctx context.Context, pageStr, limitStr, yearStr string, maxDelegationID int64) (*model.RewardsResponse, error) {
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
	delegations, count, err := uc.dbAdapter.GetRewards(ctx, page, limit, year, maxDelegationIDUint)
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
	hasMore := (pageInt * limitInt) < count

	paginationInfo := model.PaginationInfo{
		CurrentPage: pageInt,
		PerPage:     limitInt,
		HasPrevPage: page > 1,
		HasNextPage: hasMore,
	}

	if page > 1 {
		paginationInfo.PrevPage = pageInt - 1
	}

	if hasMore {
		paginationInfo.NextPage = pageInt + 1
	}

	return &model.RewardsResponse{
		Data:       delegations,
		Pagination: paginationInfo,
	}, nil
}

// parsePage parses the page from the string and returns it as an integer.
func (uc *getRewards) parsePage(pageStr string) (uint32, error) {
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
func (uc *getRewards) parseLimit(limitStr string) (uint16, error) {
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
func (uc *getRewards) parseYear(yearStr string) (uint16, error) {
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

// withMonitorer wraps the GetRewards function with telemetry monitoring.
func (uc *getRewards) withMonitorer(getRewards GetRewardsFunc, metricsClient metrics.Adapter) GetRewardsFunc {
	return func(ctx context.Context, pageStr, limitStr string) (result *model.RewardsResponse, err error) {
		startTime := time.Now()

		defer func() {
			if metricsClient != nil {
				duration := time.Since(startTime)

				metricsClient.RecordServiceOperation("GetRewards", "UseCase", duration, err)
				if err == nil && result != nil {
					metricsClient.RecordDelegationsFetched(len(result.Data))
				}
			}
		}()

		return getRewards(ctx, pageStr, limitStr)
	}
}
