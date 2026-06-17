#!/bin/bash
set -e

echo "=========================================="
echo "IBM Storage Virtualize Snapshot Manager"
echo "Container Deployment Setup"
echo "=========================================="
echo ""

# Check for required tools
echo "Checking prerequisites..."

if ! command -v podman &> /dev/null; then
    echo "❌ Error: Podman is not installed"
    echo "   Install from: https://podman.io/getting-started/installation"
    exit 1
fi
echo "✓ Podman found: $(podman --version)"

if ! command -v podman-compose &> /dev/null; then
    echo "❌ Error: Podman Compose is not installed"
    echo "   Install with: pip3 install podman-compose"
    exit 1
fi
echo "✓ Podman Compose found: $(podman-compose --version)"

if ! command -v openssl &> /dev/null; then
    echo "❌ Error: OpenSSL is not installed"
    exit 1
fi
echo "✓ OpenSSL found"

echo ""
echo "Prerequisites check passed!"
echo ""

# Create .env file if it doesn't exist
if [ -f .env ]; then
    echo "⚠️  .env file already exists"
    read -p "Do you want to regenerate security keys? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Keeping existing .env file"
    else
        echo "Regenerating security keys..."
        
        # Generate new keys
        JWT_SECRET=$(openssl rand -base64 32)
        ENCRYPTION_KEY=$(openssl rand -base64 32)
        
        # Update existing .env file
        if grep -q "^JWT_SECRET=" .env; then
            sed -i.bak "s|^JWT_SECRET=.*|JWT_SECRET=$JWT_SECRET|" .env
        else
            echo "JWT_SECRET=$JWT_SECRET" >> .env
        fi
        
        if grep -q "^ENCRYPTION_KEY=" .env; then
            sed -i.bak "s|^ENCRYPTION_KEY=.*|ENCRYPTION_KEY=$ENCRYPTION_KEY|" .env
        else
            echo "ENCRYPTION_KEY=$ENCRYPTION_KEY" >> .env
        fi
        
        rm -f .env.bak
        echo "✓ Security keys regenerated in .env file"
    fi
else
    echo "Creating .env file from template..."
    cp .env.example .env
    
    # Generate JWT secret (32+ characters)
    JWT_SECRET=$(openssl rand -base64 32)
    sed -i.bak "s|^JWT_SECRET=.*|JWT_SECRET=$JWT_SECRET|" .env
    
    # Generate encryption key (exactly 32 bytes when decoded)
    ENCRYPTION_KEY=$(openssl rand -base64 32)
    sed -i.bak "s|^ENCRYPTION_KEY=.*|ENCRYPTION_KEY=$ENCRYPTION_KEY|" .env
    
    # Generate random database password
    DB_PASSWORD=$(openssl rand -base64 24)
    sed -i.bak "s|^DB_PASSWORD=.*|DB_PASSWORD=$DB_PASSWORD|" .env
    
    rm -f .env.bak
    
    echo "✓ Created .env file with generated security keys"
    echo ""
    echo "⚠️  IMPORTANT: Review and update .env file if needed"
    echo "   - Database password has been auto-generated"
    echo "   - JWT_SECRET and ENCRYPTION_KEY have been generated"
fi

echo ""

# Create SSL directory and generate self-signed certificate
if [ ! -d ssl ]; then
    mkdir -p ssl
    echo "Generating self-signed SSL certificate..."
    
    openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
        -keyout ssl/nginx-selfsigned.key \
        -out ssl/nginx-selfsigned.crt \
        -subj "/C=US/ST=State/L=City/O=IBM Storage Snapshot Manager/CN=localhost" \
        2>/dev/null
    
    echo "✓ Self-signed SSL certificate generated in ./ssl/"
    echo ""
    echo "⚠️  For production use, replace with proper SSL certificates:"
    echo "   - Place certificate in: ssl/nginx-selfsigned.crt"
    echo "   - Place private key in: ssl/nginx-selfsigned.key"
else
    echo "✓ SSL directory already exists"
    if [ -f ssl/nginx-selfsigned.crt ] && [ -f ssl/nginx-selfsigned.key ]; then
        echo "✓ SSL certificates found"
    else
        echo "⚠️  SSL certificates not found, generating..."
        openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
            -keyout ssl/nginx-selfsigned.key \
            -out ssl/nginx-selfsigned.crt \
            -subj "/C=US/ST=State/L=City/O=IBM Storage Snapshot Manager/CN=localhost" \
            2>/dev/null
        echo "✓ Self-signed SSL certificate generated"
    fi
fi

echo ""
echo "=========================================="
echo "Setup Complete!"
echo "=========================================="
echo ""
echo "Next steps:"
echo "1. Review .env file and update if needed"
echo "2. Start the application: ./deploy/start.sh"
echo ""
echo "After starting, access the application at:"
echo "  https://localhost"
echo ""
echo "Default credentials:"
echo "  Username: admin"
echo "  Password: admin123"
echo ""
echo "⚠️  Change the default password after first login!"
echo ""
