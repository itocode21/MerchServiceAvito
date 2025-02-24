#!/bin/bash
echo "Building Avito Merch Service..."
go build -o bin/merch-service cmd/server/main.go
if [ $? -eq 0 ]; then
    echo "Build successful. Binary located at bin/merch-service"
else
    echo "Build failed."
    exit 1
fi