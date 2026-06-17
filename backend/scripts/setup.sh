#!/bin/bash

# IBM Storage Virtualize Snapshot Manager - Setup Script

set -e

echo "==================================="
echo "IBM Storage Virtualize Snapshot Manager"
echo "Initial Setup"
echo "==================================="
echo ""

# Create directories
echo "Creating directories..."
mkdir -p data logs
echo "✓ Directories created"
echo ""

# Generate encryption keys
echo "Generating secure keys..."
go run -c 'package main; import ("crypto/rand"; "encoding/base64"; "fmt"; "os"); func main() { key := make([]byte, 32); rand.Read(key); fmt.Fprintf(os.Stderr, "ENCRYPTION_KEY=%s\n", base64.StdEncoding.EncodeToString(key)); secret := make([]byte, 32); rand.Read(secret); fmt.Fprintf(os.Stderr, "JWT_SECRET=%s\n", base64.StdEncoding.EncodeToString(secret)) }' 2>&1 | tee .env.generated

echo ""
echo "✓ Keys generated and saved to .env.generated"
echo ""

# Check if .env exists
if [ -f .env ]; then
    echo "⚠️  .env file already exists. Please manually add the keys from .env.generated"
else
    echo "Creating .env file..."
    cp .env.example .env
    
    # Append generated keys
    cat .env.generated >> .env
    
    echo "✓ .env file created with generated keys"
fi

echo ""
echo "==================================="
echo "Setup Complete!"
echo "==================================="
echo ""
echo "Next steps:"
echo "1. Review and edit .env file if needed"
echo "2. Run: go run cmd/server/main.go"
echo "3. Create initial user (see README.md)"
echo ""

# 
