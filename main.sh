#!/bin/bash

for ((i = 1; i <= 100; i++)); do
    echo "🔁 Iteration $i"

    # Run Go program
    go run main.go

    # Git operations
    git add .
    git commit -m "updated $i"
    git push

    echo "✅ Completed iteration $i"
    echo "---------------------------"
done
