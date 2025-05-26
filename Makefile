.PHONY: up down restart logs

up:
	
	docker compose -f account-service/docker-compose.yml up -d
	docker compose -f promocode-service/docker-compose.yml up -d
	docker compose -f auth-service/docker-compose.yml up -d
	docker compose -f nginx/docker-compose.yml up -d

down:
	docker compose -f nginx/docker-compose.yml down
	docker compose -f auth-service/docker-compose.yml down
	docker compose -f account-service/docker-compose.yml down
	docker compose -f promocode-service/docker-compose.yml down

restart: down up

logs:
	docker compose -f -f auth-service/docker-compose.yml logs -f
