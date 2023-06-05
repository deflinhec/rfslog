# Define
VERSION=0.0.1
BUILD=$(shell git rev-parse HEAD)

# Setup linker flags option for build that interoperate with variable names in src code
LDFLAGS='-s -w -X "main.Version=$(VERSION)" -X "main.Build=$(BUILD)"'
ifeq ($(OS),Windows_NT)
	PLATFORM := windows
else
	PLATFORM := $(shell uname -s | tr A-Z a-z)
endif

CGO_ENABLED=1

.PHONY: default all build

default: fmt build tidy

fmt:
	go fmt ./...

tidy:
	go mod tidy

# Sperate "linux-amd64" as GOOS and GOARCH
OSARCH_SPERATOR = $(word $2,$(subst -, ,$1))

# Platform build options
cross-compile-%: export GOOS=$(call OSARCH_SPERATOR,$*,1)
cross-compile-%: export GOARCH=$(call OSARCH_SPERATOR,$*,2)
cross-compile-%: fmt tidy
	go build -trimpath -mod=vendor -ldflags $(LDFLAGS) -o ./build/$(GOOS)-$(GOARCH)/ .

# Arch build options
arch-%: export GOARCH=$(call OSARCH_SPERATOR,$*,1)
arch-%: fmt tidy
	go build -trimpath -mod=vendor -ldflags $(LDFLAGS) -o ./build/$(GOARCH)/ .

# Local build options
build: fmt tidy
	go build -trimpath -mod=vendor -ldflags $(LDFLAGS) -o ./build/$(PLATFORM)/ .