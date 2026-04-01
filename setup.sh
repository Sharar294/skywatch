#!/bin/bash

echo "  Scaffolding Go project structure in $(pwd)"

# 1. Create Directory Hierarchy
# Uses brace expansion for internal subdirectories
mkdir -p cmd/api \
         internal/{config,server,domain,database,service} \
         pkg \
         api \
         configs \
         deployments \
         scripts \
         test

# 2. Create Empty Placeholder Files
# Creating the foundation for your architecture
touch cmd/api/main.go
touch configs/.env
touch Makefile
touch .gitignore
touch README.md

echo "-------------------------------------------------------"
echo "📂 Structure created successfully."
echo "🛠️  Reminder: Run 'go mod init <module-name>' to begin."