package qdrant

import (
	"context"
	"fmt"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	"path"
)

const (
	keyStorageRolePath = "role"
	keyRoleName        = "name"
	keyClaims          = "claims"
)

type Role struct {

	// Claims defines claim values to be set on the issued JWT; each claim must be allowed by the plugin config.
	Claims map[string]interface{} `json:"claims"`
}

// Return response data for a role
func (r *Role) toResponseData() map[string]interface{} {
	respData := map[string]interface{}{
		keyClaims: r.Claims,
	}
	return respData
}

func pathRole(b *backend) []*framework.Path {
	return []*framework.Path{
		{
			Pattern: "roles/" + framework.GenericNameRegex(keyRoleName),
			Fields: map[string]*framework.FieldSchema{
				keyRoleName: {
					Type:        framework.TypeLowerCaseString,
					Description: `Specifies the name of the role to create. This is part of the request URL.`,
					Required:    true,
				},
				keyClaims: {
					Type:        framework.TypeMap,
					Description: `Claims to be set on issued JWTs. Each claim must be allowed by the configuration.`,
					Required:    true,
				},
			},
			Operations: map[logical.Operation]framework.OperationHandler{
				logical.ReadOperation: &framework.PathOperation{
					Callback: b.pathRolesRead,
				},
				logical.CreateOperation: &framework.PathOperation{
					Callback: b.pathRolesWrite,
				},
				logical.UpdateOperation: &framework.PathOperation{
					Callback: b.pathRolesWrite,
				},
				logical.DeleteOperation: &framework.PathOperation{
					Callback: b.pathRolesDelete,
				},
			},
			ExistenceCheck:  b.pathRoleExistenceCheck,
			HelpSynopsis:    pathRoleHelpSyn,
			HelpDescription: pathRoleHelpDesc,
		},
		{
			Pattern: "roles/?$",
			Operations: map[logical.Operation]framework.OperationHandler{
				logical.ListOperation: &framework.PathOperation{
					Callback: b.pathRolesList,
				},
			},
			HelpSynopsis:    pathRoleListHelpSyn,
			HelpDescription: pathRoleListHelpDesc,
		},
	}
}

func (b *backend) pathRoleExistenceCheck(ctx context.Context, req *logical.Request, d *framework.FieldData) (bool, error) {
	name := d.Get(keyRoleName).(string)

	role, err := req.Storage.Get(ctx, path.Join(keyStorageRolePath, name))
	if err != nil {
		return false, err
	}

	return role != nil, nil
}

// pathRolesList makes a request to Vault storage to retrieve a list of roles for the backend
func (b *backend) pathRolesList(ctx context.Context, req *logical.Request, _ *framework.FieldData) (*logical.Response, error) {
	entries, err := req.Storage.List(ctx, keyStorageRolePath+"/")
	if err != nil {
		return nil, err
	}

	return logical.ListResponse(entries), nil
}

// pathRolesRead makes a request to Vault storage to read a role and return response data
func (b *backend) pathRolesRead(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	role, err := b.getRole(ctx, req.Storage, d.Get(keyRoleName).(string))
	if err != nil {
		return nil, err
	}

	if role == nil {
		return nil, nil
	}

	return &logical.Response{
		Data: role.toResponseData(),
	}, nil
}

// pathRolesWrite makes a request to Vault storage to update a role based on the attributes passed to the role configuration
func (b *backend) pathRolesWrite(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	name, ok := d.GetOk(keyRoleName)
	if !ok {
		return logical.ErrorResponse("missing role name"), nil
	}

	role, err := b.getRole(ctx, req.Storage, name.(string))
	if err != nil {
		return nil, err
	}

	if role == nil {
		role = &Role{}
	}

	if newClaims, ok := d.GetOk(keyClaims); ok {
		role.Claims = newClaims.(map[string]interface{})
	}

	if err := b.setRole(ctx, req.Storage, name.(string), role); err != nil {
		return nil, err
	}

	return nil, nil
}

// pathRolesDelete makes a request to Vault storage to delete a role
func (b *backend) pathRolesDelete(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	err := req.Storage.Delete(ctx, path.Join(keyStorageRolePath, d.Get(keyRoleName).(string)))
	if err != nil {
		return nil, fmt.Errorf("error deleting role: %w", err)
	}
	return nil, nil
}

// getRole gets the role from the Vault storage API
func (b *backend) getRole(ctx context.Context, stg logical.Storage, name string) (*Role, error) {
	if name == "" {
		return nil, fmt.Errorf("missing role name")
	}

	entry, err := stg.Get(ctx, path.Join(keyStorageRolePath, name))
	if err != nil {
		return nil, err
	}

	if entry == nil {
		return nil, nil
	}

	var role Role

	if err := entry.DecodeJSON(&role); err != nil {
		return nil, err
	}
	return &role, nil
}

// setRole adds the role to the Vault storage API
func (b *backend) setRole(ctx context.Context, stg logical.Storage, name string, role *Role) error {
	entry, err := logical.StorageEntryJSON(path.Join(keyStorageRolePath, name), role)
	if err != nil {
		return err
	}

	if entry == nil {
		return fmt.Errorf("failed to create storage entry for role")
	}

	if err := stg.Put(ctx, entry); err != nil {
		return err
	}

	return nil
}

const pathRoleHelpSyn = `
Manages Vault role for generating tokens.
`

const pathRoleHelpDesc = `
Manages Vault role for generating tokens.
`

const pathRoleListHelpSyn = `
This endpoint returns a list of available roles.
`

const pathRoleListHelpDesc = `
This endpoint returns a list of available roles. Only the role names are returned, not any values.
`
