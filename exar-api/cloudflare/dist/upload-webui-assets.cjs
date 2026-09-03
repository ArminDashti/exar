const fs = require('fs')
const path = require('path')
const https = require('https')

const accountId = '0b9173381c4580ae6fc430fff3d08018'
const jwtPath = path.join(__dirname, 'upload-jwt.txt')
const assets = JSON.parse(fs.readFileSync(path.join(__dirname, 'webui-assets.json'), 'utf8'))
const session = JSON.parse(fs.readFileSync(path.join(__dirname, 'upload-session.json'), 'utf8'))

let jwt = session.result.jwt
const buckets = session.result.buckets

function multipart(parts) {
  const b = '----formboundary' + Date.now()
  const chunks = []
  for (const [name, value, contentType] of parts) {
    chunks.push(`--${b}\r\n`)
    chunks.push(`Content-Disposition: form-data; name="${name}"\r\n`)
    if (contentType) chunks.push(`Content-Type: ${contentType}\r\n`)
    chunks.push(`\r\n`)
    chunks.push(value)
    chunks.push(`\r\n`)
  }
  chunks.push(`--${b}--\r\n`)
  return { body: Buffer.concat(chunks.map((c) => (Buffer.isBuffer(c) ? c : Buffer.from(c, 'utf8')))), boundary: b }
}

function request(method, urlPath, headers, body) {
  return new Promise((resolve, reject) => {
    const req = https.request(
      {
        hostname: 'api.cloudflare.com',
        path: urlPath,
        method,
        headers,
      },
      (res) => {
        const chunks = []
        res.on('data', (d) => chunks.push(d))
        res.on('end', () => {
          const text = Buffer.concat(chunks).toString('utf8')
          let json
          try {
            json = JSON.parse(text)
          } catch {
            json = { raw: text }
          }
          resolve({ status: res.statusCode, json })
        })
      }
    )
    req.on('error', reject)
    if (body) req.write(body)
    req.end()
  })
}

;(async () => {
  for (let i = 0; i < buckets.length; i++) {
    const bucket = buckets[i]
    const parts = []
    for (const hash of bucket) {
      const a = assets[hash]
      if (!a) throw new Error('missing asset ' + hash)
      parts.push([hash, a.b64, a.mime])
    }
    const { body, boundary } = multipart(parts)
    const res = await request(
      'POST',
      `/client/v4/accounts/${accountId}/workers/assets/upload?base64=true`,
      {
        Authorization: 'Bearer ' + jwt,
        'Content-Type': `multipart/form-data; boundary=${boundary}`,
        'Content-Length': body.length,
      },
      body
    )
    console.log('bucket', i, 'status', res.status, 'success', res.json.success, 'hasJwt', !!(res.json.result && res.json.result.jwt))
    if (!res.json.success) {
      console.log(JSON.stringify(res.json).slice(0, 500))
      process.exit(1)
    }
    if (res.json.result && res.json.result.jwt) jwt = res.json.result.jwt
  }
  fs.writeFileSync(jwtPath, jwt)
  console.log('completion jwt saved, len', jwt.length)
})().catch((e) => {
  console.error(e)
  process.exit(1)
})
