.PHONY: build clean test web go dev build-all

VERSION ?= dev
LDFLAGS := -ldflags "-X main.version=$(VERSION)"
STATIC_DIR := internal/server/static
DIST_DIR := dist

build: web go

web:
	cd web && npm install && npx tsc --noEmit && npx vite build
	rm -rf $(STATIC_DIR)
	cp -r web/dist $(STATIC_DIR)

go:
	go build $(LDFLAGS) -o tgviz ./cmd/tgviz

build-all: web
	mkdir -p $(DIST_DIR)
	GOOS=darwin  GOARCH=arm64 go build $(LDFLAGS) -o $(DIST_DIR)/tgviz-darwin-arm64  ./cmd/tgviz
	GOOS=darwin  GOARCH=amd64 go build $(LDFLAGS) -o $(DIST_DIR)/tgviz-darwin-amd64  ./cmd/tgviz
	GOOS=linux   GOARCH=amd64 go build $(LDFLAGS) -o $(DIST_DIR)/tgviz-linux-amd64   ./cmd/tgviz
	GOOS=linux   GOARCH=arm64 go build $(LDFLAGS) -o $(DIST_DIR)/tgviz-linux-arm64   ./cmd/tgviz
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(DIST_DIR)/tgviz-windows-amd64.exe ./cmd/tgviz

test:
	go test ./... -v -count=1

clean:
	rm -rf tgviz $(DIST_DIR) web/dist web/node_modules $(STATIC_DIR)

dev:
	cd web && npx vite
