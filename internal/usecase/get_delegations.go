package usecase

import (
	"context"
	"strconv"
	"time"

	"github.com/tezos-delegation-service/internal/adapter/metrics"
	"github.com/tezos-delegation-service/internal/adapter/tzktapi"
	"github.com/tezos-delegation-service/internal/model"
)

// getDelegations handles business logic for delegations.
type getDelegations struct {
	tzktApiAdapter tzktapi.Adapter
}

// GetDelegationsFunc defines the function signature for fetching delegations.
type GetDelegationsFunc func(ctx context.Context, pageStr, limitStr, yearStr string) (*model.DelegationResponse, error)

// NewGetDelegationsFunc creates a new instance of getDelegations.
func NewGetDelegationsFunc(adapter tzktapi.Adapter, metricsClient metrics.Adapter) GetDelegationsFunc {
	uc := &getDelegations{
		tzktApiAdapter: adapter,
	}
	return uc.withMonitorer(uc.GetDelegations, metricsClient)
}

// GetDelegations returns delegations with pagination and optional year filter.
func (uc *getDelegations) GetDelegations(ctx context.Context, pageStr, limitStr, yearStr string) (*model.DelegationResponse, error) {
	page := uc.parsePage(pageStr)
	limit := uc.parseLimit(limitStr)
	year := uc.parseYear(yearStr)

	tzktDelegations, err := uc.tzktApiAdapter.FetchDelegations(ctx, limit, (page-1)*limit)
	if err != nil {
		return nil, err
	}

	delegations := make([]model.Delegation, 0, len(tzktDelegations))
	for _, tzktDel := range tzktDelegations {
		if tzktDel.Status != "applied" {
			continue
		}

		if year > 0 {
			delegationYear := tzktDel.Timestamp.Year()
			if delegationYear != year {
				continue
			}
		}

		delegations = append(delegations, model.Delegation{
			Delegator: tzktDel.Sender.Address,
			Delegate:  tzktDel.Delegate.Address,
			Amount:    float64(tzktDel.Amount) / 1000000.0, // Convertion mutez into tez
			Timestamp: tzktDel.Timestamp.Unix(),
			Level:     tzktDel.Level,
		})
	}

	hasMore := len(delegations) >= limit

	paginationInfo := model.PaginationInfo{
		CurrentPage: page,
		PerPage:     limit,
		HasPrevPage: page > 1,
		HasNextPage: hasMore,
	}

	if page > 1 {
		paginationInfo.PrevPage = page - 1
	}

	if hasMore {
		paginationInfo.NextPage = page + 1
	}

	return &model.DelegationResponse{
		Data:       delegations,
		Pagination: paginationInfo,
	}, nil
}

// parsePage parses the page from the string and returns it as an integer.
func (uc *getDelegations) parsePage(pageStr string) int {
	page := 1
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	return page
}

// parseLimit parses the limit from the string and returns it as an integer.
func (uc *getDelegations) parseLimit(limitStr string) int {
	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	return limit
}

// parseYear parses the year from the string and returns it as an integer.
func (uc *getDelegations) parseYear(yearStr string) int {
	year := 0
	if yearStr != "" {
		if y, err := strconv.Atoi(yearStr); err == nil && y > 0 {
			year = y
		}
	}
	return year
}

// withMonitorer wraps the GetDelegations function with telemetry monitoring.
func (uc *getDelegations) withMonitorer(getDelegations GetDelegationsFunc, metricsClient metrics.Adapter) GetDelegationsFunc {
	return func(ctx context.Context, pageStr, limitStr, yearStr string) (result *model.DelegationResponse, err error) {
		startTime := time.Now()

		defer func() {
			if metricsClient != nil {
				duration := time.Since(startTime)

				metricsClient.RecordServiceOperation("GetDelegations", "UseCase", duration, err)
				if err == nil && result != nil {
					metricsClient.RecordDelegationsFetched(len(result.Data))
				}
			}
		}()

		return getDelegations(ctx, pageStr, limitStr, yearStr)
	}
}
