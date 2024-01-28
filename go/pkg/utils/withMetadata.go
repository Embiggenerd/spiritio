package utils

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
)

type ctxMetadataKey string

const Metadata_name ctxMetadataKey = "metadata"

// WithMetadata uses a map to hold arbitrary data for logging purposes
func WithMetadata(parent context.Context) context.Context {
	m := map[string]data{}
	md := &Metadata{metadata: m}

	parent = context.WithValue(parent, Metadata_name, md)
	return parent
}

func ExposeContextMetadata(ctx context.Context) MetadataInterface {
	metadata := ctx.Value(Metadata_name)
	md, _ := metadata.(*Metadata)
	return md
}

func (l *Metadata) Get(key string) (interface{}, error) {
	var err error
	value, ok := l.metadata[key]
	if !ok {
		err = errors.New("key not found")
	}
	return value.Value, err
}

func (l *Metadata) ToJSON() string {
	b, err := json.MarshalIndent(l.metadata, "", "  ")
	if err != nil {
		fmt.Println("error:", err)
	}
	return string(b)
}

func (l *Metadata) Set(key string, val interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	dx := data{Key: key, Value: val}
	l.metadata[key] = dx
}

type Metadata struct {
	metadata map[string]data
	mu       sync.Mutex
}

type MetadataInterface interface {
	Set(key string, val interface{})
	Get(key string) (interface{}, error)
	ToJSON() string
}

type data struct {
	Key   string
	Value interface{}
}
