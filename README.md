# Go Fiber Template

This project provides a starting point for building REST APIs with
[Fiber](https://github.com/gofiber/fiber). It includes basic user
authentication and MongoDB integration.
Users now store a `name` field and belong to role groups for authorization.
Admins can manage role groups with dedicated CRUD endpoints.

## Running locally

```bash
go run main.go
```

Create an `.env` file (see `env` for an example) containing your database
credentials.

Set `APP_ENV=production` in environments where you want Cloudflare DNS records
to be created automatically. The default (`development`) skips Cloudflare
provisioning for unit subdomains.

## Cloudflare DNS for unit subdomains

When super admins create or update a unit, the API will provision a Cloudflare DNS record for that unit's subdomain. Configure these environment variables so DNS can be created:

- `CLOUDFLARE_ZONE_ID` – Cloudflare zone containing your root domain.
- `CLOUDFLARE_API_TOKEN` – token with DNS edit permissions for the zone.
- `CLOUDFLARE_ROOT_DOMAIN` – apex domain (e.g., `example.com`), without protocol.
- `CLOUDFLARE_CNAME_TARGET` – host that unit subdomains should CNAME to (e.g., `app.example.com`).
- `CLOUDFLARE_PROXIED` – `true` to proxy through Cloudflare (default), `false` for DNS-only.
- `CLOUDFLARE_TTL` – TTL in seconds; `1` uses Cloudflare's automatic TTL.

## Postman Collection

To quickly explore the API you can import
`postman/go-fiber-template.postman_collection.json` into Postman. The collection
assumes two variables:

- `baseUrl` – base address of your running server, e.g. `http://localhost:4000`
- `token` – JWT obtained from the `Login` request

The collection contains examples for logging in, retrieving and updating the
current user, and managing users as an admin.
