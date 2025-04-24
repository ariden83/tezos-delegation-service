package tzktapi

// OperationFilter defines the filter for fetching staking operations.
type OperationFilter struct {
	Limit    int
	Offset   int
	Wallet   string
	Baker    string
	FromDate *int64
	ToDate   *int64
}
