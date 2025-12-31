# Project Context: Go Fiber Template

## Overview
This project is a REST API starter template built with **Go (Golang)** and **Fiber v2**, using **MongoDB** as the primary database. It is designed to support multi-tenancy via "Units" (tenants) and includes built-in features for authentication, role management, and integration with external services like Cloudflare and MinIO.

## Tech Stack
*   **Language:** Go (1.24+)
*   **Web Framework:** [Fiber v2](https://github.com/gofiber/fiber)
*   **Database:** MongoDB (official `mongo-driver`)
*   **Authentication:** JWT (JSON Web Tokens)
*   **Storage:** MinIO / S3 compatible object storage
*   **Infrastructure:** Docker & Docker Compose

## Project Structure
The project follows a layered architecture (Controller-Repository pattern):

*   **`main.go`**: Application entry point. Initializes DB, middleware, and registers routes.
*   **`config/`**: Configuration logic (e.g., MongoDB connection, environment variables).
*   **`controllers/`**: HTTP request handlers. Responsible for parsing requests, calling repositories/services, and returning responses.
*   **`models/`**: Go structs representing database entities and API DTOs (Data Transfer Objects).
*   **`repositories/`**: Data access layer. direct interaction with MongoDB collections.
*   **`routes/`**: Route definitions. Maps HTTP endpoints to Controller methods and applies middleware.
*   **`pkg/`**: Shared utility packages:
    *   `auth/`: JWT middleware and verification logic.
    *   `response/`: Standardized JSON response helpers (`Success`, `Error`, `ListData`).
    *   `cloudflare/`: DNS management for unit subdomains.
    *   `httpcache/`: In-memory HTTP response caching.
*   **`seed/`**: Scripts to seed initial database data (e.g., default admins).

## Key Features

### 1. Multi-tenancy (Units)
The system is built around "Units". A Unit represents a tenant or organization.
*   **Subdomains:** Units are identified by unique subdomains.
*   **Cloudflare Integration:** The `UnitController` can automatically provision DNS CNAME records via Cloudflare API when a new Unit is created.

### 2. Authentication & Authorization
*   **JWT:** Uses `golang-jwt` for token generation and validation.
*   **Roles:**
    *   **Super Admin:** Global administrators.
    *   **Unit User/Admin:** Users scoped to a specific Unit.
*   **Middleware:** `auth.RequireAdmin` and `auth.RequireUser` middleware protect routes.

### 3. Caching
*   In-memory HTTP cache is implemented in `pkg/httpcache` to speed up GET requests.
*   Automatically invalidated on write operations (POST, PUT, DELETE).

## Building and Running

### Prerequisites
*   Go 1.24+
*   MongoDB (local or remote)
*   MinIO (optional, for uploads)

### Local Development
1.  **Setup Environment:**
    Copy `env` (if available) or create a `.env` file based on the keys used in `docker-compose.yml`.
    ```bash
    cp env .env
    ```
2.  **Run Application:**
    ```bash
    go run main.go
    ```
    The server typically starts on port `4000` (defined in `.env` or defaults to 4000).

### Docker
To run the full stack (App + MongoDB + MinIO):
```bash
docker-compose up -d
```

## Configuration (.env)
Key environment variables include:

*   **Database:** `MONGO_URL`, `MONGO_NAME`
*   **Server:** `PORT`, `APP_ENV` (dev/production)
*   **Auth:** `JWT_SECRET`, `JWT_SECRET_ADMIN`
*   **Cloudflare:** `CLOUDFLARE_API_TOKEN`, `CLOUDFLARE_ZONE_ID`, `CLOUDFLARE_ROOT_DOMAIN` (for dynamic subdomains)
*   **Storage:** `MINIO_ENDPOINT`, `MINIO_ACCESS_KEY`, `MINIO_SECRET_KEY`, `MINIO_BUCKET`

## Development Conventions

### Adding a New Feature
1.  **Model:** Define the data structure in `models/`.
2.  **Repository:** Create a repository in `repositories/` to handle DB operations for the model.
3.  **Controller:** Create a controller in `controllers/` to handle business logic and HTTP interactions.
4.  **Routes:** Define endpoints in `routes/` and register them in `main.go`.

### Response Format
Use the `pkg/response` helpers for consistent API output:
*   `response.Success(c, data, "message")`
*   `response.Error(c, "message", statusCode, details)`

### Database
*   Use `primitive.ObjectID` for IDs.
*   Ensure proper indexing in the repository constructor (see `repositories/unit_repository.go` as an example).
