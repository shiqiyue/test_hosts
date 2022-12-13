#!/bin/bash
GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o testHosts .
