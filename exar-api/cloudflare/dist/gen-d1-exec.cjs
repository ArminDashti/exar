const fs = require('fs')
const path = require('path')
const dir = path.join(__dirname, 'd1-import-batches')
const files = fs.readdirSync(dir).filter((f) => f.endsWith('.sql')).sort()
const db = '9bd95dca-6c2f-439a-a4d7-53c44a82118b'
for (const f of files) {
  const sql = fs.readFileSync(path.join(dir, f), 'utf8')
  const code =
    'async () => {\n' +
    '  const sql = ' +
    JSON.stringify(sql) +
    ';\n' +
    '  return cloudflare.request({\n' +
    '    method: "POST",\n' +
    '    path: "/accounts/" + accountId + "/d1/database/' +
    db +
    '/query",\n' +
    '    body: { sql },\n' +
    '  });\n' +
    '}\n'
  fs.writeFileSync(path.join(__dirname, 'd1-exec-' + f.replace('.sql', '') + '.js'), code)
}
console.log('wrote', files.length, 'exec scripts')
