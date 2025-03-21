# Chat Server

The loro chat server implements websockets to communicate between clients.

### Setup

1. Create a database in Postgres e.g. chat_experiment
2. Add .env file in root project
```
SIGNING_KEY=custom_key
DB_HOST=localhost
DB_PORT=5432
DB_NAME=postgres
DB_USER=postgres
DB_PASSWORD=password
GOOSE_DRIVER=postgres
GOOSE_DBSTRING=postgres://postgres:password@localhost:5432/postgres
GOOSE_MIGRATION_DIR=./migrations
```
3. Execute ```goose up```
4. Execute ```go run cmd/main.go```
