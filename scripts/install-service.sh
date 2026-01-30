#!/bin/bash

# SocksBalance Service Installation Script
# This script installs SocksBalance as a systemd service on Linux

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}SocksBalance Service Installer${NC}"
echo "================================"
echo ""

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}Error: This script must be run as root${NC}"
    echo "Please run: sudo $0"
    exit 1
fi

# Check if systemd is available
if ! command -v systemctl &> /dev/null; then
    echo -e "${RED}Error: systemd not found${NC}"
    echo "This script requires systemd for service management"
    exit 1
fi

# Variables (customize these)
USER="socksbalance"
GROUP="socksbalance"
INSTALL_DIR="/opt/socksbalance"
CONFIG_DIR="/etc/socksbalance"
LOG_DIR="/var/log/socksbalance"
BINARY_NAME="socksbalance"
SERVICE_FILE="socksbalance.service"

# Check if binary exists in current directory
if [ ! -f "./$BINARY_NAME" ]; then
    echo -e "${RED}Error: Binary '$BINARY_NAME' not found in current directory${NC}"
    echo "Please build the binary first:"
    echo "  go build -o $BINARY_NAME ./cmd/socksbalance"
    exit 1
fi

echo -e "${YELLOW}Step 1: Creating user and group${NC}"
if ! id "$USER" &>/dev/null; then
    useradd --system --no-create-home --shell /bin/false "$USER"
    echo -e "${GREEN}✓ User '$USER' created${NC}"
else
    echo -e "${GREEN}✓ User '$USER' already exists${NC}"
fi

echo ""
echo -e "${YELLOW}Step 2: Creating directories${NC}"
mkdir -p "$INSTALL_DIR"
mkdir -p "$CONFIG_DIR"
mkdir -p "$LOG_DIR"
chown -R "$USER:$GROUP" "$INSTALL_DIR"
chown -R "$USER:$GROUP" "$LOG_DIR"
echo -e "${GREEN}✓ Directories created${NC}"

echo ""
echo -e "${YELLOW}Step 3: Installing binary${NC}"
cp "./$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
chmod +x "$INSTALL_DIR/$BINARY_NAME"
chown "$USER:$GROUP" "$INSTALL_DIR/$BINARY_NAME"
echo -e "${GREEN}✓ Binary installed to $INSTALL_DIR/$BINARY_NAME${NC}"

echo ""
echo -e "${YELLOW}Step 4: Installing configuration${NC}"
if [ -f "./config.yaml" ]; then
    cp "./config.yaml" "$CONFIG_DIR/config.yaml"
    echo -e "${GREEN}✓ Copied existing config.yaml${NC}"
elif [ -f "./config.example.yaml" ]; then
    cp "./config.example.yaml" "$CONFIG_DIR/config.yaml"
    echo -e "${YELLOW}⚠ Copied config.example.yaml as config.yaml${NC}"
    echo -e "${YELLOW}  Please edit $CONFIG_DIR/config.yaml before starting the service${NC}"
else
    echo -e "${RED}✗ No config file found${NC}"
    echo -e "${YELLOW}  Please create $CONFIG_DIR/config.yaml manually${NC}"
fi
chown "$USER:$GROUP" "$CONFIG_DIR/config.yaml" 2>/dev/null || true
chmod 640 "$CONFIG_DIR/config.yaml" 2>/dev/null || true

echo ""
echo -e "${YELLOW}Step 5: Installing systemd service${NC}"
if [ -f "./scripts/$SERVICE_FILE" ]; then
    cp "./scripts/$SERVICE_FILE" "/etc/systemd/system/$SERVICE_FILE"
    echo -e "${GREEN}✓ Service file installed${NC}"
elif [ -f "./$SERVICE_FILE" ]; then
    cp "./$SERVICE_FILE" "/etc/systemd/system/$SERVICE_FILE"
    echo -e "${GREEN}✓ Service file installed${NC}"
else
    echo -e "${RED}✗ Service file not found${NC}"
    exit 1
fi

echo ""
echo -e "${YELLOW}Step 6: Reloading systemd${NC}"
systemctl daemon-reload
echo -e "${GREEN}✓ Systemd reloaded${NC}"

echo ""
echo -e "${GREEN}================================================${NC}"
echo -e "${GREEN}Installation complete!${NC}"
echo -e "${GREEN}================================================${NC}"
echo ""
echo "Next steps:"
echo ""
echo "1. Edit configuration (if needed):"
echo -e "   ${YELLOW}sudo nano $CONFIG_DIR/config.yaml${NC}"
echo ""
echo "2. Enable service to start on boot:"
echo -e "   ${YELLOW}sudo systemctl enable socksbalance${NC}"
echo ""
echo "3. Start the service:"
echo -e "   ${YELLOW}sudo systemctl start socksbalance${NC}"
echo ""
echo "4. Check service status:"
echo -e "   ${YELLOW}sudo systemctl status socksbalance${NC}"
echo ""
echo "5. View logs:"
echo -e "   ${YELLOW}sudo journalctl -u socksbalance -f${NC}"
echo ""
echo "Useful commands:"
echo -e "   Stop:    ${YELLOW}sudo systemctl stop socksbalance${NC}"
echo -e "   Restart: ${YELLOW}sudo systemctl restart socksbalance${NC}"
echo -e "   Disable: ${YELLOW}sudo systemctl disable socksbalance${NC}"
echo ""
