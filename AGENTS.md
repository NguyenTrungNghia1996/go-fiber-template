API Guidelines for this repository

- Postman sync: When you change API routes or controllers, update `postman/go-fiber-template.postman_collection.json` so it reflects all available endpoints. Include example request bodies for create and update operations, and example query params for list/detail/delete.

- VS Code REST Client sync: Alongside Postman, maintain `rest-client/go-fiber-template.http` for the REST Client extension. Ensure it mirrors all available endpoints with working sample requests:
  - Include base variables (e.g., `@baseUrl`) and an auth login request that captures `adminToken` from the response for reuse.
  - Provide example request bodies for create and update, and example query params for list/detail/delete.
  - Update this file whenever API routes, request/response shapes, or auth change.

- Standard response envelope: Always return JSON in the form `{ "data": any, "message": string, "status": "success"|"error" }`.
  - For list endpoints, `data` must be an object: `{ items: [], page: number, limit: number, total: number }`.
  - Use the helpers in `pkg/response` (`Success`, `Error`, `ParsePageLimit`, `ListData[T]`).

- HTTP method and parameter conventions (apply to all resources):
  - GET list: `GET /<resource>` without `id` returns a paginated list. Support `?page`, `?limit` and optional `?q` for text search.
    - If `page` is missing or equals `0`, return the full dataset (no pagination).
  - GET detail: `GET /<resource>?id=<id>` returns a single record (do not use path params like `/:id`).
  - POST create: `POST /<resource>` with JSON body creates a record.
  - PUT update: `PUT /<resource>` with JSON body must include `id` and fields to update.
  - DELETE: `DELETE /<resource>?id=<id>` deletes a record (do not use path params like `/:id`).

- Search: When supported, `q` is a case-insensitive text search applied to relevant fields of the resource. Keep implementation efficient and safe (escape/quote regex where applicable).

- Security: Never include sensitive fields (e.g., password hashes) in responses. Hash incoming passwords with bcrypt before storage.
