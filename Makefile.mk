SHELL := /bin/bash
BUILDPATH=$(CURDIR)
UTILS_PATH=/workdir
UTILS_BIN_PATH=$(CURDIR)
PROG_NAME=before_reindex

# docker parameters
DOCKERCMD=$(shell which docker)

GOBUILDPATHINCONTAINER=/workdir
GOBUILDIMAGE=golang:1.18.4


compile:
	@echo "compiling binary for $(PROG_NAME)..."
	@echo $(GOBUILDPATHINCONTAINER)
	@$(DOCKERCMD) run --rm -v $(BUILDPATH):$(GOBUILDPATHINCONTAINER) -w $(UTILS_PATH) $(GOBUILDIMAGE) go build  -o $(PROG_NAME)
	@echo "Done."

container:
	@echo "build container"
	@$(DOCKERCMD) build -f Dockerfile.dockerfile -t firstfloor/$(PROG_NAME):latest .
	@echo "Done."

all: compile container	