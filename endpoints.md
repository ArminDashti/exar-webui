# API Endpoints

REST API reference for the Daily Expenses backend. All routes are prefixed with `/api`.

**Base URL:** `http://<host>:8080/api` (default port `8080`, configurable via `ADDR`)

**Content type:** `application/json` for request and response bodies.

**Errors:** Failed requests return JSON `{"error": "<message>"}` with an appropriate HTTP status code.

There is no authentication. Anyone who can reach the server can call these endpoints.

Dates in the API and database are Gregorian `YYYY-MM-DD`. The web UI displays Jalali dates.

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

List all shops, sorted by name.

**Response `200`**

```json
[
  { "id": 1, "name": "Grocery Store" }
]
```

### `POST /shops`

Create a new shop.

**Request body**

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `name` | string | yes | Trimmed; must be non-empty |

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

**Request body**

```json
{ "name": "Corner Market" }
```

**Response `200`**

```json
{ "id": 3, "name": "Corner Market" }
```

**Errors**

| Status | Condition |
|--------|-----------|
| `400` | Invalid `id` or empty `name` |
| `404` | Shop not found |
| `409` | Shop name already exists |

### `DELETE /shops/:id`

Delete a shop by ID.

**Response `204`** — no body.

**Errors**

| Status | Condition |
|--------|-----------|
| `400` | Invalid `id` |
| `404` | Shop not found |
| `409` | Shop is referenced by one or more invoices |

---

## Items

Catalog of item names used for expense line-item autocomplete. Invoice lines store the name as text (no foreign key).

### `GET /items`

List all items, sorted by name. Optional search with `q` (minimum 3 characters; shorter queries return `[]`).

**Query parameters**

| Parameter | Type | Description |
|-----------|------|-------------|
| `q` | string | Case-insensitive substring search; ignored (empty result) if shorter than 3 characters |

**Response `200`**

```json
[
  { "id": 1, "name": "Milk" }
]
```

### `POST /items`

Create a catalog item.

**Request body**

```json
{ "name": "Milk" }
```

**Response `201`**

```json
{ "id": 1, "name": "Milk" }
```

**Errors**

| Status | Condition |
|--------|-----------|
| `400` | Missing or empty `name` |
| `409` | Item name already exists |

### `PUT /items/:id`

Rename a catalog item.

**Request body**

```json
{ "name": "Organic Milk" }
```

**Response `200`**

```json
{ "id": 1, "name": "Organic Milk" }
```

**Errors**

| Status | Condition |
|--------|-----------|
| `400` | Invalid `id` or empty `name` |
| `404` | Item not found |
| `409` | Item name already exists |

### `DELETE /items/:id`

Delete a catalog item. Always allowed (invoices keep historical names).

**Response `204`** — no body.

**Errors**

| Status | Condition |
|--------|-----------|
| `400` | Invalid `id` |
| `404` | Item not found |

---

## Stats

### `GET /stats`

Aggregated spend totals. `by_person` uses share-weighted amounts (`invoice total × share`).

**Query parameters** (all optional, Gregorian)

| Parameter | Type | Description |
|-----------|------|-------------|
| `from_date` | string | Include invoices on or after (`YYYY-MM-DD`) |
| `to_date` | string | Include invoices on or before (`YYYY-MM-DD`) |

**Response `200`**

```json
{
  "total": 100.0,
  "by_person": [
    { "person_id": 1, "person_name": "Armin", "total": 60.0 },
    { "person_id": 2, "person_name": "Ramin", "total": 40.0 }
  ],
  "by_shop": [
    { "shop_id": 1, "shop_name": "Grocery Store", "total": 100.0 }
  ]
}
```

---

## Invoices

### `GET /invoices`

List invoices with line items and shares, newest first.

**Query parameters** (all optional)

| Parameter | Type | Description |
|-----------|------|-------------|
| `person_id` | integer | Filter by person who paid (`1` or `2`) |
| `from_date` | string | Include invoices on or after this date (`YYYY-MM-DD`) |
| `to_date` | string | Include invoices on or before this date (`YYYY-MM-DD`) |

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
    "total": 9.0,
    "items": [
      {
        "id": 21,
        "invoice_id": 10,
        "description": "Milk",
        "amount": 4.50,
        "quantity": 1
      },
      {
        "id": 22,
        "invoice_id": 10,
        "description": "Bread",
        "amount": 4.50,
        "quantity": 1
      }
    ],
    "shares": [
      { "person_id": 1, "person_name": "Armin", "share": 0.3 },
      { "person_id": 2, "person_name": "Ramin", "share": 0.7 }
    ]
  }
]
```

### `GET /invoices/:id`

Get a single invoice with line items and shares.

**Response `200`** — same shape as one element in the list response above.

**Errors**

| Status | Condition |
|--------|-----------|
| `400` | Invalid `id` |
| `404` | Invoice not found |

### `POST /invoices`

Create an invoice, line items, and per-person shares in one transaction. The server computes `total` as the sum of `amount` for each item. Quantity is always stored as `1`. Shares must include every person and sum to `1` (within `0.001`).

**Request body**

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `person_id` | integer | yes | Who paid; must be `1` or `2` |
| `shop_id` | integer | yes | Must reference an existing shop |
| `date` | string | yes | Gregorian `YYYY-MM-DD` |
| `items` | array | yes | At least one item |
| `items[].description` | string | yes | Item name (trimmed) |
| `items[].amount` | number | yes | Line amount |
| `shares` | array | yes | One entry per person |
| `shares[].person_id` | integer | yes | Must be `1` or `2` |
| `shares[].share` | number | yes | Fraction of the expense (`>= 0`) |

```json
{
  "person_id": 1,
  "shop_id": 2,
  "date": "2026-06-15",
  "items": [
    { "description": "Milk", "amount": 4.50 }
  ],
  "shares": [
    { "person_id": 1, "share": 0.3 },
    { "person_id": 2, "share": 0.7 }
  ]
}
```

**Response `201`** — full invoice object (same shape as `GET /invoices/:id`). If the created invoice cannot be loaded, the response is `{"id": <new_id>}`.

**Errors**

| Status | Condition |
|--------|-----------|
| `400` | Validation failure, invalid `person_id`/`shop_id`, or shares do not sum to 1 |

### `PUT /invoices/:id`

Replace payer, shop, date, line items, and shares. Same body and validation as `POST /invoices`.

**Response `200`** — full updated invoice.

**Errors**

| Status | Condition |
|--------|-----------|
| `400` | Validation failure |
| `404` | Invoice not found |

### `DELETE /invoices/:id`

Delete an invoice, its line items, and its shares.

**Response `204`** — no body.

**Errors**

| Status | Condition |
|--------|-----------|
| `400` | Invalid `id` |
| `404` | Invoice not found |

---

## Static files and unknown routes

When a static frontend is deployed (`STATIC_DIR`), the server also serves:

- `GET /assets/*` — built frontend assets
- `GET /*` (non-API) — `index.html` for client-side routing

Unknown paths under `/api` return `404` with `{"error": "not found"}`.
