# API endpoints

See [../endpoints.md](../endpoints.md) for the full REST reference.

Summary:

- `GET /api/persons`
- `GET|POST /api/shops`, `PUT|DELETE /api/shops/:id` (`GET` supports `?q=`)
- `GET|POST /api/items`, `PUT|DELETE /api/items/:id` (`GET` supports `?q=`)
- `GET /api/stats` → `{ by_month: [...] }`
- `GET|POST /api/expenses`, `GET|PUT|DELETE /api/expenses/:id`
