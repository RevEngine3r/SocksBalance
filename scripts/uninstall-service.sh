#!/bin/bash

# SocksBalance Service Uninstallation Script
# This script removes SocksBalance systemd service from Linux

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${RED}SocksBalance Service Uninstaller${NC}"
echo "=================================="
echo ""

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}Error: This script must be run as root${NC}"
    echo "Please run: sudo $0"
    exit 1
fi

# Variables
USER="socksbalance"
INSTALL_DIR="/opt/socksbalance"
CONFIG_DIR="/etc/socksbalance"
LOG_DIR="/var/log/socksbalance"
SERVICE_FILE="socksbalance.service"

echo -e "${YELLOW}This will remove:${NC}"
echo "  - Systemd service"
echo "  - Binary from $INSTALL_DIR"
echo "  - Configuration from $CONFIG_DIR (optional)"
echo "  - Logs from $LOG_DIR (optional)"
echo "  - System user '$USER' (optional)"
echo ""
read -p "Continue? (y/N): " -n 1 -r
echo ""

if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Aborted."
    exit 0
fi

echo ""
echo -e "${YELLOW}Step 1: Stopping and disabling service${NC}"
if systemctl is-active --quiet socksbalance; then
    systemctl stop socksbalance
    echo -e "${GREEN}✓ Service stopped${NC}"
else
    echo -e "${GREEN}✓ Service already stopped${NC}"
fi

if systemctl is-enabled --quiet socksbalance 2>/dev/null; then
    systemctl disable socksbalance
    echo -e "${GREEN}✓ Service disabled${NC}"
else
    echo -e "${GREEN}✓ Service already disabled${NC}"
fi

echo ""
echo -e "${YELLOW}Step 2: Removing service file${NC}"
if [ -f "/etc/systemd/system/$SERVICE_FILE" ]; then
    rm "/etc/systemd/system/$SERVICE_FILE"
    systemctl daemon-reload
    echo -e "${GREEN}✓ Service file removed${NC}"
else
    echo -e "${GREEN}✓ Service file not found${NC}"
fi

echo ""
echo -e "${YELLOW}Step 3: Removing binary${NC}"
if [ -d "$INSTALL_DIR" ]; then
    rm -rf "$INSTALL_DIR"
    echo -e "${GREEN}✓ Binary directory removed${NC}"
else
    echo -e "${GREEN}✓ Binary directory not found${NC}"
fi

echo ""
read -p "Remove configuration files from $CONFIG_DIR? (y/N): " -n 1 -r
echo ""
if [[ $REPLY =~ ^[Yy]$ ]]; then
    if [ -d "$CONFIG_DIR" ]; then
        rm -rf "$CONFIG_DIR"
        echo -e "${GREEN}✓ Configuration removed${NC}"
    else
        echo -e "${GREEN}✓ Configuration not found${NC}"
    fi
else
    echo -e "${YELLOW}→ Configuration kept at $CONFIG_DIR${NC}"
fi

echo ""
read -p "Remove log files from $LOG_DIR? (y/N): " -n 1 -r
echo ""
if [[ $REPLY =~ ^[Yy]$ ]]; then
    if [ -d "$LOG_DIR" ]; then
        rm -rf "$LOG_DIR"
        echo -e "${GREEN}✓ Logs removed${NC}"
    else
        echo -e "${GREEN}✓ Logs not found${NC}"
    fi
else
    echo -e "${YELLOW}→ Logs kept at $LOG_DIR${NC}"
fi

echo ""
read -p "Remove system user '$USER'? (y/N): " -n 1 -r
echo ""
if [[ $REPLY =~ ^[Yy]$ ]]; then
    if id "$USER" &>/dev/null; then
        userdel "$USER" 2>/dev/null || true
        echo -e "${GREEN}✓ User removed${NC}"
    else
        echo -e "${GREEN}✓ User not found${NC}"
    fi
else
    echo -e "${YELLOW}→ User '$USER' kept${NC}"
fi

echo ""
echo -e "${GREEN}================================================${NC}"
echo -e "${GREEN}Uninstallation complete!${NC}"
echo -e "${GREEN}================================================${NC}"
echo ""
