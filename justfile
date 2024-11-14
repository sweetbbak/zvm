default:
    just build
build:
    go build -ldflags="-s -w" ./cmd/zvm
