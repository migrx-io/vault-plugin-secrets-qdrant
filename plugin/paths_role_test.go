package qdrant

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/assert"
)

func TestCRUDRole(t *testing.T) {

	b, reqStorage := getTestBackend(t)

	t.Run("Test roles", func(t *testing.T) {

		pathConfig := "config/instance1"
		pathNotConfigRole := "role/noinstance/admin"
		pathRole1 := "role/instance1/write"
		claimsRole1 := `
            {
                "claims":{
                    "access": "w"
                }
            }`
		pathRole2 := "role/instance1/read"
		claimsRole2 := `
            {
                "claims":{
                    "value_exists": {
                        "collection": "users",
                        "matches": [
                        { "key": "role", "value": "admin" }
                        ]
                    },
                    "access": [
                        {
                        "collection": "my_collection",
                        "access": "rw"
                        }
                    ]
                }
            }`

		var current RoleParameters
		var expected RoleParameters

		// first create config
		resp, err := b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.CreateOperation,
			Path:      pathConfig,
			Storage:   reqStorage,
			Data: map[string]interface{}{
				"url":          "http://localhost:6333",
				"sig_key":      "secret",
				"sig_alg":      "RSA256",
				"jwt_ttl":      "3s",
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

		// create role read
		json.Unmarshal([]byte(claimsRole2), &claims)

		resp, err = b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.CreateOperation,
			Path:      pathRole2,
			Storage:   reqStorage,
			Data:      claims,
		})
		assert.NoError(t, err)
		assert.False(t, resp.IsError())

		// list all roles
		resp, err = b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.ListOperation,
			Path:      "role/instance1",
			Storage:   reqStorage,
		})
		//t.Log(err, resp.Data)
		assert.NoError(t, err)
		assert.False(t, resp.IsError())
		assert.Equal(t, resp.Data, map[string]interface{}{
			"keys": []string{"read", "write"},
		})

		// create role for non existing instance
		json.Unmarshal([]byte(claimsRole1), &claims)

		resp, err = b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.CreateOperation,
			Path:      pathNotConfigRole,
			Storage:   reqStorage,
			Data:      claims,
		})
		//t.Log(err, resp)
		assert.NoError(t, err)
		assert.True(t, resp.IsError())

		// call read
		resp, err = b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.ReadOperation,
			Path:      pathRole1,
			Storage:   reqStorage,
		})

		MapToStruct(resp.Data, &current)
		json.Unmarshal([]byte(claimsRole1), &claims)

		expected = RoleParameters{
			DBId:   "instance1",
			RoleId: "write",
			Claims: claims["claims"].(map[string]interface{}),
		}

		assert.NoError(t, err)
		assert.False(t, resp.IsError())

		assert.Equal(t, expected, current)

		// call delete
		resp, err = b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.DeleteOperation,
			Path:      pathRole1,
			Storage:   reqStorage,
		})
		assert.NoError(t, err)
		assert.False(t, resp.IsError())

		// call list
		resp, err = b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.ListOperation,
			Path:      "role/instance1",
			Storage:   reqStorage,
		})

		json.Unmarshal([]byte(claimsRole2), &claims)

		assert.NoError(t, err)
		assert.False(t, resp.IsError())
		assert.Equal(t, resp.Data, map[string]interface{}{"keys": []string{"read"}})

		// delete instance
		// call delete
		resp, err = b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.DeleteOperation,
			Path:      "config/instance1",
			Storage:   reqStorage,
		})
		assert.NoError(t, err)
		assert.False(t, resp.IsError())

		// call list
		resp, err = b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.ListOperation,
			Path:      "role/instance1",
			Storage:   reqStorage,
		})

		json.Unmarshal([]byte(claimsRole2), &claims)

		assert.NoError(t, err)
		assert.False(t, resp.IsError())
		assert.Equal(t, resp.Data, map[string]interface{}{})

	})
}
