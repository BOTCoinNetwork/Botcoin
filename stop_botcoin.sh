#!/bin/bash

set -e  # 

# stop
echo "Stopping botcoin service..."
systemctl stop botcoin || echo "Failed to stop service, continuing..."

# Waiting
for i in {1..10}; do
    if ! systemctl is-active --quiet botcoin; then
        echo "Service stopped successfully"
        break
    fi
    sleep 1
    echo "Waiting for botcoin service to stop... ($i/10)"
    [ $i -eq 10 ] && echo "Warning: Service did not stop gracefully"
done

echo "Stopping existing botcoin processes..."
pkill -9 -f "botcoin" >/dev/null 2>&1 && echo "Killed botcoin processes" || echo "No botcoin processes found"
echo "Stopping processes on port 8080..."
fuser -k 8080/tcp && echo "Killed processes on port 8080" || echo "No processes found on port 8080"

# ... rest of the script remains the same ...