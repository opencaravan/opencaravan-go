# List available recipes
default:
    @just --list

[group('test')]
fmt:
    gofmt -w .

[group('test')]
fmt-check:
    @test -z "$(gofmt -l .)" || (echo "Files need formatting:" && gofmt -l . && exit 1)

[group('test')]
test:
    go test ./...

[group('test')]
ci: fmt-check test
