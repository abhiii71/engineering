# Account Service (standalone)

A **monolithic** Go project with a **REST API** for accounts: register, login, and account retrieval. Uses PostgreSQL. Not a microservice—single app, HTTP/JSON only.

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

### Docker Compose reference

The project includes a `docker-compose.yml` that runs the **single-server** setup: one container runs both the account service and PostgreSQL.

| Item | Description |
|------|-------------|
| **Compose file** | `docker-compose.yml` |
| **Dockerfile** | `Dockerfile.single` (builds app and runs Postgres inside the same image) |
| **Service name** | `server` (single service) |
| **Port** | Host `8080` → container `8080` (override in `ports` if needed) |
| **Volume** | `pgdata` → `/app/pgdata` (PostgreSQL data; persists across restarts) |

**Commands**

| Command | Description |
|---------|-------------|
| `docker compose up -d --build` | Build image and start the container in the background |
| `docker compose up --build` | Build and run in foreground (logs in terminal) |
| `docker compose down` | Stop and remove the container |
| `docker compose down -v` | Stop and remove the container **and** the `pgdata` volume (full reset) |
| `docker compose logs -f server` | Stream logs from the `server` service |

**Environment variables** (set in `docker-compose.yml` or override with `.env` / `environment`)

| Variable | Default in compose | Description |
|----------|--------------------|-------------|
| `PORT` | `8080` | HTTP port inside the container |
| `SECRET_KEY` | `change-me-in-production` | JWT signing secret; set a strong value in production |
| `ISSUER` | `account-service` | JWT issuer claim |
| `DATABASE_URL` | (set in entrypoint) | Postgres URL; entrypoint defaults to `postgres://account:secret@localhost:5432/accountdb?sslmode=disable` inside the container |

**Overriding port or env**

- Change host port: edit `ports` in `docker-compose.yml`, e.g. `"8082:8080"`.
- Override env: add to the `environment` section or use an `.env` file in the same directory and run `docker compose up -d --build`.

**Migrations**

On first start, the container entrypoint runs migrations under `/app/migrations/` (e.g. `000001_create_accounts_table.up.sql`, `000002_create_transactions_and_activity_log.up.sql`). No manual migration step is needed when using Docker Compose.

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

| Method | Path                         | Description        |
|--------|------------------------------|--------------------|
| POST   | /register                    | Register; body: `{"name","email","password"}` → `{"token"}` |
| POST   | /login                       | Login; body: `{"email","password"}` → `{"token"}` |
| GET    | /accounts/{id}               | Get account by ID → `{"id","name","email"}` |
| GET    | /accounts                    | List accounts; query: `?skip=0&take=10` → `{"accounts":[...]}` |
| DELETE | /accounts/{id}               | Delete account (204 No Content) |
| POST   | /accounts/{id}/transactions  | Record a transaction; body: `{"amount_cents","kind":"credit\|debit","description"}` → transaction |
| GET    | /accounts/{id}/transactions  | List transactions; query: `?skip=0&take=10` → `{"transactions":[...]}` |
| POST   | /accounts/{id}/activity      | Log activity; body: `{"action","ip_address"}` → activity |
| GET    | /accounts/{id}/activity      | List activity; query: `?skip=0&take=10` → `{"activity":[...]}` |

All request/response bodies are JSON. Use **transactions** and **activity** endpoints to simulate concurrent writes and observe failure/load behaviour (e.g. run many simultaneous POSTs to the same account).

---

## Testing the API

Base URL (local): **`http://localhost:8080`**. All requests use `Content-Type: application/json` unless noted.

### 1. Register

Creates a new account and returns a JWT.

| Item | Value |
|------|--------|
| **Method** | `POST` |
| **Path** | `/register` |
| **Request body** | `name` (string), `email` (string), `password` (string) |

**Example request**

```bash
curl -s -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice","email":"alice@example.com","password":"secret123"}'
```

**Example response** (201 Created)

```json
{"token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."}
```

**Errors**

- **400** — Missing or invalid body: `{"error":"name, email and password required"}`
- **409** — Email already registered: `{"error":"account already exists"}`

---

### 2. Login

Authenticates with email/password and returns a JWT.

| Item | Value |
|------|--------|
| **Method** | `POST` |
| **Path** | `/login` |
| **Request body** | `email` (string), `password` (string) |

**Example request**

```bash
curl -s -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email":"alice@example.com","password":"secret123"}'
```

**Example response** (200 OK)

```json
{"token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."}
```

**Errors**

- **400** — Missing fields: `{"error":"email and password required"}`
- **401** — Wrong email or password: `{"error":"invalid email or password"}`

---

### 3. Get account by ID

Returns a single account by numeric ID (no auth required in current implementation).

| Item | Value |
|------|--------|
| **Method** | `GET` |
| **Path** | `/accounts/{id}` |
| **Path params** | `id` — account ID (integer) |

**Example request**

```bash
curl -s http://localhost:8080/accounts/1
```

**Example response** (200 OK)

```json
{"id":1,"name":"Alice","email":"alice@example.com"}
```

**Errors**

- **400** — Missing or invalid id: `{"error":"id required"}` or `{"error":"invalid id"}`
- **404** — Account not found: `{"error":"not found"}`

---

### 4. List accounts

Returns a paginated list of accounts.

| Item | Value |
|------|--------|
| **Method** | `GET` |
| **Path** | `/accounts` |
| **Query params** | `skip` (optional, default 0), `take` (optional, default 100) |

**Example request**

```bash
curl -s "http://localhost:8080/accounts?skip=0&take=10"
```

**Example response** (200 OK)

```json
{
  "accounts": [
    {"id":1,"name":"Alice","email":"alice@example.com"},
    {"id":2,"name":"Bob","email":"bob@example.com"}
  ]
}
```

**Errors**

- **500** — Server error: `{"error":"failed to list accounts"}`

---

### 5. Delete account

Deletes an account by ID. Related transactions and activity logs are removed (CASCADE).

| Item | Value |
|------|--------|
| **Method** | `DELETE` |
| **Path** | `/accounts/{id}` |
| **Path params** | `id` — account ID (integer) |

**Example request**

```bash
curl -s -X DELETE http://localhost:8080/accounts/1
```

**Example response** (204 No Content) — empty body.

**Errors**

- **400** — Invalid id: `{"error":"id required"}` or `{"error":"invalid id"}`
- **404** — Account not found: `{"error":"not found"}`
- **500** — Server error: `{"error":"failed to delete account"}`

---

### 6. Record transaction

Adds a credit or debit transaction for an account.

| Item | Value |
|------|--------|
| **Method** | `POST` |
| **Path** | `/accounts/{id}/transactions` |
| **Path params** | `id` — account ID |
| **Request body** | `amount_cents` (integer), `kind` ("credit" or "debit"), `description` (string, optional) |

**Example request**

```bash
curl -s -X POST http://localhost:8080/accounts/1/transactions \
  -H "Content-Type: application/json" \
  -d '{"amount_cents":1000,"kind":"credit","description":"Initial deposit"}'
```

**Example response** (201 Created)

```json
{
  "id":1,
  "account_id":1,
  "amount_cents":1000,
  "kind":"credit",
  "description":"Initial deposit",
  "created_at":"2025-03-01T12:00:00Z"
}
```

**Errors**

- **400** — Invalid id or body: `{"error":"kind required (credit or debit)"}` or `{"error":"kind must be credit or debit"}`
- **500** — Server error: `{"error":"failed to record transaction"}`

---

### 7. List transactions

Returns paginated transactions for an account.

| Item | Value |
|------|--------|
| **Method** | `GET` |
| **Path** | `/accounts/{id}/transactions` |
| **Path params** | `id` — account ID |
| **Query params** | `skip` (optional), `take` (optional, default 100) |

**Example request**

```bash
curl -s "http://localhost:8080/accounts/1/transactions?skip=0&take=10"
```

**Example response** (200 OK)

```json
{
  "transactions": [
    {
      "id":1,
      "account_id":1,
      "amount_cents":1000,
      "kind":"credit",
      "description":"Initial deposit",
      "created_at":"2025-03-01T12:00:00Z"
    }
  ]
}
```

**Errors**

- **400** — Invalid id: `{"error":"id required"}` or `{"error":"invalid id"}`
- **500** — Server error: `{"error":"failed to list transactions"}`

---

### 8. Record activity

Logs an activity event for an account (e.g. login, IP).

| Item | Value |
|------|--------|
| **Method** | `POST` |
| **Path** | `/accounts/{id}/activity` |
| **Path params** | `id` — account ID |
| **Request body** | `action` (string), `ip_address` (string, optional) |

**Example request**

```bash
curl -s -X POST http://localhost:8080/accounts/1/activity \
  -H "Content-Type: application/json" \
  -d '{"action":"login","ip_address":"192.168.1.1"}'
```

**Example response** (201 Created)

```json
{
  "id":1,
  "account_id":1,
  "action":"login",
  "ip_address":"192.168.1.1",
  "created_at":"2025-03-01T12:00:00Z"
}
```

**Errors**

- **400** — Missing action or invalid id: `{"error":"action required"}`
- **500** — Server error: `{"error":"failed to record activity"}`

---

### 9. List activity

Returns paginated activity log entries for an account.

| Item | Value |
|------|--------|
| **Method** | `GET` |
| **Path** | `/accounts/{id}/activity` |
| **Path params** | `id` — account ID |
| **Query params** | `skip` (optional), `take` (optional, default 100) |

**Example request**

```bash
curl -s "http://localhost:8080/accounts/1/activity?skip=0&take=10"
```

**Example response** (200 OK)

```json
{
  "activity": [
    {
      "id":1,
      "account_id":1,
      "action":"login",
      "ip_address":"192.168.1.1",
      "created_at":"2025-03-01T12:00:00Z"
    }
  ]
}
```

**Errors**

- **400** — Invalid id: `{"error":"id required"}` or `{"error":"invalid id"}`
- **500** — Server error: `{"error":"failed to list activity"}`

---

### Full test flow (copy-paste)

Run with the service at `http://localhost:8080` (e.g. `docker compose up -d --build`).

```bash
# Register
curl -s -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Test User","email":"test@example.com","password":"pass123"}'

# Login
curl -s -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"pass123"}'

# Get account (use id from DB or list; e.g. 1)
curl -s http://localhost:8080/accounts/1

# List accounts
curl -s "http://localhost:8080/accounts?skip=0&take=10"

# Add transaction
curl -s -X POST http://localhost:8080/accounts/1/transactions \
  -H "Content-Type: application/json" \
  -d '{"amount_cents":500,"kind":"credit","description":"Test"}'

# List transactions
curl -s "http://localhost:8080/accounts/1/transactions?skip=0&take=10"

# Log activity
curl -s -X POST http://localhost:8080/accounts/1/activity \
  -H "Content-Type: application/json" \
  -d '{"action":"api_test","ip_address":"127.0.0.1"}'

# List activity
curl -s "http://localhost:8080/accounts/1/activity?skip=0&take=10"

# Delete account (204, empty body)
curl -s -w "\nHTTP %{http_code}\n" -X DELETE http://localhost:8080/accounts/1
```

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
