#!/bin/bash

# Health check script for Languager Bot

# Ensure we're in the project directory
cd /opt/LanguagerEN2 || {
    echo "ERROR: Cannot change to /opt/LanguagerEN2"
    exit 1
}

SERVICE_NAME="languager_bot"
MAX_RETRIES=5
RETRY_DELAY=2

echo "Running health check for $SERVICE_NAME..."
echo "Current directory: $(pwd)"

# Check if container is running
if ! docker ps --format '{{.Names}}' | grep -q "^${SERVICE_NAME}$"; then
    echo "❌ Container $SERVICE_NAME is not running"
    exit 1
fi

echo "✅ Container is running"

# Check container status
STATUS=$(docker inspect --format='{{.State.Status}}' $SERVICE_NAME)
if [ "$STATUS" != "running" ]; then
    echo "❌ Container status: $STATUS"
    exit 1
fi

echo "✅ Container status is healthy"

# Check logs for errors
LOGS=$(docker logs --tail=50 $SERVICE_NAME 2>&1)

if echo "$LOGS" | grep -qi "error\|fatal\|panic"; then
    ERRORS=$(echo "$LOGS" | grep -i "error\|fatal\|panic" | tail -5)
    echo "⚠️  Found errors in logs:"
    echo "$ERRORS"
    echo ""
    echo "Checking if bot started successfully..."
fi

# Check if bot started successfully
if echo "$LOGS" | grep -q "Bot started successfully"; then
    echo "✅ Bot started successfully"
elif echo "$LOGS" | grep -q "Starting Languager Bot"; then
    echo "✅ Bot is starting..."
else
    echo "❌ Bot startup message not found in logs"
    echo "Last 10 lines of logs:"
    docker logs --tail=10 $SERVICE_NAME
    exit 1
fi

# Check container restart count
RESTART_COUNT=$(docker inspect --format='{{.RestartCount}}' $SERVICE_NAME)
if [ "$RESTART_COUNT" -gt 3 ]; then
    echo "⚠️  Warning: Container has restarted $RESTART_COUNT times"
fi

# Check PostgreSQL connectivity (check if postgres container is running)
if docker ps --format '{{.Names}}' | grep -q "languager_postgres"; then
    echo "✅ PostgreSQL container is running"
else
    echo "❌ PostgreSQL container is not running"
    exit 1
fi

echo ""
echo "✅ All health checks passed!"
exit 0

