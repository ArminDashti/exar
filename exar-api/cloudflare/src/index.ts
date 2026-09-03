import { Hono } from 'hono'
import { cors } from 'hono/cors'
import { monthKeyFromGregorian } from './jalali'
import type {
  Env,
  Expense,
  ExpenseShare,
  ExpenseShareInput,
  Item,
  MonthStats,
  Person,
  Shop,
} from './types'

const app = new Hono<{ Bindings: Env }>()

app.use('*', async (c, next) => {
  const origins = (c.env.CORS_ORIGINS || '')
    .split(',')
    .map((s) => s.trim())
    .filter(Boolean)
  const handler = cors({
    origin: (origin) => {
      if (!origin) return origins[0] || '*'
      if (origins.includes(origin)) return origin
      if (origin.includes('exar.armindashti.workers.dev')) return origin
      if (origins.includes('*')) return origin
      return origins[0] || ''
    },
    allowMethods: ['GET', 'POST', 'PUT', 'PATCH', 'DELETE', 'OPTIONS'],
    allowHeaders: ['Authorization', 'Content-Type'],
  })
  return handler(c, next)
})

function jsonError(error: string, status: number) {
  return Response.json({ error }, { status })
}

function isUniqueError(err: unknown): boolean {
  const msg = err instanceof Error ? err.message : String(err)
  return /UNIQUE|constraint failed/i.test(msg)
}

function validateShares(shares: ExpenseShareInput[]): string | null {
  const seen = new Set<number>()
  let sum = 0
  for (const s of shares) {
    if (s.person_id !== 1 && s.person_id !== 2) return 'share person_id must be 1 or 2'
    if (s.share < 0) return 'share must be >= 0'
    if (seen.has(s.person_id)) return 'duplicate person in shares'
    seen.add(s.person_id)
    sum += s.share
  }
  for (const id of [1, 2]) {
    if (!seen.has(id)) return 'shares must include every person'
  }
  if (Math.abs(sum - 1) > 0.001) return 'shares must sum to 1'
  return null
}

function validateWholeAmount(amount: number): string | null {
  if (amount < 0) return 'amount must be >= 0'
  if (amount !== Math.trunc(amount)) return 'amount must be a whole number'
  return null
}

async function upsertItem(db: D1Database, name: string): Promise<number> {
  const existing = await db
    .prepare(`SELECT id FROM items WHERE name = ? COLLATE NOCASE`)
    .bind(name)
    .first<{ id: number }>()
  if (existing) return existing.id

  try {
    const result = await db.prepare(`INSERT INTO items (name) VALUES (?)`).bind(name).run()
    const id = result.meta.last_row_id
    if (!id) throw new Error('failed to resolve item')
    return Number(id)
  } catch (err) {
    if (!isUniqueError(err)) throw err
    const again = await db
      .prepare(`SELECT id FROM items WHERE name = ? COLLATE NOCASE`)
      .bind(name)
      .first<{ id: number }>()
    if (!again) throw err
    return again.id
  }
}

async function insertExpenseShares(
  db: D1Database,
  expenseId: number,
  shares: ExpenseShareInput[],
): Promise<void> {
  const stmts = shares.map((s) =>
    db
      .prepare(`INSERT INTO expense_shares (expense_id, person_id, share) VALUES (?, ?, ?)`)
      .bind(expenseId, s.person_id, s.share),
  )
  if (stmts.length) await db.batch(stmts)
}

async function loadSharesForExpenses(
  db: D1Database,
  expenseIDs: number[],
): Promise<Map<number, ExpenseShare[]>> {
  const result = new Map<number, ExpenseShare[]>()
  if (!expenseIDs.length) return result

  // D1 allows at most ~100 bound parameters per query.
  const chunkSize = 80
  for (let offset = 0; offset < expenseIDs.length; offset += chunkSize) {
    const chunk = expenseIDs.slice(offset, offset + chunkSize)
    const placeholders = chunk.map(() => '?').join(',')
    const rows = await db
      .prepare(
        `SELECT sh.expense_id, sh.person_id, p.name AS person_name, sh.share
         FROM expense_shares sh
         JOIN persons p ON p.id = sh.person_id
         WHERE sh.expense_id IN (${placeholders})
         ORDER BY sh.person_id`,
      )
      .bind(...chunk)
      .all<{ expense_id: number; person_id: number; person_name: string; share: number }>()

    for (const row of rows.results || []) {
      const list = result.get(row.expense_id) || []
      list.push({
        person_id: row.person_id,
        person_name: row.person_name,
        share: row.share,
      })
      result.set(row.expense_id, list)
    }
  }
  return result
}

async function fetchExpense(db: D1Database, id: number): Promise<Expense | null> {
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
    .first<{
      id: number
      person_id: number
      person_name: string
      shop_id: number
      shop_name: string
      item_id: number
      name: string
      date: string
      amount: number
    }>()

  if (!row) return null
  const sharesMap = await loadSharesForExpenses(db, [id])
  return {
    id: row.id,
    person_id: row.person_id,
    person_name: row.person_name,
    shop_id: row.shop_id,
    shop_name: row.shop_name,
    item_id: row.item_id,
    name: row.name,
    date: row.date,
    amount: row.amount,
    shares: sharesMap.get(id) || [],
  }
}

app.get('/api/health', (c) => c.json({ status: 'ok' }))

app.get('/api/persons', async (c) => {
  const rows = await c.env.DB.prepare(`SELECT id, name FROM persons ORDER BY id`).all<Person>()
  return c.json(rows.results || [])
})

app.get('/api/shops', async (c) => {
  const q = (c.req.query('q') || '').trim()
  if (q && [...q].length < 3) return c.json([])

  const rows = q
    ? await c.env.DB.prepare(
        `SELECT id, name FROM shops WHERE name LIKE ? COLLATE NOCASE ORDER BY name`,
      )
        .bind(`%${q}%`)
        .all<Shop>()
    : await c.env.DB.prepare(`SELECT id, name FROM shops ORDER BY name`).all<Shop>()
  return c.json(rows.results || [])
})

app.post('/api/shops', async (c) => {
  const body = await c.req.json<{ name?: string }>()
  const name = (body.name || '').trim()
  if (!name) return jsonError('name is required', 400)
  try {
    const result = await c.env.DB.prepare(`INSERT INTO shops (name) VALUES (?)`).bind(name).run()
    return c.json({ id: Number(result.meta.last_row_id), name }, 201)
  } catch (err) {
    if (isUniqueError(err)) return jsonError('shop already exists', 409)
    return jsonError('failed to create shop', 500)
  }
})

app.put('/api/shops/:id', async (c) => {
  const id = Number(c.req.param('id'))
  if (!Number.isFinite(id)) return jsonError('invalid shop id', 400)
  const body = await c.req.json<{ name?: string }>()
  const name = (body.name || '').trim()
  if (!name) return jsonError('name is required', 400)
  try {
    const result = await c.env.DB.prepare(`UPDATE shops SET name = ? WHERE id = ?`).bind(name, id).run()
    if (!result.meta.changes) return jsonError('shop not found', 404)
    return c.json({ id, name })
  } catch (err) {
    if (isUniqueError(err)) return jsonError('shop already exists', 409)
    return jsonError('failed to update shop', 500)
  }
})

app.delete('/api/shops/:id', async (c) => {
  const id = Number(c.req.param('id'))
  if (!Number.isFinite(id)) return jsonError('invalid shop id', 400)
  const inUse = await c.env.DB.prepare(`SELECT COUNT(*) AS c FROM expenses WHERE shop_id = ?`)
    .bind(id)
    .first<{ c: number }>()
  if ((inUse?.c ?? 0) > 0) return jsonError('shop is used by expenses', 409)
  const result = await c.env.DB.prepare(`DELETE FROM shops WHERE id = ?`).bind(id).run()
  if (!result.meta.changes) return jsonError('shop not found', 404)
  return c.body(null, 204)
})

app.get('/api/items', async (c) => {
  const q = (c.req.query('q') || '').trim()
  if (q && [...q].length < 3) return c.json([])

  const rows = q
    ? await c.env.DB.prepare(
        `SELECT id, name FROM items WHERE name LIKE ? COLLATE NOCASE ORDER BY name`,
      )
        .bind(`%${q}%`)
        .all<Item>()
    : await c.env.DB.prepare(`SELECT id, name FROM items ORDER BY name`).all<Item>()
  return c.json(rows.results || [])
})

app.post('/api/items', async (c) => {
  const body = await c.req.json<{ name?: string }>()
  const name = (body.name || '').trim()
  if (!name) return jsonError('name is required', 400)
  try {
    const result = await c.env.DB.prepare(`INSERT INTO items (name) VALUES (?)`).bind(name).run()
    return c.json({ id: Number(result.meta.last_row_id), name }, 201)
  } catch (err) {
    if (isUniqueError(err)) return jsonError('item already exists', 409)
    return jsonError('failed to create item', 500)
  }
})

app.put('/api/items/:id', async (c) => {
  const id = Number(c.req.param('id'))
  if (!Number.isFinite(id)) return jsonError('invalid item id', 400)
  const body = await c.req.json<{ name?: string }>()
  const name = (body.name || '').trim()
  if (!name) return jsonError('name is required', 400)
  try {
    const result = await c.env.DB.prepare(`UPDATE items SET name = ? WHERE id = ?`).bind(name, id).run()
    if (!result.meta.changes) return jsonError('item not found', 404)
    return c.json({ id, name })
  } catch (err) {
    if (isUniqueError(err)) return jsonError('item already exists', 409)
    return jsonError('failed to update item', 500)
  }
})

app.delete('/api/items/:id', async (c) => {
  const id = Number(c.req.param('id'))
  if (!Number.isFinite(id)) return jsonError('invalid item id', 400)
  const inUse = await c.env.DB.prepare(`SELECT COUNT(*) AS c FROM expenses WHERE item_id = ?`)
    .bind(id)
    .first<{ c: number }>()
  if ((inUse?.c ?? 0) > 0) return jsonError('item is used by expenses', 409)
  const result = await c.env.DB.prepare(`DELETE FROM items WHERE id = ?`).bind(id).run()
  if (!result.meta.changes) return jsonError('item not found', 404)
  return c.body(null, 204)
})

app.get('/api/stats', async (c) => {
  const fromDate = c.req.query('from_date') || ''
  const toDate = c.req.query('to_date') || ''
  const where = ['1=1']
  const args: (string | number)[] = []
  if (fromDate) {
    where.push('e.date >= ?')
    args.push(fromDate)
  }
  if (toDate) {
    where.push('e.date <= ?')
    args.push(toDate)
  }

  const rows = await c.env.DB.prepare(
    `SELECT e.date, e.person_id, e.amount,
      COALESCE((SELECT share FROM expense_shares WHERE expense_id = e.id AND person_id = 1), 0) AS armin_share,
      COALESCE((SELECT share FROM expense_shares WHERE expense_id = e.id AND person_id = 2), 0) AS ramin_share
     FROM expenses e
     WHERE ${where.join(' AND ')}`,
  )
    .bind(...args)
    .all<{
      date: string
      person_id: number
      amount: number
      armin_share: number
      ramin_share: number
    }>()

  const byMonth = new Map<string, MonthStats>()
  for (const row of rows.results || []) {
    const month = monthKeyFromGregorian(row.date)
    if (!month) continue
    let stats = byMonth.get(month)
    if (!stats) {
      stats = { month, armin: 0, ramin: 0, total: 0, armin_share: 0, ramin_share: 0 }
      byMonth.set(month, stats)
    }
    stats.total += row.amount
    if (row.person_id === 1) stats.armin += row.amount
    else if (row.person_id === 2) stats.ramin += row.amount
    stats.armin_share += row.amount * row.armin_share
    stats.ramin_share += row.amount * row.ramin_share
  }

  const months = [...byMonth.values()].sort((a, b) => (a.month < b.month ? 1 : -1))
  return c.json({ by_month: months })
})

app.get('/api/expenses/check-duplicate', async (c) => {
  const date = (c.req.query('date') || '').trim()
  if (!date) return jsonError('date is required', 400)
  const excludeID = (c.req.query('exclude_id') || '').trim()
  const itemIDStr = (c.req.query('item_id') || '').trim()
  const name = (c.req.query('name') || '').trim()

  let count = 0
  if (itemIDStr) {
    const itemID = Number(itemIDStr)
    if (!Number.isFinite(itemID)) return jsonError('invalid item_id', 400)
    const row = excludeID
      ? await c.env.DB.prepare(
          `SELECT COUNT(*) AS c FROM expenses WHERE item_id = ? AND date = ? AND id != ?`,
        )
          .bind(itemID, date, excludeID)
          .first<{ c: number }>()
      : await c.env.DB.prepare(`SELECT COUNT(*) AS c FROM expenses WHERE item_id = ? AND date = ?`)
          .bind(itemID, date)
          .first<{ c: number }>()
    count = row?.c ?? 0
  } else if (name) {
    const row = excludeID
      ? await c.env.DB.prepare(
          `SELECT COUNT(*) AS c FROM expenses e
           JOIN items i ON i.id = e.item_id
           WHERE i.name = ? COLLATE NOCASE AND e.date = ? AND e.id != ?`,
        )
          .bind(name, date, excludeID)
          .first<{ c: number }>()
      : await c.env.DB.prepare(
          `SELECT COUNT(*) AS c FROM expenses e
           JOIN items i ON i.id = e.item_id
           WHERE i.name = ? COLLATE NOCASE AND e.date = ?`,
        )
          .bind(name, date)
          .first<{ c: number }>()
    count = row?.c ?? 0
  } else {
    return jsonError('item_id or name is required', 400)
  }

  return c.json({ exists: count > 0, count })
})

app.get('/api/expenses', async (c) => {
  const where = ['1=1']
  const args: (string | number)[] = []
  const personID = c.req.query('person_id')
  const from = c.req.query('from_date')
  const to = c.req.query('to_date')
  if (personID) {
    where.push('e.person_id = ?')
    args.push(personID)
  }
  if (from) {
    where.push('e.date >= ?')
    args.push(from)
  }
  if (to) {
    where.push('e.date <= ?')
    args.push(to)
  }

  const rows = await c.env.DB.prepare(
    `SELECT e.id, e.person_id, p.name AS person_name, e.shop_id, s.name AS shop_name,
            e.item_id, i.name AS name, e.date, e.amount
     FROM expenses e
     JOIN persons p ON p.id = e.person_id
     JOIN shops s ON s.id = e.shop_id
     JOIN items i ON i.id = e.item_id
     WHERE ${where.join(' AND ')}
     ORDER BY e.date DESC, e.id DESC`,
  )
    .bind(...args)
    .all<{
      id: number
      person_id: number
      person_name: string
      shop_id: number
      shop_name: string
      item_id: number
      name: string
      date: string
      amount: number
    }>()

  const expenses: Expense[] = (rows.results || []).map((row) => ({
    id: row.id,
    person_id: row.person_id,
    person_name: row.person_name,
    shop_id: row.shop_id,
    shop_name: row.shop_name,
    item_id: row.item_id,
    name: row.name,
    date: row.date,
    amount: row.amount,
    shares: [],
  }))
  const ids = expenses.map((e) => e.id!).filter(Boolean)
  const sharesMap = await loadSharesForExpenses(c.env.DB, ids)
  for (const expense of expenses) {
    expense.shares = sharesMap.get(expense.id!) || []
  }
  return c.json(expenses)
})

app.get('/api/expenses/:id', async (c) => {
  const id = Number(c.req.param('id'))
  if (!Number.isFinite(id)) return jsonError('invalid expense id', 400)
  const expense = await fetchExpense(c.env.DB, id)
  if (!expense) return jsonError('expense not found', 404)
  return c.json(expense)
})

app.post('/api/expenses', async (c) => {
  const body = await c.req.json<{
    person_id?: number
    shop_id?: number
    date?: string
    items?: { name?: string; amount?: number; shares?: ExpenseShareInput[] }[]
  }>()

  const personId = Number(body.person_id)
  const shopId = Number(body.shop_id)
  const date = (body.date || '').trim()
  const items = body.items || []

  if (personId !== 1 && personId !== 2) return jsonError('person_id must be 1 or 2', 400)
  if (!date) return jsonError('date is required', 400)
  if (!items.length) return jsonError('items is required', 400)

  for (const item of items) {
    const shareErr = validateShares(item.shares || [])
    if (shareErr) return jsonError(shareErr, 400)
    if (!(item.name || '').trim()) return jsonError('item name is required', 400)
    const amountErr = validateWholeAmount(Number(item.amount))
    if (amountErr) return jsonError(amountErr, 400)
  }

  const shop = await c.env.DB.prepare(`SELECT id FROM shops WHERE id = ?`)
    .bind(shopId)
    .first<{ id: number }>()
  if (!shop) return jsonError('invalid shop_id', 400)

  const createdIDs: number[] = []
  for (const item of items) {
    const name = (item.name || '').trim()
    const itemId = await upsertItem(c.env.DB, name)
    const result = await c.env.DB.prepare(
      `INSERT INTO expenses (person_id, shop_id, item_id, date, amount) VALUES (?, ?, ?, ?, ?)`,
    )
      .bind(personId, shopId, itemId, date, Number(item.amount))
      .run()
    const expenseId = Number(result.meta.last_row_id)
    await insertExpenseShares(c.env.DB, expenseId, item.shares || [])
    createdIDs.push(expenseId)
  }

  const created: Expense[] = []
  for (const id of createdIDs) {
    const expense = await fetchExpense(c.env.DB, id)
    if (expense) created.push(expense)
  }
  return c.json(created, 201)
})

app.put('/api/expenses/:id', async (c) => {
  const id = Number(c.req.param('id'))
  if (!Number.isFinite(id)) return jsonError('invalid expense id', 400)

  const body = await c.req.json<{
    person_id?: number
    shop_id?: number
    date?: string
    name?: string
    amount?: number
    shares?: ExpenseShareInput[]
  }>()

  const personId = Number(body.person_id)
  const shopId = Number(body.shop_id)
  const date = (body.date || '').trim()
  const name = (body.name || '').trim()
  const amount = Number(body.amount)
  const shares = body.shares || []

  if (personId !== 1 && personId !== 2) return jsonError('person_id must be 1 or 2', 400)
  const shareErr = validateShares(shares)
  if (shareErr) return jsonError(shareErr, 400)
  const amountErr = validateWholeAmount(amount)
  if (amountErr) return jsonError(amountErr, 400)
  if (!name) return jsonError('name is required', 400)

  const shop = await c.env.DB.prepare(`SELECT id FROM shops WHERE id = ?`)
    .bind(shopId)
    .first<{ id: number }>()
  if (!shop) return jsonError('invalid shop_id', 400)

  const existing = await c.env.DB.prepare(`SELECT id FROM expenses WHERE id = ?`)
    .bind(id)
    .first<{ id: number }>()
  if (!existing) return jsonError('expense not found', 404)

  const itemId = await upsertItem(c.env.DB, name)
  await c.env.DB.prepare(
    `UPDATE expenses SET person_id = ?, shop_id = ?, item_id = ?, date = ?, amount = ? WHERE id = ?`,
  )
    .bind(personId, shopId, itemId, date, amount, id)
    .run()
  await c.env.DB.prepare(`DELETE FROM expense_shares WHERE expense_id = ?`).bind(id).run()
  await insertExpenseShares(c.env.DB, id, shares)

  const expense = await fetchExpense(c.env.DB, id)
  return c.json(expense || { id })
})

app.delete('/api/expenses/:id', async (c) => {
  const id = Number(c.req.param('id'))
  if (!Number.isFinite(id)) return jsonError('invalid expense id', 400)
  await c.env.DB.prepare(`DELETE FROM expense_shares WHERE expense_id = ?`).bind(id).run()
  const result = await c.env.DB.prepare(`DELETE FROM expenses WHERE id = ?`).bind(id).run()
  if (!result.meta.changes) return jsonError('expense not found', 404)
  return c.body(null, 204)
})

app.notFound((c) => {
  if (c.req.path.startsWith('/api')) return jsonError('not found', 404)
  return jsonError('not found', 404)
})

export default app
