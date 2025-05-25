# AccountService

#### <b>AccountService</b> — это микросервис, реализующий управление учетными записями пользователей, построенный с использованием языка Go и архитектурного подхода Clean Architecture. Проект включает в себя поддержку gRPC API, миграции базы данных, Docker-окружение и CI/CD с использованием GitLab CI.


## OpenApi документация в [тут](./api/account.swagger.json)

#### 📦 Стек технологий
- Язык программирования: Go

- Архитектура: Clean Architecture

- API: gRPC

- База данных: PostgreSQL

- Контейнеризация: Docker, Docker Compose

- CI/CD: GitLab CI

- Инструменты: Makefile, golangci-lint

### 📁 Структура проекта

    ├── api/                 # gRPC-протоколы и OpenAPI  документация
    ├── certs/               # ключи для jwt
    ├── cmd/                 # Точка входа в приложение
    ├── db/migrations/       # Миграции базы данных
    ├── google/api/          # Протоколы Google API (опционально)
    ├── internal/            # Бизнес-логика и слои приложения
    ├── pkg/                 # Переиспользуемые пакеты
    ├── coverage/            # Отчеты о покрытии тестами
    ├── .env.example         # Пример файла переменных окружения
    ├── Dockerfile           # Docker-конфигурация для сборки образа
    ├── docker-compose.yml   # Docker Compose для локального запуска
    ├── Makefile             # Утилиты для сборки и запуска
    ├── .gitlab-ci.yml       # Конфигурация CI/CD
    ├── .golangci.yml        # Настройки линтера
    ├── LICENSE              # Лицензия MIT
    └── README.md            # Документация проекта


------

## 🚀 Быстрый старт
#### 1. Клонируйте репозиторий:


    git clone https://github.com/SussyaPusya/AccountService.git
    cd AccountService

#### 2.Создайте файл .env на основе примера:

    cp .env.example .env
#### Запустите сервис с помощью Docker Compose:


    docker-compose up --build


#### Используйте gRPC-клиент, такой как grpcurl илии postman, для тестирования методов API.