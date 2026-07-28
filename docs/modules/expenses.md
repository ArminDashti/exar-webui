# Expenses module

Flat spending records in `expenses` + `expense_shares`.

- Each row: `person_id` (payer), `shop_id`, `date` (Gregorian), `name`, `amount`.
- Shares: fraction each person must cover (`1` = 100%); must sum to 1 per expense.
- `POST /expenses` accepts a batch header + lines and inserts independent rows (no parent invoice).
- Existing invoice tables are migrated once into expenses, then dropped.
