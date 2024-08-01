package qdrant

import (
	"context"
	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	"regexp"
	"time"
)

const (
	keyClaims = "claims"
)

func pathSign(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: "sign/" + framework.GenericNameRegex(keyRoleName),
		Fields: map[string]*framework.FieldSchema{
			keyRoleName: {
				Type:        framework.TypeLowerCaseString,
				Description: "Name of the role",
				Required:    true,
			},
			keyClaims: {
				Type:        framework.TypeMap,
				Description: `JSON claims set to sign.`,
				Required:    true,
			},
		},
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.UpdateOperation: &framework.PathOperation{
				Callback: b.pathSignWrite,
			},
		},
		HelpSynopsis:    pathSignHelpSyn,
		HelpDescription: pathSignHelpDesc,
	}
}

func (b *backend) pathSignWrite(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	roleName := d.Get("name").(string)

	role, err := b.getRole(ctx, req.Storage, roleName)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return logical.ErrorResponse("unknown role"), logical.ErrInvalidRequest
	}

	// Gather "freeform" claims
	rawClaims, ok := d.GetOk(keyClaims)
	if !ok {
		rawClaims = map[string]interface{}{}
	}

	claims, ok := rawClaims.(map[string]interface{})
	if !ok {
		return logical.ErrorResponse("claims not a map"), logical.ErrInvalidRequest
	}

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

	policy, err := b.getPolicy(ctx, req.Storage, config, req.MountPoint)
	if err != nil {
		return logical.ErrorResponse("error getting key: %v", err), err
	}

	signer := &PolicySigner{
		BackendId:          b.id,
		SignatureAlgorithm: config.SignatureAlgorithm,
		Policy:             policy,
		SignerOptions:      (&jose.SignerOptions{}).WithType("JWT"),
	}

	token, err := jwt.Signed(signer).Claims(claims).CompactSerialize()
	if err != nil {
		return logical.ErrorResponse("error serializing jwt: %v", err), err
	}

	resp := b.Secret(jwtSecretsTokenType).Response(
		map[string]interface{}{
			"token": token,
		},
		map[string]interface{}{},
	)
	resp.Secret.TTL = config.TokenTTL

	return resp, nil
}

const pathSignHelpSyn = `
Sign a set of claims.
`

const pathSignHelpDesc = `
Sign a set of claims.
`
