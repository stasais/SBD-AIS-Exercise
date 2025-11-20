#!/bin/sh
# Exit if any command fails
set -e
cd /app
go mod tidy
go mod download
CGO_ENABLED=0 GOOS=linux go build -o /app/ordersystem