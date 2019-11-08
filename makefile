# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

EXAMPLES := examples/Pattern
SOURCES := $(wildcard $(EXAMPLES)/*.go)
EXECUTABLES := $(patsubst $(EXAMPLES)/%.go, %.exe, $(SOURCES))

all: build
build: deps $(EXECUTABLES)

%.exe: $(EXAMPLES)/%.go
	$(GOBUILD) $< 

clean: 
	$(GOCLEAN)
	$(GOCLEAN) dd
	rm -f *.exe 
	cd src/dd/device-detection-cxx && cmake .
	cd src/dd/device-detection-cxx && $(MAKE) clean

deps:
	cd src/dd/device-detection-cxx && cmake  -G "Unix Makefiles" .
	cd src/dd/device-detection-cxx && $(MAKE) fiftyone-pattern-c
	go install dd
