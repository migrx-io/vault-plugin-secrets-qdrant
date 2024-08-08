# Test environment
FROM hashicorp/vault:1.15.2 

RUN apk add bash jq curl
COPY tests/testdata/* tests/e2e.sh /tests/

WORKDIR /tests
RUN chmod +x /tests/e2e.sh
RUN /tests/e2e.sh
