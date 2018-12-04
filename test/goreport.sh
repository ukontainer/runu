#!/bin/sh

set -e
set -x

gofmt -s -d .
gometalinter --deadline=180s --disable-all --enable=gofmt
gometalinter --deadline=180s --disable-all --enable=vet

# gocyclo
gometalinter --deadline=180s --disable-all --enable=gocyclo --cyclo-over=15
# golint
gometalinter --deadline=180s --disable-all --enable=golint --min-confidence=0.85 --vendor

# ineffassign
gometalinter --deadline=180s --disable-all --enable=ineffassign

# misspell
gometalinter --deadline=180s --disable-all --enable=misspell
