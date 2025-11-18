default:
    @just --list

up:
    docker compose up -d --build

down:
    docker compose down

restart:
    docker compose down
    docker compose up -d

stop:
    docker compose stop

logs:
    docker compose logs -f

redis:
    docker compose exec redis redis-cli

insight:
    open http://localhost:5540 || xdg-open http://localhost:5540 || true

clean:
    docker compose down -v
    rm -rf .redis

env:
    source .env

embedder:
    .venv/bin/python embedder/embedder.py

quality:
    .venv/bin/python embedder/quality.py

migrate:
    .venv/bin/python migrations/categories.py
    .venv/bin/python migrations/places.py
