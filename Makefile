.PHONY: build install install-hooks clean test coverage dev themes

BINARY_NAME=helm
INSTALL_DIR=$(HOME)/.local/bin

build:
	go build -o $(BINARY_NAME) ./cmd/helm/

install: build install-hooks
	mkdir -p $(INSTALL_DIR)
	ln -sf $(CURDIR)/$(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)
	ln -sf $(CURDIR)/hooks/helm-tmux-hook.sh $(INSTALL_DIR)/helm-tmux-hook
	@echo "Installed $(BINARY_NAME) and helm-tmux-hook to $(INSTALL_DIR) (symlinks)"

install-hooks:
	git config core.hooksPath .githooks
	@echo "Git hooks installed (.githooks)"

clean:
	rm -f $(BINARY_NAME)
	go clean

test:
	go test ./...

coverage:
	go test ./... -cover

# Development helpers
run: build
	./$(BINARY_NAME)

dev:
	find cmd internal -name '*.go' | entr -r make build

fmt:
	go fmt ./...
	# Format all markdown files
	prettier --write '**/*.md'

lint:
	golangci-lint run

tidy:
	go mod tidy

# Regenerate Black Atom theme files from templates (requires deno)
themes:
	deno task generate
	gofmt -w internal/ui/theme
	go build ./...
