#!/bin/bash

set -e

expect_equal() {
    # Usage: expect_equal op1 op2 message
    if [[ ! "$1" = "$2" ]]; then
        echo "$3: $1 != $2"
        exit 1
    fi
}

expect_not_equal() {
    # Usage: expect_equal op1 op2 message
    if [[ $1 = $2 ]]; then
        echo "$3: $1 = $2"
        exit 1
    fi
}

expect_match() {
    # Usage: expect_match str pattern message
    if [[ ! $1 =~ $2 ]]; then
        echo "$3: $1 does not match $2"
        exit 1
    fi
}

expect_no_match() {
    # Usage: expect_no_match str pattern message
    if [[ $1 =~ $2 ]]; then
        echo "$3: $1 matches $2"
        exit 1
    fi
}



export VAULT_ADDR='http://127.0.0.1:8200'

vault login root

echo -e "\n\n### Add instance config"
vault write qdrant/testdb/config "key_ttl=3s" "jwt_ttl=3s" 

echo -e "\n\n### Read instance config"
vault read qdrant/testdb/config


echo -e "\n\n### Attempt to create a token before role is created"
if vault write -field=token qdrant/testdb/sign/admin @claims.json; then
    echo "Signing with unknown role incorrectly succeeded."
    exit 1
fi

echo -e "\n\n### Adding role admin"
vault write qdrant/testdb/roles/admin claim="TEST"

echo -e "\n\n### Reading role admin"
vault read qdrant/testdb/roles/admin

echo -e "\n\n### Create a token with test role"
vault write -field=token qdrant/testdb/sign/admin @claims.json > jwt.txt
