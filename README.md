# Go Fiber Template

This project provides a basic setup for a Go Fiber REST API. A Postman collection is generated from the route definitions.

## Updating the Postman collection

Whenever the API routes change, run:

```bash
make postman
```

This executes `scripts/generate_postman.py` which parses `routes/routes.go` and updates `postman_collection.json` for import into Postman.
