#!/bin/bash

cd "$(dirname "$0")" || exit 1

if [ ! -d "./cmd/botcoin" ]; then
    echo "Error: cmd/botcoin directory not found!"
    exit 1
fi

# clear log
echo "Backing up log files..."
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
[ -f "./cmd/botcoin/out.log" ] && mv "./cmd/botcoin/out.log" "./cmd/botcoin/out.log.$TIMESTAMP"
[ -f "./cmd/botcoin/info.log" ] && mv "./cmd/botcoin/info.log" "./cmd/botcoin/info.log.$TIMESTAMP"

echo "Cleaning up log files..."
rm -f ./cmd/botcoin/out.log ./cmd/botcoin/info.log

# start
echo "Starting botcoin service..."
systemctl start botcoin

# Checking
echo "Checking service status..."
systemctl status botcoin --no-pager