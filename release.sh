#!/bin/bash

go test -v
go build

export Commit=$(git rev-list -1 HEAD)
export Version=$(git describe --tags $(git rev-list --tags --max-count=1))

# brew install goreleaser/tap/goreleaser
# brew install goreleaser
# sudo snap install --classic goreleaser

goreleaser --rm-dist
