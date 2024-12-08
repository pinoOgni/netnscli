include Makefile.defs

GO := go
GO_VERSION := $(shell cat GO_VERSION)
GO_NATIVE_OS := $(shell $(GO) env GOOS)
GO_NATIVE_ARCH := $(shell $(GO) env GOARCH)
GOOS ?= $(GO_NATIVE_OS)
GOARCH ?= $(GO_NATIVE_ARCH)
ifeq ($(GOOS),)
	GOOS := linux
endif
ifeq ($(GOACH),)
	GOARCH := amd64
endif

GO_BUILD_FLAGS =
GO_BUILD_TAGS =
GO_BUILD_FLAGS += -tags="$(call JOIN_WITH_COMMA,$(GO_BUILD_TAGS))"

### Base directories ###
CMD_DIR := cmd
BIN_DIR := bin
OUTPUT_DIR = .output
RELEASE_DIR := release
OUTPUT_DIRS := $(BIN_DIR) $(TOOLS_DIR) $(OUTPUT_DIR) $(RELEASE_DIR)
SCRIPTS_DIR := scripts
BIN := netnscli


.PHONY: help all clean
all: help

clean-all: clean-build

clean-build:
	rm ${BIN_DIR}

build:
	$(call msg,BUILD,$@)
	$(Q) GO=$(GO) GOOS=$(GOOS) GOARCH=$(GOARCH) GO_BUILD_FLAGS='$(GO_BUILD_FLAGS)' BIN=$(BIN) \
		 $(SCRIPTS_DIR)/bin_build_netnscli.sh

run: build
	./${BIN_DIR}/${BIN}

help:
	@echo "Make Targets:"
	@echo " build		- build the Netnscli"
	@echo " run			- build and run the Netnscli"
	@echo " clean-build	- clean the Netnscli build"
	@echo " clean-all	- clean the Netnscli build and release"
	@echo ""
# TODO release