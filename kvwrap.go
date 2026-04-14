package kvwrap

import (
	"context"
	"fmt"
)

var (
	ErrKeyNotFound      = fmt.Errorf("key not found")
	ErrInvalidPartition = fmt.Errorf("invalid partition")
	ErrInvalidKey       = fmt.Errorf("invalid key")
	ErrInvalidValue     = fmt.Errorf("invalid value")
)

type ScanResult struct {
	Key   []byte
	Value []byte
	Err   error
}

type WatchEvent struct {
	Type      string // "Set" or "Delete"
	Partition string
	Key       []byte
	Value     []byte // Only set for "Set" events
	Err       error
}

type KVWrapClient interface {
	// Get retrieves the value associated with the given key.
	Get(ctx context.Context, partition string, key []byte) ([]byte, error)
	// Set sets the value for the given key.
	Set(ctx context.Context, partition string, key []byte, value []byte) error
	// Delete removes the key-value pair associated with the given key.
	Delete(ctx context.Context, partition string, key []byte) error
	// Scan retrieves all key-value pairs that match the given optional prefix.
	Scan(ctx context.Context, partition string, prefix []byte) <-chan ScanResult
	// WatchKey watches for changes to a specific key and sends events to the provided channel.
	WatchKey(ctx context.Context, partition string, key []byte) <-chan WatchEvent
	// WatchPrefix watches for changes to keys with a specific prefix and sends events to the provided channel.
	WatchPrefix(ctx context.Context, partition string, prefix []byte) <-chan WatchEvent
	// GetJSON retrieves the value associated with the given key and unmarshals it into the provided struct.
	GetJSON(ctx context.Context, partition string, key []byte, value any) error
	// SetJSON marshals the provided struct and sets it as the value for the given key.
	SetJSON(ctx context.Context, partition string, key []byte, value any) error
}
