package qdrant

import (
	"context"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/assert"
)

func TestCRUDConfig(t *testing.T) {

	b, reqStorage := getTestBackend(t)

	t.Run("Test initial state of config", func(t *testing.T) {

		path := "config/instance1"

		var current ConfigParameters
		var expected ConfigParameters

		// first create config
		resp, err := b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.CreateOperation,
			Path:      path,
			Storage:   reqStorage,
			Data: map[string]interface{}{
				"url":     "localhost:6334",
				"sig_key": "secret",
				"jwt_ttl": "3s",
                "tls": true,
                "ca": "",
			},
		})
		assert.NoError(t, err)
		assert.False(t, resp.IsError())

		// list all instances

		resp, err = b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.ListOperation,
			Path:      "config",
			Storage:   reqStorage,
		})
		assert.NoError(t, err)
		assert.False(t, resp.IsError())
		assert.Equal(t, resp.Data, map[string]interface{}{
			"keys": []string{"instance1"},
		})

		// call read
		resp, err = b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.ReadOperation,
			Path:      path,
			Storage:   reqStorage,
		})

		MapToStruct(resp.Data, &current)

		expected = ConfigParameters{
			DBId:     "instance1",
			URL:      "localhost:6334",
			SignKey:  "secret",
			TokenTTL: "3s",
            TLS: true,
            CA: "",
		}

		assert.NoError(t, err)
		assert.False(t, resp.IsError())

		assert.Equal(t, expected, current)

		// call delete
		resp, err = b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.DeleteOperation,
			Path:      path,
			Storage:   reqStorage,
		})
		assert.NoError(t, err)
		assert.False(t, resp.IsError())

		// call list
		resp, err = b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.ListOperation,
			Path:      "config",
			Storage:   reqStorage,
		})
		assert.NoError(t, err)
		assert.False(t, resp.IsError())
		assert.Equal(t, resp.Data, map[string]interface{}{})

	})
}
