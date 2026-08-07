MIGRATE_DIR := migrations

# shell не читает .env автоматически — вытаскиваем URL прямо из файла
DATABASE_URL := $(shell grep '^DATABASE_URL=' .env 2>/dev/null | cut -d= -f2-)

.PHONY: run migrate-up migrate-down migrate-down-all migrate-fresh migrate-force psql

run:
	go run .

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

migrate-force:
	@test -n "$(DATABASE_URL)" || (echo "DATABASE_URL not found in .env" && exit 1)
	migrate -path $(MIGRATE_DIR) -database "$(DATABASE_URL)" force $(VERSION)

psql:
	@test -n "$(DATABASE_URL)" || (echo "DATABASE_URL not found in .env" && exit 1)
	psql "$(DATABASE_URL)"
