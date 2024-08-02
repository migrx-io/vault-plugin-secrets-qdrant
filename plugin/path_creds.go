package qdrant

import (
	"context"
	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	"time"
)

func pathCreds(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: "creds/" + framework.GenericNameRegex(keyRoleName),
		Fields: map[string]*framework.FieldSchema{
			keyRoleName: {
				Type:        framework.TypeLowerCaseString,
				Description: "Name of the role",
				Required:    true,
			},
		},
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.ReadOperation: &framework.PathOperation{
				Callback: b.pathCredsRead,
			},
		},
		HelpSynopsis:    pathCredsHelpSyn,
		HelpDescription: pathCredsHelpDesc,
	}
}

func (b *backend) pathCredsRead(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	roleName := d.Get("name").(string)

	role, err := b.getRole(ctx, req.Storage, roleName)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return logical.ErrorResponse("unknown role"), logical.ErrInvalidRequest
	}

    claims := map[string]interface{}{}

	config, err := b.getConfig(ctx, req.Storage)
	if err != nil {
		return nil, err
	}

	for roleClaim := range role.Claims {
		claims[roleClaim] = role.Claims[roleClaim]
	}

	now := time.Now()

	expiry := now.Add(config.TokenTTL)
	claims["exp"] = jwt.NumericDate(expiry.Unix())

	// Get key and issue JWT
	key := []byte(config.SignKey)
	sig, err := jose.NewSigner(jose.SigningKey{Algorithm: config.SignatureAlgorithm, Key: key}, (&jose.SignerOptions{}).WithType("JWT"))

	if err != nil {
		return logical.ErrorResponse("error making signer jwt: %v", err), err
	}

	token, err := jwt.Signed(sig).Claims(claims).Serialize()
	if err != nil {
		return logical.ErrorResponse("error serializing jwt: %v", err), err
	}

	// TODO send data to database

	resp := b.Secret(jwtSecretsTokenType).Response(
		map[string]interface{}{
			"token": token,
		},
		map[string]interface{}{},
	)
	resp.Secret.TTL = config.TokenTTL

	return resp, nil
}

const pathCredsHelpSyn = `
Generate JWT token. 
`

const pathCredsHelpDesc = `
Generate JWT token. 
`
