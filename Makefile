build: lint tidy dirty
	go build ./...
lint:
	golangci-lint run ./...
tidy:
	go mod tidy
dirty:
	git diff --exit-code
ci : build dirty
.PHONY: build lint tidy vendor dirty ci
