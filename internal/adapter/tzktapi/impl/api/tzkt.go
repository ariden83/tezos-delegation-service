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
func (a *Adapter) FetchDelegations(ctx context.Context, limit uint16, offset int) (model.TzktDelegationResponse, error) {
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
func (a *Adapter) FetchDelegationsFromLevel(ctx context.Context, level uint64, limit uint8) (model.TzktDelegationResponse, error) {
	url := fmt.Sprintf("%s/v1/operations/delegations?level.gt=%d", a.apiURL, level)

	// Ajouter une limite si spécifiée
	if limit > 0 {
		url += fmt.Sprintf("&limit=%d", limit)
	}

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
			a.logger.Errorf("error closing response body: %v", err) // Utilisation du logger au lieu de fmt
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

// FetchOperationsFromTezos fetches operations from the Tezos node.
func (a *Adapter) FetchOperationsFromTezos(blockID string) ([]model.Operation, error) {
	url := fmt.Sprintf("%s/chains/main/blocks/%s/operations", a.apiURL, blockID)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			a.logger.Errorf("error closing response body: %v", err)
		}
	}()

	var operations []model.Operation
	if err := json.NewDecoder(resp.Body).Decode(&operations); err != nil {
		return nil, err
	}
	return operations, nil
}

// FetchRewardsForBaker fetches rewards for a specific baker from the Tezos node.
func (a *Adapter) FetchRewardsForBaker(blockID, bakerAddress string) (model.Reward, error) {
	url := fmt.Sprintf("%s/chains/main/blocks/%s/context/delegates/%s", a.apiURL, blockID, bakerAddress)
	resp, err := http.Get(url)
	if err != nil {
		return model.Reward{}, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			a.logger.Errorf("error closing response body: %v", err)
		}
	}()

	var reward model.Reward
	if err := json.NewDecoder(resp.Body).Decode(&reward); err != nil {
		return model.Reward{}, err
	}
	return reward, nil
}

// GetCurrentCycle returns the current cycle from the TzKT API.
func (a *Adapter) GetCurrentCycle(ctx context.Context) (int, error) {
	url := fmt.Sprintf("%s/v1/head", a.apiURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("error creating request: %w", err)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("error fetching current cycle: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			a.logger.Errorf("error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var head struct {
		Cycle int `json:"cycle"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&head); err != nil {
		return 0, fmt.Errorf("error decoding response: %w", err)
	}

	return head.Cycle, nil
}

// FetchRewardsForCycle fetches rewards for a specific delegator and baker in a given cycle.
func (a *Adapter) FetchRewardsForCycle(ctx context.Context, delegator model.WalletAddress, baker model.WalletAddress, cycle int) ([]model.Reward, error) {
	// TzKT API endpoint for rewards
	url := fmt.Sprintf("%s/v1/rewards/delegators/%s/%d", a.apiURL, delegator, cycle)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching rewards: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			a.logger.Errorf("error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// TzKT API response for rewards
	var rewardsResponse struct {
		RewardsShare float64 `json:"rewardsShare"`
		Baker        struct {
			Address string `json:"address"`
		} `json:"baker"`
		Cycle     int   `json:"cycle"`
		Timestamp int64 `json:"timestamp"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&rewardsResponse); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	// Only create a reward if there's an actual reward amount
	if rewardsResponse.RewardsShare == 0 {
		return nil, nil
	}

	// Create reward model
	reward := model.Reward{
		RecipientAddress: delegator,
		SourceAddress:    model.WalletAddress(rewardsResponse.Baker.Address),
		Cycle:            rewardsResponse.Cycle,
		Amount:           rewardsResponse.RewardsShare,
		Timestamp:        rewardsResponse.Timestamp,
		TimestampTime:    time.Unix(rewardsResponse.Timestamp, 0).Format(time.RFC3339),
	}

	return []model.Reward{reward}, nil
}

// FetchWalletInfo fetches wallet information from the Tezos node.
func (a *Adapter) FetchWalletInfo(blockID, walletAddress string) (model.WalletInfo, error) {
	url := fmt.Sprintf("%s/chains/main/blocks/%s/context/contracts/%s", blockID, walletAddress)
	resp, err := http.Get(url)
	if err != nil {
		return model.WalletInfo{}, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			a.logger.Errorf("error closing response body: %v", err)
		}
	}()

	var walletInfo model.WalletInfo
	if err := json.NewDecoder(resp.Body).Decode(&walletInfo); err != nil {
		return model.WalletInfo{}, err
	}
	return walletInfo, nil
}

// FetchStakingOperations fetches staking operations from the TzKT API.
func (a *Adapter) FetchStakingOperations(ctx context.Context, filter tzktapi.OperationFilter) ([]model.StakingOperation, error) {
	delegations, err := a.fetchDelegationOperations(ctx, filter)
	if err != nil {
		return nil, err
	}

	transactions, err := a.fetchTransactionOperations(ctx, filter)
	if err != nil {
		return nil, err
	}

	return append(delegations, transactions...), nil
}

// fetchDelegationOperations fetches delegation operations from the TzKT API.
func (a *Adapter) fetchDelegationOperations(ctx context.Context, filter tzktapi.OperationFilter) ([]model.StakingOperation, error) {
	var ops []model.StakingOperation

	formatDate := func(t *int64) string {
		if t == nil {
			return ""
		}
		timestamp := time.Unix(*t, 0)
		return timestamp.Format("2006-01-02T15:04:05Z")
	}

	// --- Fetch delegations ---
	delegationURL := fmt.Sprintf("%s/v1/operations/delegations?limit=%d&offset=%d", a.apiURL, filter.Limit, filter.Offset)
	if filter.Wallet != "" {
		delegationURL += fmt.Sprintf("&sender=%s", filter.Wallet)
	}
	if filter.Baker != "" {
		delegationURL += fmt.Sprintf("&newDelegate=%s", filter.Baker)
	}
	if from := formatDate(filter.FromDate); from != "" {
		delegationURL += fmt.Sprintf("&timestamp.ge=%s", from)
	}
	if to := formatDate(filter.ToDate); to != "" {
		delegationURL += fmt.Sprintf("&timestamp.le=%s", to)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", delegationURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			a.logger.Errorf("error closing response body: %v", err)
		}
	}()

	var delegations []struct {
		Hash        string                   `json:"hash"`
		Sender      struct{ Address string } `json:"sender"`
		NewDelegate struct{ Address string } `json:"newDelegate"`
		Timestamp   time.Time                `json:"timestamp"`
		Status      string                   `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&delegations); err != nil {
		return nil, err
	}

	for _, d := range delegations {
		ops = append(ops, model.StakingOperation{
			Hash:      d.Hash,
			Type:      "delegation",
			Wallet:    model.WalletAddress(d.Sender.Address),
			Baker:     model.WalletAddress(d.NewDelegate.Address),
			Timestamp: d.Timestamp,
			Status:    d.Status,
		})
	}

	return ops, nil
}

// fetchTransactionOperations fetches transaction operations (stake/unstake/claim_rewards) from the TzKT API.
func (a *Adapter) fetchTransactionOperations(ctx context.Context, filter tzktapi.OperationFilter) ([]model.StakingOperation, error) {
	var ops []model.StakingOperation

	formatDate := func(t *int64) string {
		if t == nil {
			return ""
		}
		timestamp := time.Unix(*t, 0)
		return timestamp.Format("2006-01-02T15:04:05Z")
	}

	// --- Fetch transactions (stake/unstake/claim_rewards) ---
	transactionURL := fmt.Sprintf("%s/v1/operations/transactions?limit=%d&offset=%d", a.apiURL, filter.Limit, filter.Offset)
	if filter.Wallet != "" {
		transactionURL += fmt.Sprintf("&sender=%s", filter.Wallet)
	}
	if filter.Baker != "" {
		transactionURL += fmt.Sprintf("&target=%s", filter.Baker)
	}
	if from := formatDate(filter.FromDate); from != "" {
		transactionURL += fmt.Sprintf("&timestamp.ge=%s", from)
	}
	if to := formatDate(filter.ToDate); to != "" {
		transactionURL += fmt.Sprintf("&timestamp.le=%s", to)
	}
	transactionURL += "&entrypoint.in=stake,unstake,claim_rewards"

	req, err := http.NewRequestWithContext(ctx, "GET", transactionURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			a.logger.Errorf("error closing response body: %v", err)
		}
	}()

	var txs []struct {
		Hash       string                   `json:"hash"`
		Sender     struct{ Address string } `json:"sender"`
		Target     struct{ Address string } `json:"target"`
		Entrypoint string                   `json:"entrypoint"`
		Amount     float64                  `json:"amount"`
		Timestamp  time.Time                `json:"timestamp"`
		Status     string                   `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&txs); err != nil {
		return nil, err
	}

	for _, t := range txs {
		ops = append(ops, model.StakingOperation{
			Hash:       t.Hash,
			Type:       model.OperationType(t.Entrypoint),
			Entrypoint: t.Entrypoint,
			Wallet:     model.WalletAddress(t.Sender.Address),
			Baker:      model.WalletAddress(t.Target.Address),
			Amount:     t.Amount / 1_000_000, // µꜩ → ꜩ
			Timestamp:  t.Timestamp,
			Status:     t.Status,
		})
	}

	return ops, nil
}
