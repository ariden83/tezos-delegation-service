package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/tezos-delegation-service/internal/adapter/database"
	"github.com/tezos-delegation-service/internal/adapter/tzktapi"
	"github.com/tezos-delegation-service/internal/model"
)

// Config holds configuration for the TzKT API adapter
type Config struct {
	URL     string        `mapstructure:"url"`
	Timeout time.Duration `mapstructure:"timeout"`
}

// Adapter implements the TzKT API adapter interface
type Adapter struct {
	apiURL string
	client *http.Client
	db     database.Adapter
	logger *logrus.Entry
}

// New creates a new real TzKT API adapter
func New(cfg Config, logger *logrus.Entry) (tzktapi.Adapter, error) {
	if cfg.URL == "" {
		return nil, errors.New("TzKT API URL is required")
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}

	return &Adapter{
		apiURL: cfg.URL,
		client: &http.Client{Timeout: cfg.Timeout},
		logger: logger,
	}, nil
}

// FetchDelegations fetches delegations from the TzKT API
func (a *Adapter) FetchDelegations(ctx context.Context, limit int, offset int) (model.TzktDelegationResponse, error) {
	url := fmt.Sprintf("%s/v1/operations/delegations?limit=%d&offset=%d", a.apiURL, limit, offset)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching delegations: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			a.logger.Errorf("error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var delegations model.TzktDelegationResponse
	if err := json.NewDecoder(resp.Body).Decode(&delegations); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return delegations, nil
}

// FetchDelegationsFromLevel fetches delegations from a specific level
func (a *Adapter) FetchDelegationsFromLevel(ctx context.Context, level uint64) (model.TzktDelegationResponse, error) {
	url := fmt.Sprintf("%s/v1/operations/delegations?level.gt=%d", a.apiURL, level)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching delegations from level: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("error closing response body: %v\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var delegations model.TzktDelegationResponse
	if err := json.NewDecoder(resp.Body).Decode(&delegations); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return delegations, nil
}
