package qdrant

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"github.com/go-jose/go-jose/v4"
	"github.com/hashicorp/vault/sdk/helper/errutil"
	"github.com/hashicorp/vault/sdk/helper/keysutil"
	"github.com/hashicorp/vault/sdk/logical"
	"time"
)

// Default values for configuration options.
const (
	DefaultSignatureAlgorithm = jose.ES256
	DefaultRSAKeyBits         = 2048
	DefaultKeyRotationPeriod  = "2h0m0s"
	DefaultTokenTTL           = "3m0s"
)

// Config holds all configuration for the backend.
type Config struct {

	// Connection string to Qdrant database
	URL string

	// API Key/ Sign key to sign and verify token
	SignKey string

	// SignatureAlgorithm is the signing algorithm to use.
	SignatureAlgorithm jose.SignatureAlgorithm

	// RSAKeyBits is size of generated RSA keys (Only for RSA)
	RSAKeyBits int

	// KeyRotationPeriod is how frequently a new key is created.
	KeyRotationPeriod time.Duration

	// TokenTTL defines how long a token is valid for after being signed.
	TokenTTL time.Duration
}

func (b *backend) getConfig(ctx context.Context, stg logical.Storage) (*Config, error) {
	b.cachedConfigLock.RLock()
	if b.cachedConfig != nil {
		defer b.cachedConfigLock.RUnlock()
		return b.cachedConfig.copy(), nil
	}

	b.cachedConfigLock.RUnlock()
	b.cachedConfigLock.Lock()
	defer b.cachedConfigLock.Unlock()

	// Double check somebody else didn't already cache it
	if b.cachedConfig != nil {
		return b.cachedConfig.copy(), nil
	}

	// Attempt to load config from storage & cache
	rawConfig, err := stg.Get(ctx, "config")
	if err != nil {
		return nil, err
	}

	if rawConfig != nil {
		// Found it, finish load from storage
		conf := &Config{}
		if err := json.Unmarshal(rawConfig.Value, conf); err == nil {
			b.cachedConfig = conf.cache()
		} else {
			b.Logger().Warn("Failed to unmarshal config, resetting to default")
		}
	}
	if b.cachedConfig == nil {
		// Nothing found, initialize configuration to default and save
		b.cachedConfig = DefaultConfig(b.System())
		if err := b.saveConfigUnlocked(ctx, stg, b.cachedConfig); err != nil {
			return nil, err
		}

		b.Logger().Debug("Config Initialized")
	}

	return b.cachedConfig.copy(), nil
}

func (c *Config) copy() *Config {
	cc := *c
	return &cc
}

func (b *backend) saveConfig(ctx context.Context, stg logical.Storage, config *Config, mount string) error {
	b.cachedConfigLock.Lock()
	defer b.cachedConfigLock.Unlock()

	keyFormatChanged := b.cachedConfig != nil

	if err := b.saveConfigUnlocked(ctx, stg, config); err != nil {
		return err
	}

	if !keyFormatChanged {
		return nil
	}

	b.Logger().Info("Key Format Rotation")

	policy, err := b.getPolicy(ctx, stg, config, mount)
	if err != nil {
		return err
	}

	policy.Lock(true)
	defer policy.Unlock()

	switch config.SignatureAlgorithm {
	case jose.RS256, jose.RS384, jose.RS512:
		switch config.RSAKeyBits {
		case 2048:
			policy.Type = keysutil.KeyType_RSA2048
		case 3072:
			policy.Type = keysutil.KeyType_RSA3072
		case 4096:
			policy.Type = keysutil.KeyType_RSA4096
		default:
			err = errutil.InternalError{Err: "unsupported RSA key size"}
		}
	case jose.ES256:
		policy.Type = keysutil.KeyType_ECDSA_P256
	case jose.ES384:
		policy.Type = keysutil.KeyType_ECDSA_P384
	case jose.ES512:
		policy.Type = keysutil.KeyType_ECDSA_P521
	default:
		err = errutil.InternalError{Err: "unknown/unsupported signature algorithm"}
	}

	if err != nil {
		return nil
	}

	defer b.lockManager.InvalidatePolicy(mainKeyName)

	return policy.Rotate(ctx, stg, rand.Reader)
}

func (b *backend) saveConfigUnlocked(ctx context.Context, stg logical.Storage, config *Config) error {

	entry, err := logical.StorageEntryJSON(configPath, config)
	if err != nil {
		return err
	}
	if err := stg.Put(ctx, entry); err != nil {
		return err
	}

	b.cachedConfig = config.cache()

	return nil
}

func (b *backend) clearConfig(ctx context.Context, stg logical.Storage) error {
	b.cachedConfigLock.Lock()
	defer b.cachedConfigLock.Unlock()

	if err := stg.Delete(ctx, configPath); err != nil {
		return err
	}

	b.cachedConfig = nil

	return nil
}

func (c *Config) cache() *Config {
	return c
}

// DefaultConfig returns a default configuration.
func DefaultConfig(sys logical.SystemView) *Config {
	defaultKeyRotationPeriod, _ := time.ParseDuration(DefaultKeyRotationPeriod)
	defaultTokenTTL, _ := time.ParseDuration(DefaultTokenTTL)

	c := &Config{}
	c.SignatureAlgorithm = DefaultSignatureAlgorithm
	c.RSAKeyBits = DefaultRSAKeyBits
	c.KeyRotationPeriod = defaultKeyRotationPeriod
	c.TokenTTL = durationMin(defaultTokenTTL, sys.DefaultLeaseTTL())
	return c
}
