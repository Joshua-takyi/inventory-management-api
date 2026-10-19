MIGRATIONS_DIR := internal/adapters/postgres/migrations

.PHONY: migration-create

migration-create:
	@test -n "$(name)" || (echo "Usage: make migration-create name=create_products"; exit 1)
	@mkdir -p $(MIGRATIONS_DIR)
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)