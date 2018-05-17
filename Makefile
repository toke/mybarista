# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOINSTALL=$(GOCMD) install
DEPENSURE=dep ensure
BINARY_NAME=mybarista
BINARY_UNIX=$(BINARY_NAME)_unix

build:
	$(DEPENSURE)
	$(GOBUILD) -o $(BINARY_NAME) -v

install:
	$(GOINSTALL)

clean: 
		$(GOCLEAN)
		rm -f $(BINARY_NAME)
		rm -f $(BINARY_UNIX)

