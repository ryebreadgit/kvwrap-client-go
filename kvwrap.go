package kvwrapclient

import (
	"context"
)

type KVPair struct {
	Key   []byte
	Value []byte
}

type WatchEvent struct {
	Type      string // "Set" or "Delete"
	Partition string
	Key       []byte
	Value     []byte // Only set for "Set" events
}

type KVWrapClient interface {
	// Get retrieves the value associated with the given key.
	Get(key []byte) ([]byte, error)
	// Set sets the value for the given key.
	Set(key []byte, value []byte) error
	// Delete removes the key-value pair associated with the given key.
	Delete(key []byte) error
	// Scan retrieves all key-value pairs that match the given optional prefix.
	Scan(ctx context.Context, partition string, prefix []byte) (<-chan KVPair, <-chan error)
	// WatchKey watches for changes to a specific key and sends events to the provided channel.
	WatchKey(ctx context.Context, partition string, key []byte) (<-chan WatchEvent, <-chan error)
	// WatchPrefix watches for changes to keys with a specific prefix and sends events to the provided channel.
	WatchPrefix(ctx context.Context, partition string, prefix []byte) (<-chan WatchEvent, <-chan error)
	// GetJSON retrieves the value associated with the given key and unmarshals it into the provided struct.
	GetJSON(key []byte, value any) error
	// SetJSON marshals the provided struct and sets it as the value for the given key.
	SetJSON(key []byte, value any) error
}
