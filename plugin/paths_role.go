package qdrant

import (
	"context"
	"encoding/json"
    "errors"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

const (
	rolePath   = "role"
	rolePrefix = "role/"
)


type RoleParameters struct {
	DBId               string `json:"dbId"`
    RoleId             string `json:"role"`
	Claims             map[string]interface{} `json:"claims"`
}

func pathRole(b *QdrantBackend) []*framework.Path {
	return []*framework.Path{
		{
			Pattern: rolePrefix + framework.GenericNameRegex("dbId") + "/" + framework.GenericNameRegex("role")+ "$",
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

				"claims": {
					Type:        framework.TypeMap,
					Description: `JSON claims set to sign.`,
                    Required: true,
				},
			},
			Operations: map[logical.Operation]framework.OperationHandler{
				logical.CreateOperation: &framework.PathOperation{
					Callback: b.pathAddRole,
				},
				logical.UpdateOperation: &framework.PathOperation{
					Callback: b.pathAddRole,
				},
				logical.ReadOperation: &framework.PathOperation{
					Callback: b.pathReadRole,
				},
				logical.DeleteOperation: &framework.PathOperation{
					Callback: b.pathDeleteRole,
				},
			},
			HelpSynopsis:    pathRoleHelpSyn,
			HelpDescription: pathRoleHelpDesc,
		},
		{
			Pattern: rolePrefix + framework.GenericNameRegex("dbId") + "?$",
			Fields: map[string]*framework.FieldSchema{

				"dbId": {
					Type:        framework.TypeString,
					Description: "DB identifier",
					Required:    false,
				},
            },
	
			Operations: map[logical.Operation]framework.OperationHandler{
				logical.ListOperation: &framework.PathOperation{
					Callback: b.pathListRole,
				},
			},
			HelpSynopsis:    pathRoleHelpSyn,
			HelpDescription: pathRoleHelpDesc,
		},
	}

}

func (b *QdrantBackend) pathAddRole(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	err := data.Validate()
	if err != nil {
		return logical.ErrorResponse(InvalidParametersError), logical.ErrInvalidRequest
	}

	jsonString, err := json.Marshal(data.Raw)

	b.Logger().Debug("pathAddRole", jsonString)

	if err != nil {
		return logical.ErrorResponse(DecodeFailedError), logical.ErrInvalidRequest
	}
	params := RoleParameters{}
	json.Unmarshal(jsonString, &params)

	err = b.addRole(ctx, req.Storage, params)

	if err != nil {
		return logical.ErrorResponse(AddingRoleFailedError + ":" + err.Error()), nil
	}
	return nil, nil
}

func (b *QdrantBackend) pathReadRole(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {

	err := data.Validate()
	if err != nil {
		return logical.ErrorResponse(InvalidParametersError), logical.ErrInvalidRequest
	}

	jsonString, err := json.Marshal(data.Raw)
	if err != nil {
		return logical.ErrorResponse(DecodeFailedError), logical.ErrInvalidRequest
	}
	params := RoleParameters{}
	json.Unmarshal(jsonString, &params)

	role, err := readRole(ctx, req.Storage, params.DBId, params.RoleId)

	if err != nil {
		return logical.ErrorResponse(ReadingRoleFailedError), nil
	}

	if role == nil {
		return logical.ErrorResponse(RoleNotFoundError), nil
	}

	return createResponseRole(role)

}

func (b *QdrantBackend) pathListRole(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {

	err := data.Validate()
	if err != nil {
		return logical.ErrorResponse(InvalidParametersError), logical.ErrInvalidRequest
	}

	jsonString, err := json.Marshal(data.Raw)
	if err != nil {
		return logical.ErrorResponse(DecodeFailedError), logical.ErrInvalidRequest
	}
	params := RoleParameters{}
	json.Unmarshal(jsonString, &params)

	b.Logger().Debug("list role path", rolePrefix + params.DBId)

	entries, err := listRole(ctx, req.Storage, params.DBId)
	if err != nil {
		return logical.ErrorResponse(ListRoleFailedError), nil
	}

	return logical.ListResponse(entries), nil
}

func (b *QdrantBackend) pathDeleteRole(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {

	err := data.Validate()
	if err != nil {
		return logical.ErrorResponse(InvalidParametersError), logical.ErrInvalidRequest
	}

	jsonString, err := json.Marshal(data.Raw)
	if err != nil {
		return logical.ErrorResponse(DecodeFailedError), logical.ErrInvalidRequest
	}
	params := RoleParameters{}
	json.Unmarshal(jsonString, &params)

	// delete role 
	err = deleteRole(ctx, req.Storage, params.DBId, params.RoleId)
	if err != nil {
		return logical.ErrorResponse(DeleteRoleFailedError), nil
	}
	return nil, nil

}

func (b *QdrantBackend) addRole(ctx context.Context, storage logical.Storage, params RoleParameters) error {

	path := rolePrefix + params.DBId + "/" + params.RoleId

	b.Logger().Debug("add role path", path)

	config, err := readConfig(ctx, storage, params.DBId)

	if err != nil {
		return err
	}

    if config == nil {
		return errors.New(ConfigNotFoundError)
    }

	err = storeInStorage[RoleParameters](ctx, storage, path, &params)

	if err != nil {
		return err
	}

	return nil

}

func readRole(ctx context.Context, storage logical.Storage, dbId string, role string) (*RoleParameters, error) {
	path := rolePrefix + dbId + "/" + role

	return getFromStorage[RoleParameters](ctx, storage, path)
}

func listRole(ctx context.Context, storage logical.Storage, dbId string) ([]string, error) {

	path := rolePrefix + dbId + "/"

	l, err := storage.List(ctx, path)

	if err != nil {
		return nil, err
	}
	var roles []string
	for _, v := range l {
		roles = append(roles, v)
	}
	return roles, nil
}

func deleteRole(ctx context.Context, storage logical.Storage, dbId string, role string) error {
	// get stored signing keys
	config, err := readRole(ctx, storage, dbId, role)
	if err != nil {
		return err
	}
	if config == nil {
		// nothing to delete
		return nil
	}

    path := rolePrefix + dbId + "/" + role

	return deleteFromStorage(ctx, storage, path)
}

func createResponseRole(role *RoleParameters) (*logical.Response, error) {

	rval := map[string]interface{}{}
	err := StructToMap(role, &rval)
	if err != nil {
		return nil, err
	}

	resp := &logical.Response{
		Data: rval,
	}
	return resp, nil
}

const pathRoleHelpSyn = `
Configure the roles.
`

const pathRoleHelpDesc = `
Configure the roles.

role:              Role name.
claims:            JSON claims.
`