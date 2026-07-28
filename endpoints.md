# API Endpoints

REST API reference for the Daily Expenses backend. All routes are prefixed with `/api`.

**Base URL:** `http://<host>:8080/api` (default port `8080`, configurable via `ADDR`)

**Content type:** `application/json` for request and response bodies.

**Errors:** Failed requests return JSON `{"error": "<message>"}` with an appropriate HTTP status code.

There is no authentication. Anyone who can reach the server can call these endpoints.

Dates in the API and database are Gregorian `YYYY-MM-DD`. The web UI displays Jalali dates.

There is no invoice model. Each spending record is a flat **expense** with its own paid-by, shop, date, name, amount, and per-person shares.

---

## Persons

### `GET /persons`

List the two fixed people in the app.

**Response `200`**

```json
[
  { "id": 1, "name": "Armin" },
  { "id": 2, "name": "Ramin" }
]
```

---

## Shops

### `GET /shops`

List all shops, sorted by name. Optional search with `q` (minimum 3 characters; shorter queries return `[]`).

**Query parameters**

| Parameter | Type | Description |
|-----------|------|-------------|
| `q` | string | Case-insensitive substring search; empty result if shorter than 3 characters |

**Response `200`**

```json
[
  { "id": 1, "name": "Grocery Store" }
]
```

### `POST /shops`

Create a new shop.

**Request body**

```json
{ "name": "Grocery Store" }
```

**Response `201`**

```json
{ "id": 3, "name": "Grocery Store" }
```

**Errors**

| Status | Condition |
|--------|-----------|
| `400` | Missing or empty `name` |
| `409` | Shop name already exists |

### `PUT /shops/:id`

Rename a shop.

**Response `200`** — updated shop.

### `DELETE /shops/:id`

**Response `204`**. Fails with `409` if the shop is used by expenses.

---

## Items

Catalog of item names used for expense name autocomplete. Expenses store the name as text (no foreign key).

### `GET /items`

Optional `q` search (minimum 3 characters).

### `POST /items` / `PUT /items/:id` / `DELETE /items/:id`

Same create/rename/delete behavior as shops (unique names).

---

## Stats

### `GET /stats`

Monthly totals in Jalali months (`YYYY/MM`).

**Query parameters** (optional, Gregorian)

| Parameter | Type | Description |
|-----------|------|-------------|
| `from_date` | string | Include expenses on or after (`YYYY-MM-DD`) |
| `to_date` | string | Include expenses on or before (`YYYY-MM-DD`) |

**Response `200`**

```json
{
  "by_month": [
    {
      "month": "1405/04",
      "armin": 100,
      "ramin": 50,
      "total": 150,
      "armin_share": 90,
      "ramin_share": 60
    }
  ]
}
```

| Field | Meaning |
|-------|---------|
| `armin` / `ramin` | Sum of amounts where that person paid |
| `total` | Sum of all amounts |
| `armin_share` / `ramin_share` | `SUM(amount × share)` obligation |

---

## Expenses

### `GET /expenses`

List flat expenses with shares, newest first.

**Query parameters** (optional): `person_id`, `from_date`, `to_date`.

**Response `200`**

```json
[
  {
    "id": 10,
    "person_id": 1,
    "person_name": "Armin",
    "shop_id": 2,
    "shop_name": "Grocery Store",
    "date": "2026-06-15",
    "name": "Milk",
    "amount": 45,
    "shares": [
      { "person_id": 1, "person_name": "Armin", "share": 0.7 },
      { "person_id": 2, "person_name": "Ramin", "share": 0.3 }
    ]
  }
]
```

### `GET /expenses/:id`

Single expense. `404` if missing.

### `POST /expenses`

Batch-create flat expenses (no parent record). Shared header fields apply to each line.

**Request body**

```json
{
  "person_id": 1,
  "shop_id": 2,
  "date": "2026-06-15",
  "items": [
    {
      "name": "Milk",
      "amount": 45,
      "shares": [
        { "person_id": 1, "share": 0.7 },
        { "person_id": 2, "share": 0.3 }
      ]
    }
  ]
}
```

Each item’s shares must include persons `1` and `2` and sum to `1` (within `0.001`).

**Response `201`** — array of created expenses.

### `PUT /expenses/:id`

Update one expense.

**Request body**

```json
{
  "person_id": 1,
  "shop_id": 2,
  "date": "2026-06-15",
  "name": "Milk",
  "amount": 45,
  "shares": [
    { "person_id": 1, "share": 0.5 },
    { "person_id": 2, "share": 0.5 }
  ]
}
```

### `DELETE /expenses/:id`

**Response `204`**.

---

## Static files and unknown routes

When a static frontend is deployed (`STATIC_DIR`), the server also serves:

- `GET /assets/*` — built frontend assets
- `GET /*` (non-API) — `index.html` for client-side routing

Unknown paths under `/api` return `404` with `{"error": "not found"}`.
