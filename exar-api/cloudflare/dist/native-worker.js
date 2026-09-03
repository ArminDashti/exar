export default {
  async fetch(request, env) {
    if (request.method === "OPTIONS") {
      return cors(new Response(null, { status: 204 }), request, env);
    }
    try {
      const res = await handle(request, env);
      return cors(res, request, env);
    } catch (err) {
      return cors(json({ error: err instanceof Error ? err.message : String(err) }, 500), request, env);
    }
  },
};

function json(body, status = 200) {
  return Response.json(body, { status });
}

function cors(res, request, env) {
  const origins = (env.CORS_ORIGINS || "")
    .split(",")
    .map((s) => s.trim())
    .filter(Boolean);
  const origin = request.headers.get("Origin") || "";
  let allow = origins[0] || "*";
  if (origin && (origins.includes(origin) || origin.includes("exar.armindashti.workers.dev") || origins.includes("*"))) {
    allow = origin;
  }
  const headers = new Headers(res.headers);
  headers.set("Access-Control-Allow-Origin", allow);
  headers.set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS");
  headers.set("Access-Control-Allow-Headers", "Authorization, Content-Type");
  return new Response(res.body, { status: res.status, statusText: res.statusText, headers });
}

function unique(err) {
  return /UNIQUE|constraint failed/i.test(err instanceof Error ? err.message : String(err));
}

function validateShares(shares) {
  const seen = new Set();
  let sum = 0;
  for (const s of shares) {
    if (s.person_id !== 1 && s.person_id !== 2) return "share person_id must be 1 or 2";
    if (s.share < 0) return "share must be >= 0";
    if (seen.has(s.person_id)) return "duplicate person in shares";
    seen.add(s.person_id);
    sum += s.share;
  }
  for (const id of [1, 2]) if (!seen.has(id)) return "shares must include every person";
  if (Math.abs(sum - 1) > 0.001) return "shares must sum to 1";
  return null;
}

function validateWholeAmount(amount) {
  if (amount < 0) return "amount must be >= 0";
  if (amount !== Math.trunc(amount)) return "amount must be a whole number";
  return null;
}

const gdm = [0, 31, 59, 90, 120, 151, 181, 212, 243, 273, 304, 334];
const jMonthLen = [31, 31, 31, 31, 31, 31, 30, 30, 30, 30, 30, 29];
function toJalali(gy, gm, gd) {
  const gy2 = gy - 1600, gm2 = gm - 1, gd2 = gd - 1;
  let gDayNo = 365 * gy2 + Math.floor((gy2 + 3) / 4) - Math.floor((gy2 + 99) / 100) + Math.floor((gy2 + 399) / 400);
  gDayNo += gdm[gm2] + gd2;
  if (gm > 2 && ((gy % 4 === 0 && gy % 100 !== 0) || gy % 400 === 0)) gDayNo++;
  let jDayNo = gDayNo - 79;
  const jNp = Math.floor(jDayNo / 12053);
  jDayNo %= 12053;
  let jy = 979 + 33 * jNp + 4 * Math.floor(jDayNo / 1461);
  jDayNo %= 1461;
  if (jDayNo >= 366) {
    jy += Math.floor((jDayNo - 1) / 365);
    jDayNo = (jDayNo - 1) % 365;
  }
  let i = 0;
  for (; i < 11 && jDayNo >= jMonthLen[i]; i++) jDayNo -= jMonthLen[i];
  return [jy, i + 1, jDayNo + 1];
}
function monthKey(date) {
  const m = /^(\d+)-(\d+)-(\d+)$/.exec(date.trim());
  if (!m) return null;
  const [jy, jm] = toJalali(+m[1], +m[2], +m[3]);
  return `${String(jy).padStart(4, "0")}/${String(jm).padStart(2, "0")}`;
}

async function upsertItem(db, name) {
  const existing = await db.prepare("SELECT id FROM items WHERE name = ? COLLATE NOCASE").bind(name).first();
  if (existing) return existing.id;
  try {
    const result = await db.prepare("INSERT INTO items (name) VALUES (?)").bind(name).run();
    return Number(result.meta.last_row_id);
  } catch (err) {
    if (!unique(err)) throw err;
    const again = await db.prepare("SELECT id FROM items WHERE name = ? COLLATE NOCASE").bind(name).first();
    if (!again) throw err;
    return again.id;
  }
}

async function insertShares(db, expenseId, shares) {
  const stmts = shares.map((s) =>
    db.prepare("INSERT INTO expense_shares (expense_id, person_id, share) VALUES (?, ?, ?)").bind(expenseId, s.person_id, s.share),
  );
  if (stmts.length) await db.batch(stmts);
}

async function loadShares(db, ids) {
  const out = new Map();
  if (!ids.length) return out;
  const chunkSize = 80;
  for (let offset = 0; offset < ids.length; offset += chunkSize) {
    const chunk = ids.slice(offset, offset + chunkSize);
    const ph = chunk.map(() => "?").join(",");
    const { results } = await db
      .prepare(
        `SELECT sh.expense_id, sh.person_id, p.name AS person_name, sh.share
       FROM expense_shares sh JOIN persons p ON p.id = sh.person_id
       WHERE sh.expense_id IN (${ph}) ORDER BY sh.person_id`,
      )
      .bind(...chunk)
      .all();
    for (const row of results || []) {
      const list = out.get(row.expense_id) || [];
      list.push({
        person_id: row.person_id,
        person_name: row.person_name,
        share: row.share,
      });
      out.set(row.expense_id, list);
    }
  }
  return out;
}

async function fetchExpense(db, id) {
  const row = await db
    .prepare(
      `SELECT e.id, e.person_id, p.name AS person_name, e.shop_id, s.name AS shop_name,
              e.item_id, i.name AS name, e.date, e.amount
       FROM expenses e
       JOIN persons p ON p.id = e.person_id
       JOIN shops s ON s.id = e.shop_id
       JOIN items i ON i.id = e.item_id
       WHERE e.id = ?`,
    )
    .bind(id)
    .first();
  if (!row) return null;
  const shares = await loadShares(db, [id]);
  return { ...row, shares: shares.get(id) || [] };
}

async function handle(request, env) {
  const url = new URL(request.url);
  const path = url.pathname.replace(/\/$/, "") || "/";
  const db = env.DB;
  const method = request.method;

  if (method === "GET" && path === "/api/health") return json({ status: "ok" });

  if (method === "GET" && path === "/api/persons") {
    const rows = await db.prepare("SELECT id, name FROM persons ORDER BY id").all();
    return json(rows.results || []);
  }

  if (path === "/api/shops") {
    if (method === "GET") {
      const q = (url.searchParams.get("q") || "").trim();
      if (q && [...q].length < 3) return json([]);
      const rows = q
        ? await db.prepare("SELECT id, name FROM shops WHERE name LIKE ? COLLATE NOCASE ORDER BY name").bind(`%${q}%`).all()
        : await db.prepare("SELECT id, name FROM shops ORDER BY name").all();
      return json(rows.results || []);
    }
    if (method === "POST") {
      const name = ((await request.json()).name || "").trim();
      if (!name) return json({ error: "name is required" }, 400);
      try {
        const result = await db.prepare("INSERT INTO shops (name) VALUES (?)").bind(name).run();
        return json({ id: Number(result.meta.last_row_id), name }, 201);
      } catch (err) {
        return unique(err) ? json({ error: "shop already exists" }, 409) : json({ error: "failed to create shop" }, 500);
      }
    }
  }

  let m = path.match(/^\/api\/shops\/(\d+)$/);
  if (m) {
    const id = Number(m[1]);
    if (method === "PUT") {
      const name = ((await request.json()).name || "").trim();
      if (!name) return json({ error: "name is required" }, 400);
      try {
        const result = await db.prepare("UPDATE shops SET name = ? WHERE id = ?").bind(name, id).run();
        if (!result.meta.changes) return json({ error: "shop not found" }, 404);
        return json({ id, name });
      } catch (err) {
        return unique(err) ? json({ error: "shop already exists" }, 409) : json({ error: "failed to update shop" }, 500);
      }
    }
    if (method === "DELETE") {
      const inUse = await db.prepare("SELECT COUNT(*) AS c FROM expenses WHERE shop_id = ?").bind(id).first();
      if ((inUse?.c || 0) > 0) return json({ error: "shop is used by expenses" }, 409);
      const result = await db.prepare("DELETE FROM shops WHERE id = ?").bind(id).run();
      if (!result.meta.changes) return json({ error: "shop not found" }, 404);
      return new Response(null, { status: 204 });
    }
  }

  if (path === "/api/items") {
    if (method === "GET") {
      const q = (url.searchParams.get("q") || "").trim();
      if (q && [...q].length < 3) return json([]);
      const rows = q
        ? await db.prepare("SELECT id, name FROM items WHERE name LIKE ? COLLATE NOCASE ORDER BY name").bind(`%${q}%`).all()
        : await db.prepare("SELECT id, name FROM items ORDER BY name").all();
      return json(rows.results || []);
    }
    if (method === "POST") {
      const name = ((await request.json()).name || "").trim();
      if (!name) return json({ error: "name is required" }, 400);
      try {
        const result = await db.prepare("INSERT INTO items (name) VALUES (?)").bind(name).run();
        return json({ id: Number(result.meta.last_row_id), name }, 201);
      } catch (err) {
        return unique(err) ? json({ error: "item already exists" }, 409) : json({ error: "failed to create item" }, 500);
      }
    }
  }

  m = path.match(/^\/api\/items\/(\d+)$/);
  if (m) {
    const id = Number(m[1]);
    if (method === "PUT") {
      const name = ((await request.json()).name || "").trim();
      if (!name) return json({ error: "name is required" }, 400);
      try {
        const result = await db.prepare("UPDATE items SET name = ? WHERE id = ?").bind(name, id).run();
        if (!result.meta.changes) return json({ error: "item not found" }, 404);
        return json({ id, name });
      } catch (err) {
        return unique(err) ? json({ error: "item already exists" }, 409) : json({ error: "failed to update item" }, 500);
      }
    }
    if (method === "DELETE") {
      const inUse = await db.prepare("SELECT COUNT(*) AS c FROM expenses WHERE item_id = ?").bind(id).first();
      if ((inUse?.c || 0) > 0) return json({ error: "item is used by expenses" }, 409);
      const result = await db.prepare("DELETE FROM items WHERE id = ?").bind(id).run();
      if (!result.meta.changes) return json({ error: "item not found" }, 404);
      return new Response(null, { status: 204 });
    }
  }

  if (method === "GET" && path === "/api/stats") {
    const where = ["1=1"];
    const args = [];
    const from = url.searchParams.get("from_date");
    const to = url.searchParams.get("to_date");
    if (from) {
      where.push("e.date >= ?");
      args.push(from);
    }
    if (to) {
      where.push("e.date <= ?");
      args.push(to);
    }
    const rows = await db
      .prepare(
        `SELECT e.date, e.person_id, e.amount,
          COALESCE((SELECT share FROM expense_shares WHERE expense_id = e.id AND person_id = 1), 0) AS armin_share,
          COALESCE((SELECT share FROM expense_shares WHERE expense_id = e.id AND person_id = 2), 0) AS ramin_share
         FROM expenses e WHERE ${where.join(" AND ")}`,
      )
      .bind(...args)
      .all();
    const byMonth = new Map();
    for (const row of rows.results || []) {
      const month = monthKey(row.date);
      if (!month) continue;
      let s = byMonth.get(month);
      if (!s) {
        s = { month, armin: 0, ramin: 0, total: 0, armin_share: 0, ramin_share: 0 };
        byMonth.set(month, s);
      }
      s.total += row.amount;
      if (row.person_id === 1) s.armin += row.amount;
      else if (row.person_id === 2) s.ramin += row.amount;
      s.armin_share += row.amount * row.armin_share;
      s.ramin_share += row.amount * row.ramin_share;
    }
    return json({ by_month: [...byMonth.values()].sort((a, b) => (a.month < b.month ? 1 : -1)) });
  }

  if (method === "GET" && path === "/api/expenses/check-duplicate") {
    const date = (url.searchParams.get("date") || "").trim();
    if (!date) return json({ error: "date is required" }, 400);
    const excludeID = (url.searchParams.get("exclude_id") || "").trim();
    const itemIDStr = (url.searchParams.get("item_id") || "").trim();
    const name = (url.searchParams.get("name") || "").trim();
    let count = 0;
    if (itemIDStr) {
      const itemID = Number(itemIDStr);
      if (!Number.isFinite(itemID)) return json({ error: "invalid item_id" }, 400);
      const row = excludeID
        ? await db.prepare("SELECT COUNT(*) AS c FROM expenses WHERE item_id = ? AND date = ? AND id != ?").bind(itemID, date, excludeID).first()
        : await db.prepare("SELECT COUNT(*) AS c FROM expenses WHERE item_id = ? AND date = ?").bind(itemID, date).first();
      count = row?.c || 0;
    } else if (name) {
      const row = excludeID
        ? await db
            .prepare(
              `SELECT COUNT(*) AS c FROM expenses e JOIN items i ON i.id = e.item_id
               WHERE i.name = ? COLLATE NOCASE AND e.date = ? AND e.id != ?`,
            )
            .bind(name, date, excludeID)
            .first()
        : await db
            .prepare(
              `SELECT COUNT(*) AS c FROM expenses e JOIN items i ON i.id = e.item_id
               WHERE i.name = ? COLLATE NOCASE AND e.date = ?`,
            )
            .bind(name, date)
            .first();
      count = row?.c || 0;
    } else return json({ error: "item_id or name is required" }, 400);
    return json({ exists: count > 0, count });
  }

  if (path === "/api/expenses") {
    if (method === "GET") {
      const where = ["1=1"];
      const args = [];
      const personID = url.searchParams.get("person_id");
      const from = url.searchParams.get("from_date");
      const to = url.searchParams.get("to_date");
      if (personID) {
        where.push("e.person_id = ?");
        args.push(personID);
      }
      if (from) {
        where.push("e.date >= ?");
        args.push(from);
      }
      if (to) {
        where.push("e.date <= ?");
        args.push(to);
      }
      const rows = await db
        .prepare(
          `SELECT e.id, e.person_id, p.name AS person_name, e.shop_id, s.name AS shop_name,
                  e.item_id, i.name AS name, e.date, e.amount
           FROM expenses e
           JOIN persons p ON p.id = e.person_id
           JOIN shops s ON s.id = e.shop_id
           JOIN items i ON i.id = e.item_id
           WHERE ${where.join(" AND ")}
           ORDER BY e.date DESC, e.id DESC`,
        )
        .bind(...args)
        .all();
      const expenses = (rows.results || []).map((row) => ({ ...row, shares: [] }));
      const shares = await loadShares(
        db,
        expenses.map((e) => e.id),
      );
      for (const e of expenses) e.shares = shares.get(e.id) || [];
      return json(expenses);
    }
    if (method === "POST") {
      const body = await request.json();
      const personId = Number(body.person_id);
      const shopId = Number(body.shop_id);
      const date = (body.date || "").trim();
      const items = body.items || [];
      if (personId !== 1 && personId !== 2) return json({ error: "person_id must be 1 or 2" }, 400);
      if (!date) return json({ error: "date is required" }, 400);
      if (!items.length) return json({ error: "items is required" }, 400);
      for (const item of items) {
        const se = validateShares(item.shares || []);
        if (se) return json({ error: se }, 400);
        if (!(item.name || "").trim()) return json({ error: "item name is required" }, 400);
        const ae = validateWholeAmount(Number(item.amount));
        if (ae) return json({ error: ae }, 400);
      }
      if (!(await db.prepare("SELECT id FROM shops WHERE id = ?").bind(shopId).first())) {
        return json({ error: "invalid shop_id" }, 400);
      }
      const createdIDs = [];
      for (const item of items) {
        const name = (item.name || "").trim();
        const itemId = await upsertItem(db, name);
        const result = await db
          .prepare("INSERT INTO expenses (person_id, shop_id, item_id, date, amount) VALUES (?, ?, ?, ?, ?)")
          .bind(personId, shopId, itemId, date, Number(item.amount))
          .run();
        const expenseId = Number(result.meta.last_row_id);
        await insertShares(db, expenseId, item.shares || []);
        createdIDs.push(expenseId);
      }
      const created = [];
      for (const id of createdIDs) {
        const e = await fetchExpense(db, id);
        if (e) created.push(e);
      }
      return json(created, 201);
    }
  }

  m = path.match(/^\/api\/expenses\/(\d+)$/);
  if (m) {
    const id = Number(m[1]);
    if (method === "GET") {
      const e = await fetchExpense(db, id);
      return e ? json(e) : json({ error: "expense not found" }, 404);
    }
    if (method === "PUT") {
      const body = await request.json();
      const personId = Number(body.person_id);
      const shopId = Number(body.shop_id);
      const date = (body.date || "").trim();
      const name = (body.name || "").trim();
      const amount = Number(body.amount);
      const shares = body.shares || [];
      if (personId !== 1 && personId !== 2) return json({ error: "person_id must be 1 or 2" }, 400);
      const se = validateShares(shares);
      if (se) return json({ error: se }, 400);
      const ae = validateWholeAmount(amount);
      if (ae) return json({ error: ae }, 400);
      if (!name) return json({ error: "name is required" }, 400);
      if (!(await db.prepare("SELECT id FROM shops WHERE id = ?").bind(shopId).first())) {
        return json({ error: "invalid shop_id" }, 400);
      }
      if (!(await db.prepare("SELECT id FROM expenses WHERE id = ?").bind(id).first())) {
        return json({ error: "expense not found" }, 404);
      }
      const itemId = await upsertItem(db, name);
      await db
        .prepare("UPDATE expenses SET person_id = ?, shop_id = ?, item_id = ?, date = ?, amount = ? WHERE id = ?")
        .bind(personId, shopId, itemId, date, amount, id)
        .run();
      await db.prepare("DELETE FROM expense_shares WHERE expense_id = ?").bind(id).run();
      await insertShares(db, id, shares);
      return json((await fetchExpense(db, id)) || { id });
    }
    if (method === "DELETE") {
      await db.prepare("DELETE FROM expense_shares WHERE expense_id = ?").bind(id).run();
      const result = await db.prepare("DELETE FROM expenses WHERE id = ?").bind(id).run();
      if (!result.meta.changes) return json({ error: "expense not found" }, 404);
      return new Response(null, { status: 204 });
    }
  }

  return json({ error: "not found" }, 404);
}
