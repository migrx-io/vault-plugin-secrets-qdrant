package qdrant

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/go-jose/go-jose/v4"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

const (
	configPath   = "config"
	configPrefix = "config/"
)

type ConfigParameters struct {
	DBId               string                  `json:"dbId"`
	URL                string                  `json:"url"`
	SignKey            string                  `json:"sig_Key"`
	SignatureAlgorithm jose.SignatureAlgorithm `json:"sig_alg,omitempty"`
	RSAKeyBits         int                     `json:"rsa_key_bits,omitempty"`
	TokenTTL           string                  `json:"jwt_ttl,omitempty"`
}

func pathConfig(b *QdrantBackend) []*framework.Path {
	return []*framework.Path{
		{
			Pattern: configPrefix + framework.GenericNameRegex("dbId") + "$",
			Fields: map[string]*framework.FieldSchema{

				"dbId": {
					Type:        framework.TypeString,
					Description: "DB identifier",
					Required:    false,
				},
				"url": {
					Type:        framework.TypeString,
					Description: `Connection string to Qdrant database`,
					Required:    true,
				},

				"sig_Key": {
					Type:        framework.TypeString,
					Description: `API Key/ Sign key to sign and verify token`,
					Required:    true,
				},

				"sig_alg": {
					Type:        framework.TypeString,
					Description: `Signature algorithm used to sign new tokens.`,
				},

				"rsa_key_bits": {
					Type:        framework.TypeInt,
					Description: `Size of generated RSA keys, when signature algorithm is one of the allowed RSA signing algorithm.`,
				},
				"jwt_ttl": {
					Type:        framework.TypeString,
					Description: `Duration a token is valid for (mapped to the 'exp' claim).`,
				},
			},
			Operations: map[logical.Operation]framework.OperationHandler{
				logical.CreateOperation: &framework.PathOperation{
					Callback: b.pathAddConfig,
				},
				logical.UpdateOperation: &framework.PathOperation{
					Callback: b.pathAddConfig,
				},
				logical.ReadOperation: &framework.PathOperation{
					Callback: b.pathReadConfig,
				},
				logical.DeleteOperation: &framework.PathOperation{
					Callback: b.pathDeleteConfig,
				},
			},
			HelpSynopsis:    pathConfigHelpSyn,
			HelpDescription: pathConfigHelpDesc,
		},
		{
			Pattern: configPrefix + "?$",
			Operations: map[logical.Operation]framework.OperationHandler{
				logical.ListOperation: &framework.PathOperation{
					Callback: b.pathListConfig,
				},
			},
			HelpSynopsis:    pathConfigHelpSyn,
			HelpDescription: pathConfigHelpDesc,
		},
	}

}

func (b *QdrantBackend) pathAddConfig(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	err := data.Validate()
	if err != nil {
		return logical.ErrorResponse(InvalidParametersError), logical.ErrInvalidRequest
	}

	jsonString, err := json.Marshal(data.Raw)

	b.Logger().Debug("pathAddConfig", jsonString)

	if err != nil {
		return logical.ErrorResponse(DecodeFailedError), logical.ErrInvalidRequest
	}
	params := ConfigParameters{}
	json.Unmarshal(jsonString, &params)

	err = b.addConfig(ctx, req.Storage, params)
	if err != nil {
		return logical.ErrorResponse(AddingConfigFailedError + ":" + err.Error()), nil
	}
	return nil, nil
}

func (b *QdrantBackend) pathReadConfig(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {

	err := data.Validate()
	if err != nil {
		return logical.ErrorResponse(InvalidParametersError), logical.ErrInvalidRequest
	}

	jsonString, err := json.Marshal(data.Raw)
	if err != nil {
		return logical.ErrorResponse(DecodeFailedError), logical.ErrInvalidRequest
	}
	params := ConfigParameters{}
	json.Unmarshal(jsonString, &params)

	config, err := readConfig(ctx, req.Storage, params.DBId)

	if err != nil {
		return logical.ErrorResponse(ReadingConfigFailedError), nil
	}

	if config == nil {
		return logical.ErrorResponse(ConfigNotFoundError), nil
	}

	return createResponseConfig(config)

}

func (b *QdrantBackend) pathListConfig(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {

	err := data.Validate()
	if err != nil {
		return logical.ErrorResponse(InvalidParametersError), logical.ErrInvalidRequest
	}

	entries, err := listConfig(ctx, req.Storage)
	if err != nil {
		return logical.ErrorResponse(ListConfigFailedError), nil
	}

	return logical.ListResponse(entries), nil
}

func (b *QdrantBackend) pathDeleteConfig(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {

	err := data.Validate()
	if err != nil {
		return logical.ErrorResponse(InvalidParametersError), logical.ErrInvalidRequest
	}

	jsonString, err := json.Marshal(data.Raw)
	if err != nil {
		return logical.ErrorResponse(DecodeFailedError), logical.ErrInvalidRequest
	}
	params := ConfigParameters{}
	json.Unmarshal(jsonString, &params)

	// delete issue and all related nkeys and jwt
	err = deleteConfig(ctx, req.Storage, params)
	if err != nil {
		return logical.ErrorResponse(DeleteConfigFailedError), nil
	}
	return nil, nil

}

func (b *QdrantBackend) addConfig(ctx context.Context, storage logical.Storage, params ConfigParameters) error {

	path := configPrefix + params.DBId

	b.Logger().Debug("add Config path", path)

	//config, err := getFromStorage[ConfigParameters](ctx, storage, path)
	//if err != nil {
	//	return nil, err
	//}

	err := storeInStorage[ConfigParameters](ctx, storage, path, &params)

	if err != nil {
		return err
	}

	return nil

}

func readConfig(ctx context.Context, storage logical.Storage, dbId string) (*ConfigParameters, error) {
	path := configPrefix + dbId
	return getFromStorage[ConfigParameters](ctx, storage, path)
}

func listConfig(ctx context.Context, storage logical.Storage) ([]string, error) {
	path := configPrefix

	l, err := storage.List(ctx, path)

	if err != nil {
		return nil, err
	}
	var configs []string
	for _, v := range l {
		configs = append(configs, v)
	}
	return configs, nil
}

func deleteConfig(ctx context.Context, storage logical.Storage, params ConfigParameters) error {
	// get stored signing keys
	config, err := readConfig(ctx, storage, params.DBId)
	if err != nil {
		return err
	}
	if config == nil {
		// nothing to delete
		return nil
	}

	// delete all associated roles
	entries, err := listRole(ctx, storage, params.DBId)
	if err != nil {
		return errors.New(ListRoleFailedError)
	}

	for _, v := range entries {
		deleteRole(ctx, storage, params.DBId, v)
	}

	// delete config
	path := configPrefix + params.DBId
	return deleteFromStorage(ctx, storage, path)
}

func createResponseConfig(config *ConfigParameters) (*logical.Response, error) {

	rval := map[string]interface{}{}
	err := StructToMap(config, &rval)
	if err != nil {
		return nil, err
	}

	resp := &logical.Response{
		Data: rval,
	}
	return resp, nil
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
jwt_ttl:          Duration before a token expires.
`
