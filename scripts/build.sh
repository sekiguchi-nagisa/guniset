#!/bin/sh

SCRIPT_DIR=$(cd $(dirname $0); pwd)

cd "$SCRIPT_DIR/../"  # move to project top

(cd ./op && go generate && go mod tidy)

GOTOOLCHAIN=auto go build