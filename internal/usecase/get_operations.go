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

// getOperations handles business logic for delegations.
type getOperations struct {
	dbAdapter    database.Adapter
	defaultLimit uint16
}

// GetOperationsInput defines the input structure for fetching delegations.
type GetOperationsInput struct {
	// @todo
	Page   string
	Limit  string
	Wallet model.WalletAddress
	Backer model.WalletAddress
	Type   model.OperationType
}

// GetOperationsFunc defines the function signature for fetching delegations.
type GetOperationsFunc func(ctx context.Context, pageStr, limitStr string, operationType model.OperationType, wallet, backer model.WalletAddress) (*model.OperationsResponse, error)

// NewGetOperationsFunc creates a new instance of getOperations.
func NewGetOperationsFunc(defaultLimit uint16, adapter database.Adapter, metricsClient metrics.Adapter) GetOperationsFunc {
	uc := &getOperations{
		dbAdapter:    adapter,
		defaultLimit: defaultLimit,
	}
	return uc.withMonitorer(uc.GetOperations, metricsClient)
}

// GetOperations returns delegations with pagination and optional year filter.
func (uc *getOperations) GetOperations(ctx context.Context, pageStr, limitStr string, operationType model.OperationType, wallet, backer model.WalletAddress) (*model.OperationsResponse, error) {
	page, err := uc.parsePage(pageStr)
	if err != nil {
		return nil, err
	}

	limit, err := uc.parseLimit(limitStr)
	if err != nil {
		return nil, err
	}

	operations, err := uc.dbAdapter.GetOperations(ctx, page, limit, operationType, wallet, backer)
	if err != nil {
		return nil, err
	}

	for i, operation := range operations {
		operations[i].TimestampTime = time.Unix(operation.Timestamp, 0).UTC().Format(time.RFC3339)
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

	return &model.OperationsResponse{
		Data:       operations,
		Pagination: paginationInfo,
	}, nil
}

// parsePage parses the page from the string and returns it as an integer.
func (uc *getOperations) parsePage(pageStr string) (uint16, error) {
	page := uint16(1)
	if pageStr != "" {
		p, err := strconv.Atoi(pageStr)
		if err != nil {
			return 0, err
		}
		if p <= 0 {
			return 0, errors.New("page must be a positive number")
		}
		if p > int(^uint16(0)) {
			return 0, errors.New("page number exceeds maximum allowed value of 4294967295")
		}
		page = uint16(p)
	}
	return page, nil
}

// parseLimit parses the limit from the string and returns it as an integer.
func (uc *getOperations) parseLimit(limitStr string) (uint16, error) {
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
func (uc *getOperations) parseYear(yearStr string) (uint16, error) {
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

// withMonitorer wraps the GetOperations function with telemetry monitoring.
func (uc *getOperations) withMonitorer(getOperations GetOperationsFunc, metricsClient metrics.Adapter) GetOperationsFunc {
	return func(ctx context.Context, pageStr, limitStr string, operationType model.OperationType, wallet, backer model.WalletAddress) (result *model.OperationsResponse, err error) {
		startTime := time.Now()

		defer func() {
			if metricsClient != nil {
				duration := time.Since(startTime)

				metricsClient.RecordServiceOperation("GetOperations", "UseCase", duration, err)
				if err == nil && result != nil {
					metricsClient.RecordDelegationsFetched(len(result.Data))
				}
			}
		}()

		return getOperations(ctx, pageStr, limitStr, operationType, wallet, backer)
	}
}
