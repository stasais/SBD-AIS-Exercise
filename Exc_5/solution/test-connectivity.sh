#!/bin/bash
# Diagnostic script to test the Traefik + OrderService + SWS setup

echo "========================================="
echo "  Traefik & Microservices Diagnostics"
echo "========================================="
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test 1: Check if containers are running
echo "1. Checking if containers are running..."
if docker-compose ps | grep -q "Up"; then
    echo -e "${GREEN}✅ Containers are running${NC}"
    docker-compose ps --format="table {{.Name}}\t{{.Status}}"
else
    echo -e "${RED}❌ Containers are not running${NC}"
    echo "Run: docker-compose up -d"
    exit 1
fi
echo ""

# Test 2: Check DNS resolution
echo "2. Checking DNS resolution..."
if ping -c 1 orders.localhost &>/dev/null; then
    echo -e "${GREEN}✅ orders.localhost resolves to $(ping -c 1 orders.localhost | grep PING | awk '{print $3}')${NC}"
else
    echo -e "${RED}❌ orders.localhost does NOT resolve${NC}"
    echo "Fix: Add '127.0.0.1 orders.localhost' to /etc/hosts"
    echo "     sudo bash -c 'echo \"127.0.0.1 orders.localhost\" >> /etc/hosts'"
fi
echo ""

# Test 3: Check if Traefik is accessible
echo "3. Checking Traefik dashboard..."
if curl -s http://localhost:8080/api/overview >/dev/null; then
    echo -e "${GREEN}✅ Traefik dashboard is accessible at http://localhost:8080${NC}"
else
    echo -e "${RED}❌ Cannot reach Traefik dashboard${NC}"
fi
echo ""

# Test 4: Check OrderService directly
echo "4. Checking OrderService (direct connection)..."
MENU=$(docker-compose exec -T orderservice wget -qO- http://localhost:3000/api/menu 2>/dev/null)
if [ ! -z "$MENU" ]; then
    DRINK_COUNT=$(echo $MENU | python3 -c "import sys, json; print(len(json.load(sys.stdin)))" 2>/dev/null || echo "?")
    echo -e "${GREEN}✅ OrderService is responding ($DRINK_COUNT drinks in menu)${NC}"
else
    echo -e "${RED}❌ OrderService is not responding${NC}"
fi
echo ""

# Test 5: Check if Traefik can reach OrderService
echo "5. Checking Traefik → OrderService connectivity..."
TRAEFIK_TO_ORDER=$(docker-compose exec -T traefik wget -qO- --timeout=2 http://orderservice:3000/api/menu 2>/dev/null)
if [ ! -z "$TRAEFIK_TO_ORDER" ]; then
    echo -e "${GREEN}✅ Traefik can reach OrderService${NC}"
else
    echo -e "${RED}❌ Traefik cannot reach OrderService${NC}"
fi
echo ""

# Test 6: Check frontend (SWS)
echo "6. Checking Frontend (SWS)..."
FRONTEND=$(curl -s http://localhost/ | head -5)
if echo "$FRONTEND" | grep -q "<!DOCTYPE html>"; then
    echo -e "${GREEN}✅ Frontend is accessible at http://localhost${NC}"
else
    echo -e "${RED}❌ Frontend is not accessible${NC}"
fi
echo ""

# Test 7: Check Traefik routing to OrderService
echo "7. Checking Traefik routing (http://orders.localhost/api/menu)..."
echo -e "${YELLOW}⚠️  This test may timeout if DNS is not configured in Windows${NC}"
ROUTED_RESPONSE=$(timeout 5 curl -s http://orders.localhost/api/menu 2>&1)
if echo "$ROUTED_RESPONSE" | grep -q "Gateway Timeout"; then
    echo -e "${RED}❌ Gateway Timeout - Traefik routing issue or DNS problem${NC}"
    echo "   This is expected in WSL if Windows hosts file is not configured"
elif echo "$ROUTED_RESPONSE" | grep -q '"name"'; then
    echo -e "${GREEN}✅ Traefik routing works! API is accessible${NC}"
else
    echo -e "${YELLOW}⚠️  Could not verify routing (timeout or network issue)${NC}"
    echo "   Response: ${ROUTED_RESPONSE:0:100}"
fi
echo ""

# Test 8: Check Traefik router configuration
echo "8. Checking Traefik router configuration..."
ROUTERS=$(curl -s http://localhost:8080/api/http/routers 2>/dev/null | python3 -c "
import sys, json
routers = json.load(sys.stdin)
for r in routers:
    if 'orderservice' in r.get('name', '') or 'sws' in r.get('name', ''):
        print(f\"  - {r['name']}: {r['rule']} → {r.get('service', 'N/A')} [{r.get('status', 'N/A')}]\")
" 2>/dev/null)

if [ ! -z "$ROUTERS" ]; then
    echo -e "${GREEN}✅ Traefik routers configured:${NC}"
    echo "$ROUTERS"
else
    echo -e "${RED}❌ Could not retrieve router configuration${NC}"
fi
echo ""

# Summary
echo "========================================="
echo "  Summary & Next Steps"
echo "========================================="
echo ""
echo "For WSL users accessing from Windows browser:"
echo "  1. WSL: Add to /etc/hosts (for curl/wget in WSL)"
echo "     ${YELLOW}sudo bash -c 'echo \"127.0.0.1 orders.localhost\" >> /etc/hosts'${NC}"
echo ""
echo "  2. WINDOWS: Add to C:\\Windows\\System32\\drivers\\etc\\hosts"
echo "     ${YELLOW}127.0.0.1 orders.localhost${NC}"
echo "     (Open Notepad as Administrator, edit hosts file, save)"
echo ""
echo "  3. Flush DNS cache in Windows:"
echo "     ${YELLOW}ipconfig /flushdns${NC}"
echo ""
echo "  4. Refresh browser (Ctrl+Shift+R)"
echo ""
echo "URLs to test:"
echo "  - Frontend:  http://localhost"
echo "  - API:       http://orders.localhost/api/menu"
echo "  - Dashboard: http://localhost:8080"
echo ""
