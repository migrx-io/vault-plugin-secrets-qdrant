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


expect_contains() {
    # Usage: expect_contains string word message
    if [[ "$1" != *"$2"* ]]; then
        echo "$3: '$2' not found in '$1'"
        exit 1
    fi
}


expect_not_contains() {
    # Usage: expect_contains string word message
    if [[ "$1" == *"$2"* ]]; then
        echo "$3: '$2' found in '$1'"
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
vault write qdrant/config/instance1 "url=qdrant:6334" "sig_key=your-very-long-256-bit-secret-key" "sig_alg=HS256" "jwt_ttl=120s" "tls=false"

echo -e "\n\n### Read instance config"
vault read qdrant/config/instance1

echo -e "\n\n### Read all instance config"
vault list qdrant/config

echo -e "\n\n### Delete instance config"
vault delete qdrant/config/instance1

echo -e "\n\n### Add instance config"
vault write qdrant/config/instance1 "url=qdrant:6334" "sig_key=your-very-long-256-bit-secret-key" "sig_alg=HS256" "jwt_ttl=10s" 

echo -e "\n\n### Adding role write"
vault write qdrant/role/instance1/write @basic.json

echo -e "\n\n### Adding role write2"
vault write qdrant/role/instance1/write2 @value_exists.json

echo -e "\n\n### Adding role admin"
vault write qdrant/role/instance1/admin @manage.json

echo -e "\n\n### Reading role write"
vault read qdrant/role/instance1/write

echo -e "\n\n### Reading role admin"
vault read qdrant/role/instance1/admin

echo -e "\n\n### read token write"
vault read qdrant/jwt/instance1/write

echo -e "\n\n### Check global token"
API_KEY=$(vault read qdrant/jwt/instance1/write|grep 'token'|awk '{print $2}')
out=$(curl -XGET -H 'Api-Key: '$API_KEY http://localhost:6333/cluster)
echo $out

expect_contains $out "result"


echo -e "\n\n### Check colletion token doesn't have access"
API_KEY=$(vault read qdrant/jwt/instance1/write2|grep 'token'|awk '{print $2}')
out=$(curl -XGET -H 'Api-Key: '$API_KEY http://localhost:6333/cluster)
echo $out

expect_not_contains $out "result"


echo -e "\n\n### Check colletion token have access"
out=$(curl -XGET -H 'Api-Key: '$API_KEY http://localhost:6333/collections/sys_roles)
echo $out

expect_contains $out "result"


echo -e "\n\n### Delete role and check access again"

vault delete qdrant/role/instance1/write2


out=$(curl -XGET -H 'Api-Key: '$API_KEY http://localhost:6333/collections/sys_roles)
echo $out

expect_not_contains $out "result"


# echo -e "\n\n### Check token expire"
# API_KEY=$(vault read qdrant/jwt/instance1/write|grep 'token'|awk '{print $2}')
# echo $API_KEY
# out=$(curl -XGET -H 'Api-Key: '$API_KEY http://localhost:6333/cluster)
# echo $out
# expect_contains $out "result"

# sleep 3

# out=$(curl -XGET -H 'Api-Key: '$API_KEY http://localhost:6333/cluster)
# echo $out

# expect_not_contains $out "result"
