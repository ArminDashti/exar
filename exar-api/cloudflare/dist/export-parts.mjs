import fs from 'fs'

const s = fs.readFileSync(new URL('./native-worker.min.js', import.meta.url), 'utf8')
const b64 = Buffer.from(s, 'utf8').toString('base64')
const size = 1200
const parts = []
for (let o = 0; o < b64.length; o += size) parts.push(b64.slice(o, o + size))

const dir = new URL('./wrapchunks/', import.meta.url)
fs.mkdirSync(dir, { recursive: true })
for (let i = 0; i < parts.length; i++) {
  fs.writeFileSync(new URL(`part${i}.b64`, dir), parts[i])
}
fs.writeFileSync(new URL('parts.json', dir), JSON.stringify(parts))
console.log('parts', parts.length)
