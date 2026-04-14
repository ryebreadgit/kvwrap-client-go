# kvwrap-client-go

A Go client library for connecting to remote [KVWrap](https://github.com/ryebreadgit/kvwrap) key-value stores via gRPC, providing the same Get, Set, Delete, Scan, and Watch operations available in the Rust client.

## Installation

```bash
go get github.com/ryebreadgit/kvwrap-client-go
```

## Usage

```go
package main

import (
	"context"
	"fmt"

	"github.com/ryebreadgit/kvwrap-client-go"
)

func main() {
	kv, err := kvwrap.NewRemoteStore(kvwrap.RemoteConfig{
		Endpoint:       "localhost:50051",
		RequestTimeout: 5,
	})
	if err != nil {
		fmt.Println("Error creating remote store:", err)
		return
	}
	defer kv.Close()

	ctx := context.Background()

	// Set a value
	err = kv.Set(ctx, "mypartition", []byte("greeting"), []byte("Hello World!"))
	if err != nil {
		fmt.Println("Set error:", err)
		return
	}

	// Get it back
	value, err := kv.Get(ctx, "mypartition", []byte("greeting"))
	if err != nil {
		fmt.Println("Get error:", err)
		return
	}
	fmt.Printf("Got: %s\n", string(value))

	// Delete it
	err = kv.Delete(ctx, "mypartition", []byte("greeting"))
	if err != nil {
		fmt.Println("Delete error:", err)
		return
	}
	fmt.Println("Deleted successfully")
}
```

## JSON Support

Store and retrieve structs directly with `SetJSON` and `GetJSON`:

```go
type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Store as JSON
kv.SetJSON(ctx, "users", []byte("user:1"), User{Name: "Alice", Email: "alice@example.com"})

// Retrieve and unmarshal
var user User
kv.GetJSON(ctx, "users", []byte("user:1"), &user)
```

## Scanning Keys

Iterate over all key-value pairs matching a prefix:

```go
results, err := kv.Scan(ctx, "mypartition", []byte("user:"), 32)
if err != nil {
    fmt.Println("Scan error:", err)
    return
}

for result := range results {
    if result.Err != nil {
        fmt.Println("Error during scan:", result.Err)
        break
    }
    fmt.Printf("Key: %s, Value: %s\n", result.Key, result.Value)
}
```

## Proto Generation

Regenerate Go bindings from the proto definitions:

```bash
buf generate
```

Requires [`buf`](https://buf.build/docs/installation), `protoc-gen-go`, and `protoc-gen-go-grpc`.

## License

[MIT](./LICENSE.md)