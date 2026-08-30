# Database schema

PostgreSQL database used by the API (`api/internal/database/database.go`).

## persons

Seeded users who can record expenses.

| Column | Type | Constraints |
|--------|------|-------------|
| id | INTEGER | PRIMARY KEY |
| name | TEXT | NOT NULL |

**Seed data:** `1` = armin, `2` = ramin

## shops

Stores where expenses were made.

| Column | Type | Constraints |
|--------|------|-------------|
| id | SERIAL | PRIMARY KEY |
| name | TEXT | NOT NULL, UNIQUE |

## products

Item names (e.g. Milk) that can be assigned to expenses.

| Column | Type | Constraints |
|--------|------|-------------|
| id | SERIAL | PRIMARY KEY |
| name | TEXT | NOT NULL, UNIQUE |

## items

Expense records linking a shop, product, person, amount, and date.

| Column | Type | Constraints |
|--------|------|-------------|
| id | SERIAL | PRIMARY KEY |
| shop_id | INTEGER | NOT NULL, FK → shops(id) |
| product_id | INTEGER | FK → products(id) |
| person_id | INTEGER | FK → persons(id) |
| amount | DOUBLE PRECISION | NOT NULL |
| date | TEXT | NOT NULL (Gregorian `YYYY-MM-DD`) |

## Indexes

| Index | Table | Column |
|-------|-------|--------|
| idx_items_date | items | date |
| idx_items_shop | items | shop_id |
| idx_items_person | items | person_id |
| idx_items_product | items | product_id |

## Relationships

| From | To | Description |
|------|----|-------------|
| items.shop_id | shops.id | Shop for the expense |
| items.product_id | products.id | Product/item name |
| items.person_id | persons.id | Person who made the expense |

## Connection

The API reads `DATABASE_URL` (default `postgres://exar:exar@localhost:5432/exar?sslmode=disable`). Docker Compose sets this to the `postgres` service hostname.
