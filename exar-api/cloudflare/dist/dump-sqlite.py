import sqlite3
import json
from pathlib import Path

db_path = Path(r"C:\Users\armin\GitHub\exar\exar-api\cloudflare\dist\expenses.db")
out_path = Path(r"C:\Users\armin\GitHub\exar\exar-api\cloudflare\dist\expenses-dump.json")
con = sqlite3.connect(db_path)
con.row_factory = sqlite3.Row
tables = [r[0] for r in con.execute("SELECT name FROM sqlite_master WHERE type='table' ORDER BY name")]
dump = {"tables": tables, "data": {}}
for t in tables:
    rows = [dict(r) for r in con.execute(f"SELECT * FROM {t}")]
    dump["data"][t] = rows
    print(t, len(rows))
out_path.write_text(json.dumps(dump, ensure_ascii=False, indent=2), encoding="utf-8")
print("wrote", out_path, "bytes", out_path.stat().st_size)
