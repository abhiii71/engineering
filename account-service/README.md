# Account Service (standalone)

A **monolithic** Go project with a **REST API** for accounts: register, login, and account retrieval. Uses PostgreSQL. Not a microservice—single app, HTTP/JSON only.

Lives inside the `orderStream` repo as a separate, self-contained project.

## Features

- Register a new account
- Login with email and password (returns JWT)
- Get account by ID
- List accounts with pagination
- PostgreSQL storage, REST API (JSON)

## Single server setup

A good way to learn system design is to start with a **single server**: one machine runs everything (web app, database, etc.). Later you can split things across many machines.

**Figure 1-1 — Single server (one container)**

```
                    ┌─────────────────────────────────────────┐
                    │           ONE CONTAINER                  │
   Internet         │  ┌─────────────┐    ┌───────────────┐  │
       │            │  │ Account     │    │   PostgreSQL  │  │
       ▼            │  │ Service     │───▶│   (database)  │  │
  ─────────         │  │ (REST API)  │    │               │  │
       │            │  └─────────────┘    └───────────────┘  │
       └───────────▶│         ▲                  │           │
                    │         │    same container             │
                    └─────────┼──────────────────┼───────────┘
                              │                  │
                         localhost           localhost
```

**Figure 1-2 — Request flow**

1. **Client** sends HTTP request (e.g. POST /register) to the server.
2. **Account service** receives it, talks to **PostgreSQL** on the same machine (e.g. insert user).
3. **PostgreSQL** responds to the service.
4. **Account service** returns HTTP response (e.g. JWT) to the client.

All of this happens in one container: one place to deploy, one place to monitor, and the app talks to the DB over localhost inside the same container.

The image is built from **Ubuntu**: we install PostgreSQL and Go manually in the Dockerfile (no separate Postgres image). One server = one container.

**Run the single server (one container = app + DB):**

```bash
docker compose up -d --build
```

Then call the API at `http://localhost:8080` (register, login, etc.). The database and app run inside the same container; the migration runs automatically on startup.

```bash
# Quick test
curl -s -X POST http://localhost:8080/register -H "Content-Type: application/json" \
  -d '{"name":"Test","email":"test@example.com","password":"secret"}'
```

Stop and remove: `docker compose down` (add `-v` to remove the database volume). If port 8080 is already in use, change the port in `docker-compose.yml` (e.g. `"8082:8080"`). Use **docker compose** (V2), not the old `docker-compose` (V1), to avoid compatibility errors with current Docker.

## Prerequisites

- Go ≥ 1.22
- PostgreSQL

## Setup

1. Copy env and set values:

   ```bash
   cp .env.example .env
   ```

2. Create the database and run migrations (e.g. apply `db/migrations/000001_create_accounts_table.up.sql`).

3. Run the service:

   ```bash
   go run ./cmd/account
   ```

## Run with Docker

1. Build the image:

   ```bash
   docker build -t account-service .
   ```

2. Run the server (set `DATABASE_URL` and optionally `SECRET_KEY`, `ISSUER`):

   ```bash
   docker run --rm -p 8080:8080 --env-file .env account-service
   ```

   Or pass env vars explicitly:

   ```bash
   docker run --rm -p 8080:8080 \
     -e DATABASE_URL="postgres://user:pass@host:5432/dbname?sslmode=disable" \
     -e SECRET_KEY="your-jwt-secret" \
     -e ISSUER="account-service" \
     account-service
   ```

   Ensure `DATABASE_URL` points to a Postgres host reachable from the container (e.g. use your machine IP or `host.docker.internal` instead of `localhost`).

### Quick test with Docker (from scratch)

```bash
# 1. Build image
docker build -t account-service .

# 2. Start Postgres and create network
docker network create account-net
docker run -d --name account-db --network account-net \
  -e POSTGRES_USER=account -e POSTGRES_PASSWORD=secret -e POSTGRES_DB=accountdb \
  postgres:16-alpine

# 3. Apply migration (after Postgres is ready, ~3s)
sleep 3
docker exec -i account-db psql -U account -d accountdb < db/migrations/000001_create_accounts_table.up.sql

# 4. Run the service
docker run -d --name account-svc --network account-net -p 8080:8080 \
  -e DATABASE_URL="postgres://account:secret@account-db:5432/accountdb?sslmode=disable" \
  -e SECRET_KEY=test-secret -e ISSUER=account-service \
  account-service

# 5. Test API
curl -s -X POST http://localhost:8080/register -H "Content-Type: application/json" \
  -d '{"name":"Test User","email":"test@example.com","password":"password123"}'
curl -s -X POST http://localhost:8080/login -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'
```

Cleanup: `docker rm -f account-svc account-db && docker network rm account-net`

## Environment

| Variable       | Description                  |
|----------------|------------------------------|
| `DATABASE_URL` | PostgreSQL connection string |
| `PORT`         | HTTP server port (default 8080) |
| `SECRET_KEY`   | JWT signing secret           |
| `ISSUER`       | JWT issuer (e.g. account-service) |

## REST API

| Method | Path           | Description        |
|--------|----------------|--------------------|
| POST   | /register      | Register; body: `{"name","email","password"}` → `{"token"}` |
| POST   | /login         | Login; body: `{"email","password"}` → `{"token"}` |
| GET    | /accounts/{id} | Get account by ID → `{"id","name","email"}` |
| GET    | /accounts      | List accounts; query: `?skip=0&take=10` → `{"accounts":[...]}` |

All request/response bodies are JSON.

## Project layout

```
account-service/
├── cmd/account/       # Entrypoint
├── config/            # Env/config
├── db/migrations/     # SQL migrations
├── internal/          # Repository, service, REST handlers + HTTP server
├── models/            # Account model
├── pkg/auth/          # JWT (GenerateToken, ValidateToken)
├── pkg/crypt/         # Password hashing
├── go.mod
└── README.md
```
