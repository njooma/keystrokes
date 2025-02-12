BIN_OUTPUT_PATH = bin
TOOL_BIN = bin/gotools/$(shell uname -s)-$(shell uname -m)
UNAME_S ?= $(shell uname -s)
GOPATH = $(HOME)/go/bin
export PATH := ${PATH}:$(GOPATH)

build: format update-rdk
	rm -f $(BIN_OUTPUT_PATH)/keystrokes
	go build $(LDFLAGS) -o $(BIN_OUTPUT_PATH)/keystrokes main.go

module.tar.gz: 
	GOPRIVATE=github.com/viam-labs/screenshot-cam GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" .
	rm -f $@
	tar czf $@ meta.json keystrokes.exe

setup:
	if [ "$(UNAME_S)" = "Linux" ]; then \
		sudo apt-get install -y apt-utils coreutils tar libnlopt-dev libjpeg-dev pkg-config; \
	fi
	# remove unused imports
	go install golang.org/x/tools/cmd/goimports@latest
	find . -name '*.go' -exec $(GOPATH)/goimports -w {} +


clean:
	rm -rf $(BIN_OUTPUT_PATH)/keystrokes $(BIN_OUTPUT_PATH)/module.tar.gz keystrokes

format:
	gofmt -w -s .

update-rdk:
	go get go.viam.com/rdk@latest
	go mod tidy
