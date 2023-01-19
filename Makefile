DATABASE_DSN = "postgresql://ubuntu:ubuntu@192.168.1.15:5432/linebot?sslmode=disable"
MIGRATION_PATH = "./migrations"
migrate-up:
	goose -dir=$(MIGRATION_PATH) postgres $(DATABASE_DSN) up
migrate-up-1:
	goose -dir=$(MIGRATION_PATH) postgres $(DATABASE_DSN) up-by-one
migrate-down:
	goose -dir=$(MIGRATION_PATH) postgres $(DATABASE_DSN) down
migrate-status:
	goose -dir=$(MIGRATION_PATH) postgres $(DATABASE_DSN) status 

sqlc:
	sqlc generate
.PHONY: migrate-up migrate-down sqlc mock