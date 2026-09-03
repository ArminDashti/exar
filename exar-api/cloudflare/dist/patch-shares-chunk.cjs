const fs = require('fs')

function patchMin() {
  const p = __dirname + '/native-worker.min.js'
  let s = fs.readFileSync(p, 'utf8')
  const start = s.indexOf('async function C(s,o){')
  if (start < 0) throw new Error('C not found')
  const endMarker = 'return a}async function T('
  const end = s.indexOf(endMarker, start)
  if (end < 0) throw new Error('end not found')
  const neu =
    'async function C(s,o){const a=new Map;if(!o.length)return a;for(let f=0;f<o.length;f+=80){const c=o.slice(f,f+80),i=c.map(()=>"?").join(","),n=await s.prepare(`SELECT sh.expense_id, sh.person_id, p.name AS person_name, sh.share\n       FROM expense_shares sh JOIN persons p ON p.id = sh.person_id\n       WHERE sh.expense_id IN (${i}) ORDER BY sh.person_id`).bind(...c).all();for(const d of n.results||[]){const h=a.get(d.expense_id)||[];h.push({person_id:d.person_id,person_name:d.person_name,share:d.share}),a.set(d.expense_id,h)}}return a}'
  s = s.slice(0, start) + neu + s.slice(end)
  fs.writeFileSync(p, s)
  console.log('min patched', s.length)
}

function patchPretty() {
  const p = __dirname + '/native-worker.js'
  let s = fs.readFileSync(p, 'utf8')
  const start = s.indexOf('async function loadShares(db, ids) {')
  if (start < 0) throw new Error('loadShares not found')
  const end = s.indexOf('\nasync function fetchExpense', start)
  if (end < 0) throw new Error('fetchExpense not found')
  const neu = `async function loadShares(db, ids) {
  const out = new Map();
  if (!ids.length) return out;
  const chunkSize = 80;
  for (let offset = 0; offset < ids.length; offset += chunkSize) {
    const chunk = ids.slice(offset, offset + chunkSize);
    const ph = chunk.map(() => "?").join(",");
    const { results } = await db
      .prepare(
        \`SELECT sh.expense_id, sh.person_id, p.name AS person_name, sh.share
       FROM expense_shares sh JOIN persons p ON p.id = sh.person_id
       WHERE sh.expense_id IN (\${ph}) ORDER BY sh.person_id\`,
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
`
  s = s.slice(0, start) + neu + s.slice(end)
  fs.writeFileSync(p, s)
  console.log('pretty patched', s.length)
}

patchMin()
patchPretty()
