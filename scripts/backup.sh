#!/bin/sh

# PostgreSQL backup script
# Runs continuously, creating backups every 24 hours

RETENTION_DAYS=${BACKUP_RETENTION_DAYS:-30}

echo "Backup service started"
echo "Backup directory: /backups"
echo "Retention period: $RETENTION_DAYS days"
echo "Database: $PGDATABASE"
echo ""

# Wait for PostgreSQL to be ready
echo "Waiting for PostgreSQL to be ready..."
until pg_isready -h $PGHOST -U $PGUSER -d $PGDATABASE; do
  echo "PostgreSQL is unavailable - sleeping"
  sleep 2
done

echo "PostgreSQL is ready!"
echo ""

# Main backup loop
while true; do
    DATE=$(date +%Y%m%d_%H%M%S)
    BACKUP_FILE="/backups/backup_${DATE}.sql"
    
    echo "[$(date)] Creating backup: $BACKUP_FILE"
    
    # Create backup
    if pg_dump -h $PGHOST -U $PGUSER -d $PGDATABASE > "$BACKUP_FILE"; then
        echo "[$(date)] Backup created successfully: $BACKUP_FILE"
        
        # Get backup size
        SIZE=$(du -h "$BACKUP_FILE" | cut -f1)
        echo "[$(date)] Backup size: $SIZE"
        
        # Clean old backups (keep only last N backups based on retention days)
        KEEP_COUNT=$((RETENTION_DAYS))
        OLD_BACKUPS=$(ls -t /backups/backup_*.sql 2>/dev/null | tail -n +$((KEEP_COUNT + 1)))
        
        if [ -n "$OLD_BACKUPS" ]; then
            echo "[$(date)] Cleaning old backups..."
            echo "$OLD_BACKUPS" | xargs rm -f
            echo "[$(date)] Old backups removed"
        fi
        
        # Show current backup count
        BACKUP_COUNT=$(ls -1 /backups/backup_*.sql 2>/dev/null | wc -l)
        echo "[$(date)] Total backups: $BACKUP_COUNT"
    else
        echo "[$(date)] ERROR: Backup failed!"
    fi
    
    echo "[$(date)] Next backup in 24 hours..."
    echo ""
    
    # Sleep for 24 hours
    sleep 86400
done

