package qdrant

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/assert"
)

func TestCRUDJWT(t *testing.T) {

	b, reqStorage := getTestBackend(t)

	t.Run("Test jwt", func(t *testing.T) {

		var current JWTParameters

		pathConfig := "config/instance1"
		pathRole1 := "role/instance1/write"
		claimsRole1 := `
            {
                "claims":{
                    "access": "w"
                }
            }`

		// first create config
		resp, err := b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.CreateOperation,
			Path:      pathConfig,
			Storage:   reqStorage,
			Data: map[string]interface{}{
				"url":     "http://localhost:6333",
				"sig_key": "your-very-long-256-bit-secret-key",
				"sig_alg": "HS256",
				"jwt_ttl": "3s",
			},
		})
		assert.NoError(t, err)
		assert.False(t, resp.IsError())

		// create role write
		var claims map[string]interface{}
		json.Unmarshal([]byte(claimsRole1), &claims)

		resp, err = b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.CreateOperation,
			Path:      pathRole1,
			Storage:   reqStorage,
			Data:      claims,
		})
		//t.Log(err, resp)
		assert.NoError(t, err)
		assert.False(t, resp.IsError())

		// call read
		resp, err = b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.ReadOperation,
			Path:      "jwt/instance1/write",
			Storage:   reqStorage,
		})

		t.Log(err, resp)

		MapToStruct(resp.Data, &current)

		assert.NoError(t, err)
		assert.False(t, resp.IsError())

		assert.NotEqual(t, current.Token, "")

	})
}
