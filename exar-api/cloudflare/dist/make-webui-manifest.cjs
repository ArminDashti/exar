const fs = require('fs')
const path = require('path')
const crypto = require('crypto')

function walk(dir, base = dir, out = []) {
  for (const name of fs.readdirSync(dir)) {
    const full = path.join(dir, name)
    if (fs.statSync(full).isDirectory()) walk(full, base, out)
    else out.push(full)
  }
  return out
}

function mimeFor(filePath) {
  const ext = path.extname(filePath).toLowerCase()
  const map = {
    '.html': 'text/html; charset=utf-8',
    '.js': 'text/javascript; charset=utf-8',
    '.mjs': 'text/javascript; charset=utf-8',
    '.css': 'text/css; charset=utf-8',
    '.json': 'application/json; charset=utf-8',
    '.webmanifest': 'application/manifest+json; charset=utf-8',
    '.svg': 'image/svg+xml',
    '.png': 'image/png',
    '.ico': 'image/x-icon',
    '.txt': 'text/plain; charset=utf-8',
  }
  return map[ext] || 'application/octet-stream'
}

const dist = path.join('C:/Users/armin/GitHub/exar/exar-webui/dist')
const files = walk(dist)
const manifest = {}
const assets = {}
for (const full of files) {
  const buf = fs.readFileSync(full)
  const hash = crypto.createHash('sha256').update(buf).digest('hex').slice(0, 32)
  const rel = '/' + path.relative(dist, full).split(path.sep).join('/')
  manifest[rel] = { hash, size: buf.length }
  assets[hash] = {
    path: rel,
    mime: mimeFor(full),
    b64: buf.toString('base64'),
  }
}
const outDir = 'C:/Users/armin/GitHub/exar/exar-api/cloudflare/dist'
fs.writeFileSync(path.join(outDir, 'webui-manifest.json'), JSON.stringify(manifest, null, 2))
fs.writeFileSync(path.join(outDir, 'webui-assets.json'), JSON.stringify(assets))
console.log(JSON.stringify({ fileCount: files.length, manifest }, null, 2))
