package model

// OperationType represents the type of operation in the Tezos blockchain.
type OperationType string

var (
	// OperationTypeDelegate represents a delegate operation.
	OperationTypeDelegate OperationType = "delegate"
	// OperationTypeUnDelegate represents an unDelegate operation.
	OperationTypeUnDelegate OperationType = "undelegate"
	// OperationTypeStake represents a stake operation.
	OperationTypeStake OperationType = "stake"
	// OperationTypeUnStake represents an unstake operation.
	OperationTypeUnStake OperationType = "unstake"
	// OperationTypeReward represents a reward operation.
	OperationTypeReward OperationType = "reward"
)

// String returns the string representation of the operation type.
func (o OperationType) String() string {
	return string(o)
}

// IsValid checks if the operation type is valid.
func (o OperationType) IsValid() bool {
	switch o {
	case OperationTypeDelegate, OperationTypeUnDelegate, OperationTypeStake, OperationTypeUnStake, OperationTypeReward:
		return true
	default:
		return false
	}
}

// Operation represents a Tezos operation.
type Operation struct {
	ID              int64         `db:"id" json:"id"`
	SenderAddress   WalletAddress `db:"sender_address" json:"sender_address"`
	ContractAddress WalletAddress `db:"contract_address" json:"contract_address"`
	Entrypoint      string        `db:"entrypoint" json:"entrypoint"`
	Amount          float64       `db:"amount" json:"amount"`
	Block           string        `db:"block" json:"block"`
	Timestamp       int64         `db:"timestamp" json:"-"`
	TimestampTime   string        `db:"-" json:"timestamp"`
	Status          string        `db:"status" json:"status"`
	Type            OperationType `db:"type" json:"type"`
}

// OperationsResponse is the response format for the API.
type OperationsResponse struct {
	Operations []Operation    `json:"data"`
	Pagination PaginationInfo `json:"-,omitempty"`
}
