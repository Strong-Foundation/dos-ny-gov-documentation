#!/bin/bash

for ((i = 1; i <= 1000; i++)); do
    # Run Go program
    go run main.go
    # Git operations
    git add .
    git commit -m "updated $i"
    git push
done
