.PHONY: build clean test web go dev

VERSION ?= dev
LDFLAGS := -ldflags "-X main.version=$(VERSION)"
STATIC_DIR := internal/server/static

build: web go

web:
	cd web && npm install && npx tsc --noEmit && npx vite build
	rm -rf $(STATIC_DIR)
	cp -r web/dist $(STATIC_DIR)

go:
	go build $(LDFLAGS) -o tgviz ./cmd/tgviz

test:
	go test ./... -v -count=1

clean:
	rm -rf tgviz web/dist web/node_modules $(STATIC_DIR)

dev:
	cd web && npx vite
