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
vault write qdrant/config "url=http://localhost:6333" "key_sig=secret" "sig_alg=RS256" "rsa_key_bits=4096" "key_ttl=3s" "jwt_ttl=3s" 

echo -e "\n\n### Read instance config"
vault read qdrant/config

echo -e "\n\n### Adding role admin"
vault write qdrant/roles/admin @basic.json

echo -e "\n\n### Reading role admin"
vault read qdrant/roles/admin

echo -e "\n\n### Create a token with test role"
vault write -field=token qdrant/sign/admin > basic_jwt.txt
