.PHONY: build install clean test coverage

BINARY_NAME=helm
INSTALL_DIR=$(HOME)/.local/bin

build:
	go build -o $(BINARY_NAME) ./cmd/helm/

install: build
	mkdir -p $(INSTALL_DIR)
	rm -f $(INSTALL_DIR)/$(BINARY_NAME)
	cp $(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)
ifeq ($(shell uname),Darwin)
	xattr -c $(INSTALL_DIR)/$(BINARY_NAME)
endif
	@echo "Installed $(BINARY_NAME) to $(INSTALL_DIR)"

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

fmt:
	go fmt ./...

lint:
	golangci-lint run

tidy:
	go mod tidy
