# Directory tree

```
cmd/server/main.go          # Gin HTTP entrypoint, /api routes, optional static SPA
internal/database/          # SQLite open + schema migrate (expenses, expense_shares)
internal/handlers/          # REST handlers (persons, shops, items, expenses, stats)
internal/jalali/            # Gregorian → Jalali month helpers for stats
internal/models/            # JSON/API structs
src/                        # Vue 3 + Vite frontend
src/api.js                  # Fetch client for /api
src/App.vue                 # Shell, nav, footer GitHub link
src/router.js               # Nested /expenses/add|/list routes
src/components/JalaliDateInput.vue  # Jalali calendar date picker
src/views/ExpensesLayout.vue        # Expenses Add|List tabs
src/views/ExpenseAddView.vue        # Multi-row add form
src/views/ExpenseListView.vue       # Expense table grid + inline edit
src/views/StatsView.vue             # Monthly stats grid
src/views/ShopsView.vue / ItemsView.vue  # Catalog CRUD
src/utils/dates.js / money.js       # Jalali helpers; integer money format
endpoints.md                # API reference (also mirrored in docs/endpoints.md)
docs/                       # Project documentation (this folder)
```
