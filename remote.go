package kvwrap

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ryebreadgit/kvwrap-client-go/kvwrappb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

var _ KVWrapClient = (*RemoteStore)(nil)

type RemoteStore struct {
	client         kvwrappb.KvServiceClient
	conn           *grpc.ClientConn
	requestTimeout time.Duration
}

func NewRemoteStore(config RemoteConfig) (*RemoteStore, error) {
	dialOptions := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	if config.ConnectionTimeout > 0 {
		dialOptions = append(dialOptions, grpc.WithBlock(), grpc.WithTimeout(time.Duration(config.ConnectionTimeout)*time.Second))
	}

	conn, err := grpc.NewClient(config.Endpoint, dialOptions...)
	if err != nil {
		return nil, ErrConnectionFailed
	}

	return &RemoteStore{
		client:         kvwrappb.NewKvServiceClient(conn),
		conn:           conn,
		requestTimeout: time.Duration(config.RequestTimeout) * time.Second,
	}, nil
}

func (r *RemoteStore) Get(ctx context.Context, partition string, key []byte) ([]byte, error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	resp, err := r.client.Get(ctx, &kvwrappb.GetRequest{
		Partition: partition,
		Key:       key,
	})
	if err != nil {
		return nil, mapError(err)
	}

	return resp.Value, nil
}

func (r *RemoteStore) Set(ctx context.Context, partition string, key []byte, value []byte) error {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	_, err := r.client.Set(ctx, &kvwrappb.SetRequest{
		Partition: partition,
		Key:       key,
		Value:     value,
	})
	return mapError(err)
}

func (r *RemoteStore) Delete(ctx context.Context, partition string, key []byte) error {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	_, err := r.client.Delete(ctx, &kvwrappb.DeleteRequest{
		Partition: partition,
		Key:       key,
	})
	return mapError(err)
}

func (r *RemoteStore) GetJSON(ctx context.Context, partition string, key []byte, value any) error {
	val, err := r.Get(ctx, partition, key)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(val, value); err != nil {
		return ErrJSONUnmarshal
	}
	return nil
}

func (r *RemoteStore) SetJSON(ctx context.Context, partition string, key []byte, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return ErrJSONMarshal
	}
	return r.Set(ctx, partition, key, data)
}

func (r *RemoteStore) Scan(ctx context.Context, partition string, prefix []byte, bufferSize uint32) (<-chan ScanResult, error) {
	// Use client.watch
	data, err := r.client.AllKeys(ctx, &kvwrappb.AllKeysRequest{
		Partition:  partition,
		Prefix:     prefix,
		BufferSize: bufferSize,
	})
	if err != nil {
		return nil, mapError(err)
	}

	results := make(chan ScanResult, bufferSize)
	go func() {
		defer close(results)
		for {
			resp, err := data.Recv()
			if err != nil {
				if err == context.Canceled || err == context.DeadlineExceeded {
					return
				}
				results <- ScanResult{Err: mapError(err)}
				return
			}
			results <- ScanResult{
				Key:   resp.Key,
				Value: resp.Value,
			}
		}
	}()
	return results, nil
}

func (r *RemoteStore) WatchKey(ctx context.Context, partition string, key []byte, bufferSize uint32) (<-chan WatchEvent, error) {
	data, err := r.client.Watch(ctx, &kvwrappb.WatchRequest{
		Partition:   partition,
		IsPrefix:    false,
		KeyOrPrefix: key,
	})
	if err != nil {
		return nil, mapError(err)
	}

	events := make(chan WatchEvent, bufferSize)
	go func() {
		defer close(events)
		for {
			resp, err := data.Recv()
			if err != nil {
				if err == context.Canceled || err == context.DeadlineExceeded {
					return
				}
				events <- WatchEvent{Err: mapError(err)}
				return
			}
			var eventType EventType
			switch resp.EventType {
			case kvwrappb.WatchEventMessage_SET:
				eventType = EventSet
			case kvwrappb.WatchEventMessage_DELETE:
				eventType = EventDelete
			default:
				events <- WatchEvent{Err: fmt.Errorf("unknown event type: %v", resp.EventType)}
				continue
			}
			events <- WatchEvent{
				Type:      eventType,
				Partition: partition,
				Key:       resp.Key,
				Value:     resp.Value,
			}
		}
	}()
	return events, nil

}

func (r *RemoteStore) WatchPrefix(ctx context.Context, partition string, prefix []byte, bufferSize uint32) (<-chan WatchEvent, error) {
	data, err := r.client.Watch(ctx, &kvwrappb.WatchRequest{
		Partition:   partition,
		IsPrefix:    true,
		KeyOrPrefix: prefix,
	})
	if err != nil {
		return nil, mapError(err)
	}

	events := make(chan WatchEvent, bufferSize)
	go func() {
		defer close(events)
		for {
			resp, err := data.Recv()
			if err != nil {
				if err == context.Canceled || err == context.DeadlineExceeded {
					return
				}
				events <- WatchEvent{Err: mapError(err)}
				return
			}
			var eventType EventType
			switch resp.EventType {
			case kvwrappb.WatchEventMessage_SET:
				eventType = EventSet
			case kvwrappb.WatchEventMessage_DELETE:
				eventType = EventDelete
			default:
				events <- WatchEvent{Err: fmt.Errorf("unknown event type: %v", resp.EventType)}
				continue
			}
			events <- WatchEvent{
				Type:      eventType,
				Partition: partition,
				Key:       resp.Key,
				Value:     resp.Value,
			}
		}
	}()
	return events, nil
}

func (r *RemoteStore) Close() error {
	return r.conn.Close()
}

func (r *RemoteStore) withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if r.requestTimeout > 0 {
		if _, ok := ctx.Deadline(); !ok {
			return context.WithTimeout(ctx, r.requestTimeout)
		}
	}
	return ctx, func() {}
}
func mapError(err error) error {
	st, ok := status.FromError(err)
	if !ok {
		return err
	}

	switch st.Code() {
	case codes.NotFound:
		return ErrKeyNotFound
	case codes.Unavailable:
		return ErrConnectionFailed
	case codes.DeadlineExceeded:
		return ErrRequestTimeout
	default:
		return fmt.Errorf("rpc error (%s): %s", st.Code(), st.Message())
	}
}
