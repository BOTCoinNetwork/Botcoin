#!/bin/bash

# stop
echo "Stopping botcoin service..."
systemctl stop botcoin || true 
# Waiting
for i in {1..10}; do
    if ! systemctl is-active --quiet botcoin; then
        break
    fi
    sleep 1
    echo "Waiting for botcoin service to stop... ($i/10)"
done

echo "Stopping existing botcoin processes..."
pkill -9 -f "botcoin" && echo "Killed botcoin processes" || echo "No botcoin processes found"
echo "Stopping processes on port 8080..."
fuser -k 8080/tcp && echo "Killed processes on port 8080" || echo "No processes found on port 8080"

# clear log
echo "Backing up log files..."
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
[ -f "/www/wwwroot/Botcoin/cmd/botcoin/out.log" ] && mv "/www/wwwroot/Botcoin/cmd/botcoin/out.log" "/www/wwwroot/Botcoin/cmd/botcoin/out.log.$TIMESTAMP"
[ -f "/www/wwwroot/Botcoin/cmd/botcoin/info.log" ] && mv "/www/wwwroot/Botcoin/cmd/botcoin/info.log" "/www/wwwroot/Botcoin/cmd/botcoin/info.log.$TIMESTAMP"

echo "Cleaning up log files..."
rm -f /www/wwwroot/Botcoin/cmd/botcoin/out.log /www/wwwroot/Botcoin/cmd/botcoin/info.log

# start
echo "Starting botcoin service..."
systemctl start botcoin

# Checking
echo "Checking service status..."
systemctl status botcoin --no-pager