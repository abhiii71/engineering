# Account Service (standalone)

A **monolithic** Go project with a **REST API** for accounts: register, login, and account retrieval. Uses PostgreSQL. Not a microservice—single app, HTTP/JSON only.

Lives inside the `orderStream` repo as a separate, self-contained project.

## Features

- Register a new account
- Login with email and password (returns JWT)
- Get account by ID
- List accounts with pagination
- PostgreSQL storage, REST API (JSON)

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
