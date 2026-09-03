const fs = require('fs')
const path = __dirname + '/'
const db = '9bd95dca-6c2f-439a-a4d7-53c44a82118b'
for (let i = 0; i < 11; i++) {
  const c = fs.readFileSync(path + 'wc' + i + '.txt', 'utf8')
  const code =
    'async () => {\n' +
    '  const i = ' +
    i +
    ';\n' +
    '  const c = ' +
    JSON.stringify(c) +
    ';\n' +
    '  return cloudflare.request({\n' +
    '    method: "POST",\n' +
    '    path: "/accounts/" + accountId + "/d1/database/' +
    db +
    '/query",\n' +
    '    body: { sql: "INSERT INTO _deploy_chunks (i, c) VALUES (?, ?)", params: [i, c] }\n' +
    '  });\n' +
    '}'
  fs.writeFileSync(path + 'ins' + i + '.js', code)
  fs.writeFileSync(path + 'insw' + i + '.txt', code.match(/.{1,80}/g).join('\n'))
}
console.log('ok')
