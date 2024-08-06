package qdrant

import (
	"context"
	"encoding/json"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

const (
	jwtPath   = "jwt"
	jwtPrefix = "jwt/"
)


type JWTParameters struct {
	DBId               string `json:"dbId"`
    RoleId             string `json:"role"`
    Token              string `json:"token"`
}

func pathJWT(b *QdrantBackend) []*framework.Path {
	return []*framework.Path{
		{
			Pattern: jwtPrefix + framework.GenericNameRegex("dbId") + "/" + framework.GenericNameRegex("role")+ "?$",
			Fields: map[string]*framework.FieldSchema{

				"dbId": {
					Type:        framework.TypeString,
					Description: "DB identifier",
					Required:    false,
				},
				"role": {
					Type:        framework.TypeString,
					Description: "Role name",
					Required:    false,
				},
			},
			Operations: map[logical.Operation]framework.OperationHandler{
				logical.ReadOperation: &framework.PathOperation{
					Callback: b.pathReadJWT,
				},
			},
			HelpSynopsis:    pathJWTHelpSyn,
			HelpDescription: pathJWTHelpDesc,
		},
	}

}

func (b *QdrantBackend) pathReadJWT(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {

	err := data.Validate()
	if err != nil {
		return logical.ErrorResponse(InvalidParametersError), logical.ErrInvalidRequest
	}

	jsonString, err := json.Marshal(data.Raw)
	if err != nil {
		return logical.ErrorResponse(DecodeFailedError), logical.ErrInvalidRequest
	}
	params := JWTParameters{}
	json.Unmarshal(jsonString, &params)

    // get config
	config, err := readConfig(ctx, req.Storage, params.DBId)

	if err != nil {
		return logical.ErrorResponse(ReadingConfigFailedError), nil
	}

    if config == nil {
		return logical.ErrorResponse(ConfigNotFoundError), nil
	}

    // get role
	role, err := readRole(ctx, req.Storage, params.DBId, params.RoleId)

	if err != nil {
		return logical.ErrorResponse(ReadingRoleFailedError), nil
	}

	if role == nil {
		return logical.ErrorResponse(RoleNotFoundError), nil
	}
    // Generate JWT token
    params.Token = "secret"

	return createResponseJWT(&params)

}

func createResponseJWT(token *JWTParameters) (*logical.Response, error) {

	rval := map[string]interface{}{}
	err := StructToMap(token, &rval)
	if err != nil {
		return nil, err
	}

	resp := &logical.Response{
		Data: rval,
	}
	return resp, nil
}

const pathJWTHelpSyn = `
Generate JWT Token.
`

const pathJWTHelpDesc = `
Generate JWT Token

dbId              Instance Id
role:             Role name.
token:            JWT Token.
`
