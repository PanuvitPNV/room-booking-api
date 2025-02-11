.PHONY: swagger docker-build docker-up docker-down

swagger:
	swag init -g cmd/main.go

docker-build:
	docker compose -f .docker/docker-compose.yml build

docker-up:
	docker compose -f .docker/docker-compose.yml up -d

docker-down:
	docker compose -f .docker/docker-compose.yml down

logs:
	docker compose -f .docker/docker-compose.yml logs -f

migrate:
	docker compose -f .docker/docker-compose.yml up migration

seed:
	docker compose -f .docker/docker-compose.yml up seeder