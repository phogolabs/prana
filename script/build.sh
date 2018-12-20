#!/bin/bash

export GO111MODULE="on"

go mod download
go mod verify
go build github.com/phogolabs/prana/cmd/prana
