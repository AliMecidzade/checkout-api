MIGRATE_DIR := migrations

# shell не читает .env автоматически — вытаскиваем URL прямо из файла
DATABASE_URL := $(shell grep '^DATABASE_URL=' .env 2>/dev/null | cut -d= -f2-)

.PHONY: run migrate-up migrate-down migrate-down-all migrate-fresh migrate-recreate migrate-force psql

run:
	go run ./cmd/server

migrate-up:
	@test -n "$(DATABASE_URL)" || (echo "DATABASE_URL not found in .env" && exit 1)
	migrate -path $(MIGRATE_DIR) -database "$(DATABASE_URL)" up

migrate-down:
	@test -n "$(DATABASE_URL)" || (echo "DATABASE_URL not found in .env" && exit 1)
	migrate -path $(MIGRATE_DIR) -database "$(DATABASE_URL)" down 1

migrate-down-all:
	@test -n "$(DATABASE_URL)" || (echo "DATABASE_URL not found in .env" && exit 1)
	migrate -path $(MIGRATE_DIR) -database "$(DATABASE_URL)" down -all

migrate-fresh:
	@test -n "$(DATABASE_URL)" || (echo "DATABASE_URL not found in .env" && exit 1)
	migrate -path $(MIGRATE_DIR) -database "$(DATABASE_URL)" down -all
	migrate -path $(MIGRATE_DIR) -database "$(DATABASE_URL)" up

# Hard reset: drop the whole schema (tables, indexes, constraints, data)
# and re-apply all migrations from scratch. Does not rely on down migrations.
migrate-recreate:
	@test -n "$(DATABASE_URL)" || (echo "DATABASE_URL not found in .env" && exit 1)
	psql "$(DATABASE_URL)" -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
	migrate -path $(MIGRATE_DIR) -database "$(DATABASE_URL)" up

migrate-force:
	@test -n "$(DATABASE_URL)" || (echo "DATABASE_URL not found in .env" && exit 1)
	migrate -path $(MIGRATE_DIR) -database "$(DATABASE_URL)" force $(VERSION)

psql:
	@test -n "$(DATABASE_URL)" || (echo "DATABASE_URL not found in .env" && exit 1)
	psql "$(DATABASE_URL)"
