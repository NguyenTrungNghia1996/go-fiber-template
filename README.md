# Go Fiber Template

This project provides a starting point for building REST APIs with
[Fiber](https://github.com/gofiber/fiber). It includes basic user
authentication, role-based authorization and MongoDB integration.
Users now store a `name` field and can belong to multiple role groups.
Admins can manage role groups with dedicated CRUD endpoints.

## Running locally

```bash
go run main.go
```

Create an `.env` file (see `env` for an example) containing your database
credentials.

## Postman Collection

To quickly explore the API you can import
`postman/go-fiber-template.postman_collection.json` into Postman. The collection
assumes two variables:

- `baseUrl` – base address of your running server, e.g. `http://localhost:4000`
- `token` – JWT obtained from the `Login` request

The collection contains examples for logging in, retrieving the current user
and managing users as an admin.
