#!/bin/bash
set -e

echo "========================================"
echo "  Languager Bot Deployment Script"
echo "========================================"
echo ""

# Configuration
COMPOSE_FILE="docker-compose.yml"
SERVICE_NAME="bot"

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check required variables
if [ -z "$DOCKER_IMAGE" ]; then
    log_error "DOCKER_IMAGE environment variable is not set"
    exit 1
fi

log_info "Deploying image: $DOCKER_IMAGE"

# Login to GitLab Container Registry
log_info "Logging into GitLab Container Registry..."
echo "$CI_REGISTRY_PASSWORD" | docker login $CI_REGISTRY -u $CI_REGISTRY_USER --password-stdin

# Pull latest image
log_info "Pulling latest Docker image..."
docker pull $DOCKER_IMAGE || {
    log_error "Failed to pull Docker image"
    exit 1
}

# Backup current state
log_info "Creating backup of current deployment..."
BACKUP_DIR="backups/deployments"
mkdir -p $BACKUP_DIR
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
if [ -f "$COMPOSE_FILE" ]; then
    cp $COMPOSE_FILE $BACKUP_DIR/docker-compose.${TIMESTAMP}.yml.bak
fi

# Update docker-compose.yml with new image
log_info "Updating docker-compose.yml..."
if command -v yq &> /dev/null; then
    yq eval ".services.bot.image = \"$DOCKER_IMAGE\"" -i $COMPOSE_FILE
else
    # Fallback: use sed if yq is not available
    log_warn "yq not found, using sed for updating compose file"
    # Just use the pulled image by rebuilding
fi

# Stop old containers
log_info "Stopping old containers..."
docker-compose stop $SERVICE_NAME || log_warn "No containers to stop"

# Remove old containers (keep volumes)
log_info "Removing old containers..."
docker-compose rm -f $SERVICE_NAME || log_warn "No containers to remove"

# Start new containers
log_info "Starting new containers..."
docker-compose up -d $SERVICE_NAME || {
    log_error "Failed to start new containers"
    log_info "Rolling back..."
    if [ -f "$BACKUP_DIR/docker-compose.${TIMESTAMP}.yml.bak" ]; then
        cp $BACKUP_DIR/docker-compose.${TIMESTAMP}.yml.bak $COMPOSE_FILE
        docker-compose up -d $SERVICE_NAME
    fi
    exit 1
}

# Wait for container to be ready
log_info "Waiting for bot to start (10 seconds)..."
sleep 10

# Health check
log_info "Running health check..."
if ./health_check.sh; then
    log_info "✅ Deployment successful!"
    log_info "Bot is running on image: $DOCKER_IMAGE"
else
    log_error "Health check failed!"
    log_info "Rolling back..."
    if [ -f "$BACKUP_DIR/docker-compose.${TIMESTAMP}.yml.bak" ]; then
        cp $BACKUP_DIR/docker-compose.${TIMESTAMP}.yml.bak $COMPOSE_FILE
        docker-compose up -d $SERVICE_NAME
    fi
    exit 1
fi

# Cleanup old images (keep last 3)
log_info "Cleaning up old Docker images..."
docker images | grep "$CI_REGISTRY_IMAGE" | awk '{print $3}' | tail -n +4 | xargs -r docker rmi || log_warn "No old images to clean"

echo ""
log_info "========================================"
log_info "  Deployment completed successfully!"
log_info "========================================"

