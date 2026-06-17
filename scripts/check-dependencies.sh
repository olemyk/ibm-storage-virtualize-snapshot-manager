#!/bin/bash
# Dependency Check Script
# Checks for outdated dependencies and security vulnerabilities

set -e

echo "=========================================="
echo "Dependency Check for Snapshot Manager"
echo "=========================================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Track if any issues found
ISSUES_FOUND=0

# Backend (Go) Dependencies
echo "📦 Checking Backend (Go) Dependencies..."
echo "------------------------------------------"

if [ -d "backend" ]; then
    cd backend
    
    echo "Checking for outdated modules..."
    OUTDATED=$(go list -u -m all 2>/dev/null | grep '\[' || true)
    
    if [ -z "$OUTDATED" ]; then
        echo -e "${GREEN}✓ All Go dependencies are up to date${NC}"
    else
        echo -e "${YELLOW}⚠ Outdated Go modules found:${NC}"
        echo "$OUTDATED"
        ISSUES_FOUND=1
    fi
    
    echo ""
    echo "Running security audit (govulncheck)..."
    
    # Check if govulncheck is installed
    if ! command -v govulncheck &> /dev/null; then
        echo "Installing govulncheck..."
        go install golang.org/x/vuln/cmd/govulncheck@latest
        # Add GOPATH/bin to PATH if not already there
        export PATH="$PATH:$(go env GOPATH)/bin"
    fi
    
    # Run govulncheck (exclude scripts directory which has multiple main functions)
    VULN_OUTPUT=$(govulncheck ./cmd/... ./internal/... 2>&1)
    if echo "$VULN_OUTPUT" | grep -q "No vulnerabilities found"; then
        echo -e "${GREEN}✓ No security vulnerabilities found${NC}"
    else
        echo -e "${YELLOW}⚠ Running vulnerability check...${NC}"
        echo "$VULN_OUTPUT"
        # Only mark as issue if actual vulnerabilities found (not just warnings)
        if echo "$VULN_OUTPUT" | grep -q "Vulnerability"; then
            ISSUES_FOUND=1
        fi
    fi
    
    cd ..
else
    echo -e "${RED}✗ Backend directory not found${NC}"
    ISSUES_FOUND=1
fi

echo ""
echo ""

# Frontend (npm) Dependencies
echo "📦 Checking Frontend (npm) Dependencies..."
echo "------------------------------------------"

if [ -d "frontend" ]; then
    cd frontend
    
    # Check if node_modules exists
    if [ ! -d "node_modules" ]; then
        echo "Installing npm dependencies..."
        npm ci
    fi
    
    echo "Checking for outdated packages..."
    OUTDATED_NPM=$(npm outdated 2>/dev/null || true)
    
    if [ -z "$OUTDATED_NPM" ]; then
        echo -e "${GREEN}✓ All npm packages are up to date${NC}"
    else
        echo -e "${YELLOW}⚠ Outdated npm packages found:${NC}"
        npm outdated
        ISSUES_FOUND=1
    fi
    
    echo ""
    echo "Running security audit..."
    
    if npm audit --audit-level=moderate 2>&1 | grep -q "found 0 vulnerabilities"; then
        echo -e "${GREEN}✓ No security vulnerabilities found${NC}"
    else
        echo -e "${RED}✗ Security vulnerabilities found!${NC}"
        npm audit
        ISSUES_FOUND=1
    fi
    
    cd ..
else
    echo -e "${RED}✗ Frontend directory not found${NC}"
    ISSUES_FOUND=1
fi

echo ""
echo ""
echo "=========================================="
echo "Summary"
echo "=========================================="

if [ $ISSUES_FOUND -eq 0 ]; then
    echo -e "${GREEN}✓ All dependencies are up to date and secure!${NC}"
    exit 0
else
    echo -e "${YELLOW}⚠ Issues found. Please review the output above.${NC}"
    echo ""
    echo "To update dependencies:"
    echo "  Backend: cd backend && make deps-update"
    echo "  Frontend: cd frontend && npm run deps:update"
    echo ""
    echo "To fix security issues:"
    echo "  Backend: Update vulnerable packages manually"
    echo "  Frontend: cd frontend && npm audit fix"
    exit 1
fi

# Made with Bob
