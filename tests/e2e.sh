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

# echo -e "\n\n### Read all instance config"
# vault read qdrant/config/

echo -e "\n\n### Add instance config"
vault write qdrant/config/instance1 "url=http://localhost:6333" "sig_key=your-very-long-256-bit-secret-key" "sig_alg=HS256" "jwt_ttl=3s" 

echo -e "\n\n### Read instance config"
vault read qdrant/config/instance1

echo -e "\n\n### Read all instance config"
vault list qdrant/config

echo -e "\n\n### Delete instance config"
vault delete qdrant/config/instance1

echo -e "\n\n### Add instance config"
vault write qdrant/config/instance1 "url=http://localhost:6333" "sig_key=your-very-long-256-bit-secret-key" "sig_alg=HS256" "jwt_ttl=3s" 

echo -e "\n\n### Adding role write"
vault write qdrant/role/instance1/write @basic.json

echo -e "\n\n### Reading role write"
vault read qdrant/role/instance1/write

echo -e "\n\n### read token write"
vault read qdrant/jwt/instance1/write

