#!/bin/sh

PLUGIN_NAME=vault-plugin-secrets-qdrant

# Configure vault
vault server -dev -log-level=debug -dev-root-token-id="root" -dev-listen-address=0.0.0.0:8200 -config=/vault/config.hcl &
VAULT_PROC=$!

sleep 3

export VAULT_ADDR='http://127.0.0.1:8200'

SHASUM=$(sha256sum "/vault/plugins/$PLUGIN_NAME" | cut -d " " -f1)

vault login root

set -e

echo -e "\n\n### Register plugin"
vault plugin register -sha256 $SHASUM $PLUGIN_NAME

echo -e "\n\n### Enable JWT engine at /qdrant path"
vault secrets enable -path=qdrant $PLUGIN_NAME

# sleep infinity
tail -f /dev/null
