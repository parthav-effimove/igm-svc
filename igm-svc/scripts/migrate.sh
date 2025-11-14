#!/bin/bash
# filepath: scripts/migrate.sh

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if DATABASE_URL is set
if [ -z "$DATABASE_URL" ]; then
    echo -e "${RED}❌ DATABASE_URL environment variable not set${NC}"
    echo "Example: export DATABASE_URL='postgres://user:pass@localhost:5432/dbname?sslmode=disable'"
    exit 1
fi

MIGRATIONS_PATH="migrations"
COMMAND=${1:-"up"}

echo -e "${YELLOW}🔧 Running migration: $COMMAND${NC}"
echo "Database: $DATABASE_URL"
echo "Migrations path: $MIGRATIONS_PATH"
echo ""

case $COMMAND in
    up)
        echo -e "${YELLOW}📈 Applying all pending migrations...${NC}"
        migrate -path $MIGRATIONS_PATH -database "$DATABASE_URL" up
        echo -e "${GREEN}✅ Migrations applied successfully${NC}"
        ;;
    down)
        echo -e "${YELLOW}📉 Rolling back last migration...${NC}"
        migrate -path $MIGRATIONS_PATH -database "$DATABASE_URL" down 1
        echo -e "${GREEN}✅ Migration rolled back${NC}"
        ;;
    drop)
        echo -e "${RED}⚠️  WARNING: This will DROP all tables!${NC}"
        read -p "Are you sure? (yes/no): " confirm
        if [ "$confirm" = "yes" ]; then
            migrate -path $MIGRATIONS_PATH -database "$DATABASE_URL" drop -f
            echo -e "${GREEN}✅ Database dropped${NC}"
        else
            echo "Cancelled"
        fi
        ;;
    version)
        echo -e "${YELLOW}📍 Current migration version:${NC}"
        migrate -path $MIGRATIONS_PATH -database "$DATABASE_URL" version
        ;;
    force)
        if [ -z "$2" ]; then
            echo -e "${RED}❌ Usage: ./migrate.sh force <version>${NC}"
            exit 1
        fi
        echo -e "${YELLOW}🔨 Forcing migration to version $2${NC}"
        migrate -path $MIGRATIONS_PATH -database "$DATABASE_URL" force $2
        echo -e "${GREEN}✅ Migration forced to version $2${NC}"
        ;;
    create)
        if [ -z "$2" ]; then
            echo -e "${RED}❌ Usage: ./migrate.sh create <migration_name>${NC}"
            exit 1
        fi
        migrate create -ext sql -dir $MIGRATIONS_PATH -seq $2
        echo -e "${GREEN}✅ Migration files created${NC}"
        ;;
    *)
        echo -e "${RED}❌ Unknown command: $COMMAND${NC}"
        echo ""
        echo "Available commands:"
        echo "  up               - Apply all pending migrations"
        echo "  down             - Rollback last migration"
        echo "  drop             - Drop all tables (use with caution!)"
        echo "  version          - Show current migration version"
        echo "  force <version>  - Force migration to specific version"
        echo "  create <name>    - Create new migration files"
        exit 1
        ;;
esac