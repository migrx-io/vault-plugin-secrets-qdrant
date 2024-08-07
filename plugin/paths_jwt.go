package qdrant

import (
	"context"
	"encoding/json"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"time"
)

const (
	jwtPath   = "jwt"
	jwtPrefix = "jwt/"
)

type JWTParameters struct {
	DBId   string `json:"dbId"`
	RoleId string `json:"role"`
	Token  string `json:"token"`
}

func pathJWT(b *QdrantBackend) []*framework.Path {
	return []*framework.Path{
		{
			Pattern: jwtPrefix + framework.GenericNameRegex("dbId") + "/" + framework.GenericNameRegex("role") + "?$",
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
		return logical.ErrorResponse(BuildErrResponse(InvalidParametersError, err)), logical.ErrInvalidRequest
	}

	jsonString, err := json.Marshal(data.Raw)
	if err != nil {
		return logical.ErrorResponse(BuildErrResponse(DecodeFailedError, err)), logical.ErrInvalidRequest
	}
	params := JWTParameters{}
	json.Unmarshal(jsonString, &params)

	// get config
	config, err := readConfig(ctx, req.Storage, params.DBId)

	if err != nil {
		return logical.ErrorResponse(BuildErrResponse(ReadingConfigFailedError, err)), nil
	}

	if config == nil {
		return logical.ErrorResponse(BuildErrResponse(ConfigNotFoundError, err)), nil
	}

	// get role
	role, err := readRole(ctx, req.Storage, params.DBId, params.RoleId)

	if err != nil {
		return logical.ErrorResponse(BuildErrResponse(ReadingRoleFailedError, err)), nil
	}

	if role == nil {
		return logical.ErrorResponse(RoleNotFoundError), nil
	}
	// Generate JWT token
	err = b.generateJWT(config, role, &params)

	if err != nil {
		return logical.ErrorResponse(err.Error()), nil
	}

	return createResponseJWT(&params)

}

func (b *QdrantBackend) generateJWT(config *ConfigParameters, role *RoleParameters, jwt_token *JWTParameters) error {

	claims := role.Claims

	claims["iss"] = role.RoleId

	now := time.Now()

    var delta time.Duration

    if role.TokenTTL != ""{
	    delta, _ = time.ParseDuration(role.TokenTTL)
    }else{
	    delta, _ = time.ParseDuration(config.TokenTTL)
    }

	expiry := now.Add(delta)

	claims["exp"] = jwt.NumericDate(expiry.Unix())

	sig, err := jose.NewSigner(
		jose.SigningKey{
			Algorithm: jose.SignatureAlgorithm(config.SignatureAlgorithm),
			Key:       []byte(config.SignKey),
		},
		(&jose.SignerOptions{}).WithType("JWT"),
	)

	if err != nil {
		return err
	}

	token, err := jwt.Signed(sig).Claims(claims).Serialize()

	if err != nil {
		return err
	}

	jwt_token.Token = token

	return nil

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
