build: lint tidy vendor
	go build ./...
lint:
	golangci-lint run ./...
tidy:
	go mod tidy
vendor:
	go mod vendor
dirty:
	git diff --exit-code
ci : build dirty
.PHONY: build lint tidy vendor dirty ci