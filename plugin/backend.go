package qdrant

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"

)

// QdrantBackend defines an object that
// extends the Vault backend and stores the
// target API's client.
type QdrantBackend struct {
	*framework.Backend
    clientMutex     sync.RWMutex
	client *QdrantClient
}

func Factory(ctx context.Context, conf *logical.BackendConfig) (logical.Backend, error) {
	b := backend()
	if err := b.Setup(ctx, conf); err != nil {
		return nil, err
	}
	return b, nil
}

// backend defines the target API backend
// for Vault. It must include each path
// and the secrets it will store.
func backend() *QdrantBackend {
	var b = QdrantBackend{}

	b.Backend = &framework.Backend{
		Help: strings.TrimSpace(backendHelp),
		PathsSpecial: &logical.Paths{
			LocalStorage: []string{},
			SealWrapStorage: []string{
				"config",
				"role/*",
			},
		},
		Paths: framework.PathAppend(
			pathConfig(&b),
			pathRole(&b),
			pathJWT(&b),
		),
		Secrets: []*framework.Secret{
			// b.hashiCupsToken(),
		},
		BackendType: logical.TypeLogical,
		Invalidate:  b.invalidate,
	}
	return &b
}

// backendHelp should contain help information for the backend
const backendHelp = `
The Qdrant secrets backend dynamically generates user tokens.
`

// reset clears any client configuration for a new
// backend to be configured
func (b *QdrantBackend) reset() {
	b.clientMutex.Lock()
	defer b.clientMutex.Unlock()
	b.client = nil
}

// invalidate clears an existing client configuration in
// the backend
func (b *QdrantBackend) invalidate(ctx context.Context, key string) {
	if key == "config" {
		b.reset()
	}
}

func getFromStorage[T any](ctx context.Context, s logical.Storage, path string) (*T, error) {
	if path == "" {
		return nil, fmt.Errorf("missing path")
	}

	// get data entry from storage backend
	entry, err := s.Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("error retrieving Data: %w", err)
	}

	if entry == nil {
		return nil, nil
	}

	// convert json data to T
	var t T
	if err := entry.DecodeJSON(&t); err != nil {
		return nil, fmt.Errorf("error decoding data: %w", err)
	}
	return &t, nil
}

func deleteFromStorage(ctx context.Context, s logical.Storage, path string) error {
	if err := s.Delete(ctx, path); err != nil {
		return fmt.Errorf("error deleting data: %w", err)
	}
	return nil
}

func storeInStorage[T any](ctx context.Context, s logical.Storage, path string, t *T) error {
	entry, err := logical.StorageEntryJSON(path, t)
	if err != nil {
		return err
	}

	if err := s.Put(ctx, entry); err != nil {
		return err
	}

	return nil
}

func readOperation[T any](ctx context.Context, s logical.Storage, path string) (*logical.Response, error) {
	t, err := getFromStorage[T](ctx, s, path)
	if err != nil {
		return nil, err
	}

	if t == nil {
		return nil, nil
	}

	var groupMap map[string]interface{}
	err = StructToMap(t, &groupMap)
	if err != nil {
		return nil, err
	}

	return &logical.Response{
		Data: groupMap,
	}, nil
}

func (b *QdrantBackend) periodicFunc(ctx context.Context, sys *logical.Request) error {
	b.Logger().Debug("Periodic: starting periodic func")
	return nil
}
