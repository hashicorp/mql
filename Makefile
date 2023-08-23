# Determine this makefile's path.
# Be sure to place this BEFORE `include` directives, if any.
THIS_FILE := $(lastword $(MAKEFILE_LIST))
THIS_DIR := $(dir $(realpath $(firstword $(MAKEFILE_LIST))))

TMP_DIR := $(shell mktemp -d)
REPO_PATH := github.com/hashicorp/mql

.PHONY: fmt
fmt:
	gofumpt -w $$(find . -name '*.go')

.PHONY: test
test: 
	go test -race -count=1 ./...

.PHONY: test-all
test-all: test test-postgres

.PHONY: test-postgres
test-postgres:
	##############################################################
	# this test is dependent on first running: docker-compose up
	##############################################################
	cd ./tests/postgres && \
	DB_DIALECT=postgres DB_DSN="postgresql://go_db:go_db@localhost:9920/go_db?sslmode=disable"  go test -race -count=1 ./...

.PHONY: coverage
coverage: 
	cd coverage && \
	./coverage.sh
