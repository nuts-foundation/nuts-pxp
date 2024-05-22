.PHONY: run-generators test api

run-generators: api

install-tools:
	go install github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen@v2.1.0
	go install go.uber.org/mock/mockgen@v0.4.0
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.57.2

api:
	oapi-codegen --config oas/pip.config.yaml oas/pip.yaml | gofmt > api/pip/generated.go
	oapi-codegen --config oas/opa.config.yaml oas/opa.yaml | gofmt > api/opa/generated.go

lint:
	golangci-lint run -v

test:
	go test ./...

docker:
	docker build -t nutsfoundation/nuts-pxp:main .
