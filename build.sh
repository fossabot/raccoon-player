#!/usr/bin/env bash

scriptDir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
go build -o /usr/local/bin/raccoon-player ${scriptDir}/main.go
