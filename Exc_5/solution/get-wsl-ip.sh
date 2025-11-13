#!/bin/bash
# Helper script to generate Windows hosts file entries for WSL2

echo "========================================="
echo "  WSL2 Windows Hosts File Generator"
echo "========================================="
echo ""

# Get WSL IP address
WSL_IP=$(hostname -I | awk '{print $1}')

echo "✅ Your current WSL IP address: $WSL_IP"
echo ""
echo "📋 Copy these lines to your Windows hosts file:"
echo "   Location: C:\\Windows\\System32\\drivers\\etc\\hosts"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "$WSL_IP orders.localhost"
echo "$WSL_IP localhost"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "📝 Steps:"
echo "  1. Open Notepad as Administrator"
echo "  2. Open C:\\Windows\\System32\\drivers\\etc\\hosts"
echo "  3. Remove any old lines with 'orders.localhost' or '127.0.0.1 localhost'"
echo "  4. Add the lines above"
echo "  5. Save the file"
echo "  6. In Windows CMD/PowerShell (as Admin): ipconfig /flushdns"
echo ""
echo "🧪 Test in Windows:"
echo "  ping orders.localhost"
echo "  (Should respond from $WSL_IP)"
echo ""
echo "⚠️  Remember: WSL IP can change after restart!"
echo "   Run this script again if services become unreachable."
echo ""
