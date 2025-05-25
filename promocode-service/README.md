# Promo Code Service

## Запуск всего приложения локально через docker compose up -d
1) Создать .env файл в корне cp .env.example .env
2) Создать или запросить GitLab Access Token (https://gitlab.com/-/user_settings/personal_access_tokens)
3) Скопировать Access token в .env GITLAB_ACCESS_TOKEN
4) docker compose up --build

5) Запустить сервис account (ниже)
6) cd ../ && git clone https://gitlab.com/promo-code-service/accaunt-service.git
7) cd accaunt-service
8) cp .env.example .env
9) docker-compose up --build

Для остановки приложения
docker compose down
Для удаления контейнеров и volumes (для чистого запуска)
docker compose rm -v

## Локальная разработка
docker-compose up db redis -d