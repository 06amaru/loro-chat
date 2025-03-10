# Chat Server

The loro chat server implements websockets to communicate between clients.

### Setup

1. Create a database in Postgres e.g. chat_experiment
2. Add .env file in root project
```
SIGNING_KEY=custom_key
DB_HOST=localhost
DB_PORT=5432
DB_NAME=chat_experiment
DB_USER=postgres
DB_PASSWORD=password
```
3. Execute ```go run cmd/main.go```
