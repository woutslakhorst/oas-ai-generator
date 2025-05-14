# Petstore Server

This repository contains Go server stubs generated from the provided `petstore.yaml` OpenAPI specification. It includes models, a simple Gin-based HTTP server, and a SQLite database setup.

## Building

Ensure you have Go installed. Run:

```
go build ./cmd
```

## Running

Set `DATABASE_PATH` if you want to override the default SQLite database path (`petstore.db`). Then start the server:

```
go run ./cmd
```

The server listens on `:8080`.

## Migrations

SQL migration scripts are located in the `migrations` directory. The server applies them automatically on startup.
