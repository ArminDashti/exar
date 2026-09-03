import json
from pathlib import Path

dump = json.loads(Path(r"C:\Users\armin\GitHub\exar\exar-api\cloudflare\dist\expenses-dump.json").read_text(encoding="utf-8"))
out = Path(r"C:\Users\armin\GitHub\exar\exar-api\cloudflare\dist\d1-import-batches")
out.mkdir(exist_ok=True)

def esc(v):
    if v is None:
        return "NULL"
    if isinstance(v, (int, float)) and not isinstance(v, bool):
        return str(v)
    s = str(v).replace("'", "''")
    return f"'{s}'"

# Clear and reseed in dependency order
stmts = [
    "DELETE FROM expense_shares;",
    "DELETE FROM expenses;",
    "DELETE FROM items;",
    "DELETE FROM shops;",
    "DELETE FROM persons;",
]

for p in dump["data"]["persons"]:
    stmts.append(f"INSERT INTO persons (id, name) VALUES ({esc(p['id'])}, {esc(p['name'])});")

for s in dump["data"]["shops"]:
    stmts.append(f"INSERT INTO shops (id, name) VALUES ({esc(s['id'])}, {esc(s['name'])});")

for it in dump["data"]["items"]:
    stmts.append(f"INSERT INTO items (id, name) VALUES ({esc(it['id'])}, {esc(it['name'])});")

for e in dump["data"]["expenses"]:
    stmts.append(
        "INSERT INTO expenses (id, person_id, shop_id, item_id, date, amount) VALUES ("
        f"{esc(e['id'])}, {esc(e['person_id'])}, {esc(e['shop_id'])}, {esc(e['item_id'])}, {esc(e['date'])}, {esc(e['amount'])});"
    )

for sh in dump["data"]["expense_shares"]:
    stmts.append(
        "INSERT INTO expense_shares (expense_id, person_id, share) VALUES ("
        f"{esc(sh['expense_id'])}, {esc(sh['person_id'])}, {esc(sh['share'])});"
    )

# Skip sqlite_sequence — D1 manages AUTOINCREMENT separately; explicit IDs are inserted above.
# batch into files of ~40 statements for D1
batch_size = 40
batches = []
for i in range(0, len(stmts), batch_size):
    batch = stmts[i : i + batch_size]
    batches.append(batch)
    (out / f"batch_{i // batch_size:03d}.sql").write_text("\n".join(batch), encoding="utf-8")

print("statements", len(stmts), "batches", len(batches))
Path(r"C:\Users\armin\GitHub\exar\exar-api\cloudflare\dist\d1-import-meta.json").write_text(
    json.dumps({"statements": len(stmts), "batches": len(batches)}, indent=2), encoding="utf-8"
)
