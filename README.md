# HashiCorp Vault Secrets Engine - Qdrant plugin

`vault-plugin-secrets-qdrant` is a Hashicorp Vault plugin that extends Vault with a secrets engine for [Qdrant](https://qdrant.tech) for JWT auth. 

It is capable of generating Qdrant credentials/JWT signed tokens with granular access control. 

The roles are stored in Vault and can be revoked at any time.

The generated JWT tokens are ephemeral and stateless; they are not stored in a vault but can be [bound to roles](https://qdrant.tech/documentation/guides/security/#granular-access-control-with-jwt) and invalidated when the role is deleted.

The plugin is also able to create/update/delete roles data to a Qdrant servers


## Features

- Support multi-instance configurations
- Allow management of Token TTL per instance and/or role
- Push role changes (create/update/delete) to Qdrant server
- Generate and sign JWT tokens based on instance and role parameters
- Allow provision of custom claims (access and filters) for roles
- Support TLS and custom CA to connect to Qdrant server

## Getting Started

The `Qdrant` secrets engine generates JWT credentials dynamically.

The plugin supports several resources, including: config, role and jwt.

Please read the official [Qdrant documentation](https://qdrant.tech/documentation/guides/security/#granular-access-control-with-jwt) to understand the concepts of token and access as well as the authentication process.

A hand full of resources can be defined within the vault plugin:

### Config

The resource of type `config` represent database instance configuration for secrets.


| Entity path                                                  | Description                    | Operations          |
| :----------------------------------------------------------- | :----------------------------- | :------------------ |
| qdrant/config                                                | List instances                 | list                |
| qdrant/config/<instance>                                     | Manage instance config         | write, read, delete |


### Role

The resource of type `role` represent database roles configuration for secrets.

| Entity path                                                  | Description                    | Operations          |
| :----------------------------------------------------------- | :----------------------------- | :------------------ |
| qdrant/role/<instance>                                       | List roles for <instance>      | list                |
| qdrant/role/<instance>/<role>                                | Manage instance role config    | write, read, delete |


### JWT

The resource of type `jwt` represent database JWT tokens.

| Entity path                                                  | Description                    | Operations          |
| :----------------------------------------------------------- | :----------------------------- | :------------------ |
| qdrant/jwt/<instance>/<role>                                 | Generate token for role        | read                |



## ⚙️ Configuration

There are arguments that can be passed to the paths for `config/` (database instance), `role/`.

### Config

| Key               | Type        | Required | Example     | Description                                                          |
| :---------------- | :---------- | :------- | :---------- | :------------------------------------------------------------------- |
| url               | bool        | true     | qdrant:6334 | URL address of Qdrant instance (grpc protocol)                       |
| sig_key           | string      | true     | secret-key  | Secret key to sign and verify(API-KEY server) tokens.                |
| sig_alg           | string      | true     | HS256       | Algorithm to decode the tokens.                                      |
| jwt_ttl           | string      | true     | 300s        | Default TTL for instance tokens (can be overwritten in roles)        |
| tls               | bool        | false    | true        | If set to true - vault will open tls grpc connection to Qdrant       |
| ca                | string      | false    | eyJhbGc...  | Base64 encoded custom CA cert for TLS                                |


**Note: When you delete an instance configuration, all associated roles will be automatically deleted from the Qdrant instance.**


### Role

| Key               | Type        | Required | Example     | Description                                                          |
| :---------------- | :---------- | :------- | :---------- | :------------------------------------------------------------------- |
| jwt_ttl           | string      | false    | 300s        | TTL for instance tokens                                              |
| claims            | json        | true     |             | Access and filters attributes (see Qdrant doc)                       |


**Note: Vault roles sync with Qdrant instance collection `sys_roles` automatically**


`claims` example

```

{
    "claims":{
        "value_exists": {
            "collection": "sys_roles",
            "matches": [
            { "key": "role", "value": "write2" }
            ]
        },
        "access": [
            {
            "collection": "my_collection",
            "access": "r"
            }
        ]
    }
}


```


## 🎯 Installation and Setup

In order to use this plugin you need to register it with Vault.
Configure your vault server to have a valid `plugins_directory` configuration. 

**Note: you might want to set `api_addr` to your listening address and `disable_mlock` to `true` in the `vault` configuration to be able to use the plugin.**

### Install from release

Download the latest stable release from the [release](https://github.com/migrx-io/vault-plugin-secrets-qdrant.git) page and put it into the `plugins_directory` of your vault server.

To use a vault plugin you need the plugin's sha256 sum. 

Example how to register the plugin:

```console
SHA256SUM=$(sha256sum vault-plugin-secrets-qdrant | cut -d' ' -f1)
vault plugin register -sha256 ${SHA256SUM} secret vault-plugin-secrets-qdrant
vault secrets enable -path=qdrant vault-plugin-secrets-qdrant
```

**Note: you might use the `-tls-skip-verify` flag if you are using a self-signed certificate.**


## Development

### Build locally

```console
$ make 
```

### Setup enviroment (docker compose)

```console
$ make setup-env
```

### Run unit tests

```console
$ make tests
```

### Run end-to-end tests

```console
$ make e2e
```

### Teardown enviroment (docker compose)

```console
$ make teardown-env
```

### Clean up

```console
$ make clean
```

# 🤝🏽 Contributing

Code contributions are very much **welcome**.

1. Fork the Project
2. Create your Branch (`git checkout -b AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature")
4. Push to the Branch (`git push origin AmazingFeature`)
5. Open a Pull Request targetting the `main` branch.
