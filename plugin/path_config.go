package qdrant

import (
	"context"
	"github.com/go-jose/go-jose/v4"
	"time"
    "errors"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

const (
	keyURL                = "url"
	keyKey                = "key"
	keySignatureAlgorithm = "sig_alg"
	keyRSAKeyBits         = "rsa_key_bits"
	keyRotationDuration   = "key_ttl"
	keyTokenTTL           = "jwt_ttl"
)

func pathConfig(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: "config",
		Fields: map[string]*framework.FieldSchema{

			keyURL: {
				Type:        framework.TypeString,
				Description: `Connection string to Qdrant database`,
                Required:    true,
			},

			keyKey: {
				Type:        framework.TypeString,
				Description: `API Key/ Sign key to sign and verify token`,
                Required:    true,
			},

			keySignatureAlgorithm: {
				Type:        framework.TypeString,
				Description: `Signature algorithm used to sign new tokens.`,
			},
			keyRSAKeyBits: {
				Type:        framework.TypeInt,
				Description: `Size of generated RSA keys, when signature algorithm is one of the allowed RSA signing algorithm.`,
			},
			keyRotationDuration: {
				Type:        framework.TypeString,
				Description: `Duration a specific key will be used to sign new tokens.`,
			},
			keyTokenTTL: {
				Type:        framework.TypeString,
				Description: `Duration a token is valid for (mapped to the 'exp' claim).`,
			},
		},

		Operations: map[logical.Operation]framework.OperationHandler{
			logical.ReadOperation: &framework.PathOperation{
				Callback: b.pathConfigRead,
			},
			logical.CreateOperation: &framework.PathOperation{
				Callback: b.pathConfigWrite,
			},
			logical.UpdateOperation: &framework.PathOperation{
				Callback: b.pathConfigWrite,
			},
			logical.DeleteOperation: &framework.PathOperation{
				Callback: b.pathConfigDelete,
			},
		},

		ExistenceCheck:  b.pathConfigExistenceCheck,
		HelpSynopsis:    pathConfigHelpSyn,
		HelpDescription: pathConfigHelpDesc,
	}
}

func (b *backend) pathConfigExistenceCheck(ctx context.Context, req *logical.Request, _ *framework.FieldData) (bool, error) {
	savedConfig, err := req.Storage.Get(ctx, configPath)
	if err != nil {
		return false, err
	}

	return savedConfig != nil, nil
}

func (b *backend) pathConfigWrite(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	config, err := b.getConfig(ctx, req.Storage)
	if err != nil {
		return nil, err
	}

	url := d.Get("url").(string)

	if url == "" {
		return nil, errors.New("url is empty")
	}

	key := d.Get("key").(string)

	if key == "" {
		return nil, errors.New("key is empty")
	}

	if newRawSignatureAlgorithmName, ok := d.GetOk(keySignatureAlgorithm); ok {
		newSignatureAlgorithmName, ok := newRawSignatureAlgorithmName.(string)
		if !ok {
			return logical.ErrorResponse("sig_alg must be a string"), logical.ErrInvalidRequest
		}
		config.SignatureAlgorithm = jose.SignatureAlgorithm(newSignatureAlgorithmName)
	}

	if newRawRSAKeyBits, ok := d.GetOk(keyRSAKeyBits); ok {
		newRSAKeyBits, ok := newRawRSAKeyBits.(int)
		if !ok {
			return logical.ErrorResponse("rsa_key_bits must be an integer"), logical.ErrInvalidRequest
		}
		config.RSAKeyBits = newRSAKeyBits
	}

	if newRotationPeriod, ok := d.GetOk(keyRotationDuration); ok {
		duration, err := time.ParseDuration(newRotationPeriod.(string))
		if err != nil {
			return nil, err
		}
		config.KeyRotationPeriod = duration
	}

	if newTTL, ok := d.GetOk(keyTokenTTL); ok {
		duration, err := time.ParseDuration(newTTL.(string))
		if err != nil {
			return nil, err
		}
		config.TokenTTL = duration
	}

	if config.TokenTTL > b.System().MaxLeaseTTL() {
		return logical.ErrorResponse("'%s' is greater that the max lease ttl", keyTokenTTL), logical.ErrInvalidRequest
	}

	if err := b.saveConfig(ctx, req.Storage, config, req.MountPoint); err != nil {
		return nil, err
	}

	return configResponse(config)
}

func (b *backend) pathConfigRead(ctx context.Context, req *logical.Request, _ *framework.FieldData) (*logical.Response, error) {
	config, err := b.getConfig(ctx, req.Storage)
	if err != nil {
		return nil, err
	}

	return configResponse(config)
}

func (b *backend) pathConfigDelete(ctx context.Context, req *logical.Request, _ *framework.FieldData) (*logical.Response, error) {
	err := b.clearConfig(ctx, req.Storage)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func configResponse(config *Config) (*logical.Response, error) {
	return &logical.Response{
		Data: map[string]interface{}{
			keyURL:                config.URL,
			keyKey:                config.Key,
			keySignatureAlgorithm: config.SignatureAlgorithm,
			keyRSAKeyBits:         config.RSAKeyBits,
			keyRotationDuration:   config.KeyRotationPeriod.String(),
			keyTokenTTL:           config.TokenTTL.String(),
		},
	}, nil
}

const pathConfigHelpSyn = `
Configure the backend.
`

const pathConfigHelpDesc = `
Configure the backend.

url:              Connection string to Qdrant database.
key:              API Key/ Sign key to sign and verify token.             
sig_alg:		  Signature algorithm used to sign new tokens.
rsa_key_bits:	  Size of generate RSA keys, when using RSA signature algorithms.
key_ttl:          Duration before a key stops signing new tokens and a new one is generated.
		          After this period the public key will still be available to verify JWTs.
jwt_ttl:          Duration before a token expires.
`
