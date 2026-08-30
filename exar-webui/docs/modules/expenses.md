# Expenses module

Flat spending records in `expenses` + `expense_shares`.

- Each row: `person_id` (payer), `shop_id`, `item_id`, `date` (Gregorian), `amount` (whole number).
- Item display name comes from `items` via FK; renaming a catalog item updates all expenses.
- Shares: fraction each person must cover (`1` = 100%); must sum to 1 per expense.
- `POST /expenses` accepts a batch header + lines, upserts item names, and inserts independent rows (no parent invoice).
- `GET /expenses/check-duplicate` warns when the same item already exists on the same date (soft warning only).
- Existing invoice tables are migrated once into expenses, then dropped. Legacy `expenses.name` text is migrated to `item_id` with backup/restore.
