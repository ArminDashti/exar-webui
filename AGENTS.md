## Learned User Preferences

- Prefer backend orchestration for expense save: upsert catalog item names (and ideally resolve/create shops) inside `POST /expenses`, not frontend chaining `POST /items` then `POST /expenses`.
- Shares are per line item, not the whole expense; changing one person’s share should auto-adjust the other so they sum to `1` (`1` = 100%).
- There is no invoice entity — spending is flat expense line items only.
- Prefer a date picker over a plain text date field when entering expenses.
- Format money amounts as whole numbers (`0`), not `0.00` or `0.0`.
- Shops should work like items: character search/autocomplete and create on save when missing.

## Learned Workspace Facts

- exar-web is a shared expense app for two people (Armin `person_id` 1, Ramin `person_id` 2) with per-item share splits.
- Production hostname is `exar.xaigrok.ir`; remote Docker deploy uses `ssh t3` and Docker network `exar-net`.
- `POST /api/items` maintains the item-name catalog for autocomplete; expense rows store the name as text with no FK to `items`.
- `POST /api/expenses` is the real spend write (`expenses` + `expense_shares`); catalog upsert is currently frontend-orchestrated and is preferred on the backend.
- Expense UI is nested under Expenses with add and list child pages; stats settle amounts owed between Armin and Ramin.
