package model

import "context"

// SyncFunc is a function type that takes a context and returns an error.
type SyncFunc func(ctx context.Context) error
