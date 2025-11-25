#!/bin/bash
set -e

RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; NC='\033[0m'

if [ -z "$DATABASE_URL" ]; then
  echo -e "${RED}❌ DATABASE_URL environment variable not set${NC}"
  echo "Example: export DATABASE_URL='postgres://user:pass@localhost:5432/dbname?sslmode=disable'"
  exit 1
fi

MIGRATIONS_PATH="${MIGRATIONS_PATH:-migrations}"
COMMAND=${1:-up}

echo -e "${YELLOW}🔧 Running migration: $COMMAND${NC}"
echo "Database: $DATABASE_URL"
echo "Migrations path: $MIGRATIONS_PATH"
echo ""

case "$COMMAND" in
  up)
    for f in $(ls "$MIGRATIONS_PATH"/*.up.sql 2>/dev/null | sort); do
      echo "Applying $f"
      psql "$DATABASE_URL" -f "$f"
    done
    ;;
  down)
    for f in $(ls "$MIGRATIONS_PATH"/*.down.sql 2>/dev/null | sort -r); do
      echo "Applying $f"
      psql "$DATABASE_URL" -f "$f"
    done
    ;;
  drop)
    echo "Dropping all known tables (use with care)"
    # optional: run specific drop SQL or down scripts
    for f in $(ls "$MIGRATIONS_PATH"/*.down.sql 2>/dev/null | sort -r); do
      echo "Applying $f"
      psql "$DATABASE_URL" -f "$f"
    done
    ;;
  version)
    psql "$DATABASE_URL" -c "SELECT current_database(), now();"
    ;;
  *)
    echo "Unknown command: $COMMAND"
    exit 2
    ;;
esac

echo -e "${GREEN}✅ Done${NC}"