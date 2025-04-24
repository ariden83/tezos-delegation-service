package model

// WalletInfo represents the information of a wallet in the Tezos blockchain.
type WalletInfo struct {
	Balance             string        `json:"balance"`
	Counter             string        `json:"counter"`
	Delegate            *DelegateInfo `json:"delegate,omitempty"`
	StorageSize         string        `json:"storage_size"`
	PaidStorageSizeDiff string        `json:"paid_storage_size_diff"`
	Script              *ScriptInfo   `json:"script,omitempty"`
}

// DelegateInfo represents the information of a delegate in the Tezos blockchain.
type DelegateInfo struct {
	Setable      bool   `json:"setable"`
	Value        string `json:"value"`
	ConsensusKey string `json:"consensus_key"`
}

// ScriptInfo represents the information of a script in the Tezos blockchain.
type ScriptInfo struct {
	Code    interface{} `json:"code"`    // ou []interface{}, si tu veux parser le script Michelson
	Storage interface{} `json:"storage"` // idem ici
}
