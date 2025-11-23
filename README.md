# PR Reviewer Assignment Service

Репозиторий кандидата (Степичев Олег) по выполнению тестового задания «Backend-trainee-assignment-autumn-2025» (см. [Backend-trainee-assignment-autumn-2025.md](https://github.com/Leganyst/avitoTrainee/blob/main/Backend-trainee-assignment-autumn-2025.md)). Сервис автоматизирует назначение ревьюверов на PR, управление командами и активностью пользователей через HTTP API. Исходники: https://github.com/Leganyst/avitoTrainee

## Стек и обоснование
- **Go** — быстрая сборка, простая конкурентность, небольшой рантайм.
- **Gin** — лёгкий HTTP-фреймворк с удобным роутингом и middleware.
- **GORM + PostgreSQL** — ORM ускоряет CRUD и миграции; Postgres покрывает реляционные требования, many-to-many (pr_reviewers) и индексы.
- **Swagger (swaggo)** — автогенерация и UI для документации.

## Навигация по проекту
- **ТЗ:** [Backend-trainee-assignment-autumn-2025.md](https://github.com/Leganyst/avitoTrainee/blob/main/Backend-trainee-assignment-autumn-2025.md) — полный текст задания и нефункциональные требования.
- **Чек-лист:** [checkpoints.md](https://github.com/Leganyst/avitoTrainee/blob/main/checkpoints.md) — актуальный прогресс по требованиям, помогает понять структуру и реализацию.
- **API спецификация:** [openapi.yml](https://github.com/Leganyst/avitoTrainee/blob/main/openapi.yml) + автосгенерированные swagger-доки (см. Makefile/`make docs`).
- **Запуск:** см. [Makefile](https://github.com/Leganyst/avitoTrainee/blob/main/Makefile) таргет `up` (сборка + docs + docker-compose) или `run` для локального бинарного файла.
- **Код:** https://github.com/Leganyst/avitoTrainee — быстрые переходы к файлам по ссылкам выше.

## Требования
- Go 1.25+
- Docker и docker-compose
- При локальном запуске без Docker — PostgreSQL с параметрами из `.env` или дефолтов `config.Load()`.

## Быстрый старт
Действия для старта:
```bash
make up            # создать .env при необходимости, сгенерировать swagger, собрать и поднять docker-compose
```

Дополнительные действия:
```bash
make run           # локальный запуск бинарного файла (использует .env)
make unit-cover    # юнит-тесты + html coverage
make integration-cover  # интеграционные тесты (поднимают тестовый compose) + html coverage
```

## Запуск через чистый docker-compose (без Makefile)
1. Необходимо убедиться, что установлены Docker и docker-compose.
2. При отсутствии `.env` нужно создать файл: `cp .env-example .env` (значения согласованы для compose: `DB_HOST=db`, `POSTGRES_*` совпадают с `DB_*`).
3. Собрать и поднять сервис:  
   ```bash
   docker compose up --build
   ```
4. После запуска сервис будет доступен на `http://localhost:8080`, Postgres — на `localhost:5432` (см. порты в `docker-compose.yml`).
5. Для остановки использовать: `docker compose down`.

## Линтеры и форматирование
- Используются стандартные инструменты расширения Go для VS Code (от Microsoft) с `gofmt`/`goimports` по умолчанию.
- Отдельного конфига линтера (`golangci-lint` и т.п.) нет; правил сверх стандартных не вводилось.

## Документация и тесты
- Сгенерировать swagger: `make docs`.
- Открыть swagger UI: http://localhost:8080/swagger/ (при запущенном сервисе). Swagger генерируется из кода, так как фактический API расширен относительно исходного `openapi.yml` (статистика, массовая деактивация).
- Интеграционные сценарии через HTTP — в [test/pr_controller_integration_test.go](https://github.com/Leganyst/avitoTrainee/blob/main/test/pr_controller_integration_test.go).
- Юнит-тесты сервисного слоя — например [internal/service/pr_test.go](https://github.com/Leganyst/avitoTrainee/blob/main/internal/service/pr_test.go) / `team_test.go` / `user_test.go`.

## Обоснование архитектурных решений и процента покрытия кода 
- **Почему структур больше, чем в исходном openapi.yml:** базовый `openapi.yml` — это входное ТЗ. В процессе реализации добавлены дополнительные DTO (статистика, массовая деактивация), поэтому модельная часть расширена сверх исходной спецификации, чтобы покрыть новые ручки.
- **Почему генерируется своя документация:** итоговый API отличается от исходного задания (добавлены новые эндпоинты), поэтому swagger собирается из кода (`make docs`), чтобы документация соответствовала фактическим маршрутам и DTO.
- **Почему не использована кодогенерация по выданному OpenAPI:** исходный `openapi.yml` — входной артефакт, но реализация расширена. Кодогенерация по нему дала бы несоответствие с новыми ручками; проще поддерживать DTO/handlers вручную и генерировать swagger из кода.
- **Почему unit-тесты в основном на service layer:** сервисный слой содержит бизнес-правила (статусы PR, выбор ревьюверов, доменные ошибки). Репозитории обёрнуты GORM и проверяются через интеграционные тесты; тесты на слой контроллеров покрыты интеграциями. Поэтому юниты сфокусированы на бизнес-логике.
- **Почему нет продвинутого DI:** проект небольшой; зависимости прокидываются вручную в `main.go`/роутер и в тестовых стабах. Вводить контейнер DI избыточно для текущего объёма.
- **Логгер в глобальном контексте:** использован глобальный zap-синглтон (`config.Logger()`), чтобы не тянуть его через каждый метод. Для этого размера проекта это упрощает код; при масштабировании можно перейти на явное внедрение логгера.
- **Почему тесты фокусируются на PR-флоу:** ключевой сценарий ТЗ — назначение ревьюверов и операции с PR. Покрыты create/reassign/merge, включая ошибки. Дополнительные фичи (bulk deactivate, stats) покрыты интеграционно/нагрузочно; оставшиеся части (например, все ветки stats) можно расширять при дальнейшем развитии.
- **Интеграционный нагрузочный тест:** в [test/pr_controller_integration_test.go](https://github.com/Leganyst/avitoTrainee/blob/main/test/pr_controller_integration_test.go) есть сценарий, который поднимает тестовый сервер на реальной БД, создаёт 10 команд по 10 пользователей и 30 открытых PR, затем массово деактивирует пользователей и проверяет успешность и укладывание в 100 мс. Это эмулирует среднюю нагрузку и проверяет SLA/SLI.
