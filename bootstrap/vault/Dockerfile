# Build the addon
FROM golang:1.22.5-alpine as plugin-builder
WORKDIR /app

ADD ../. /app
RUN go mod download
RUN go env -w GO111MODULE=on

RUN go build -o /vault/plugins/vault-plugin-secrets-qdrant cmd/vault-plugin-secrets-qdrant/main.go

# Package Vault
FROM hashicorp/vault:1.15.2 
COPY --from=plugin-builder /vault/plugins /vault/plugins/

COPY ./bootstrap/vault/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
