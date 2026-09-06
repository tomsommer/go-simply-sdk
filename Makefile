.PHONY: fmt vet test vulncheck check build clean verify-generated update-spec
gofmt:
	gofmt -w .
fmt: gofmt
vet:
	go vet ./...
test:
	go test ./...
vulncheck:
	@if command -v govulncheck >/dev/null 2>&1; then govulncheck ./...; else echo 'govulncheck not installed; skipped'; fi
build:
	go build ./...
check: fmt vet test build
verify-generated:
	@test ! -f internal/generated/client.gen.go || test -f openapi/doc.json
update-spec:
	@echo 'Review and pin https://api.simply.com/2/doc.json manually; no silent overwrite'
clean:
	go clean -testcache
