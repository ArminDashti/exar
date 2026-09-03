import fs from 'fs'

const s = fs.readFileSync(new URL('./native-worker.min.js', import.meta.url), 'utf8')
const b64 = Buffer.from(s, 'utf8').toString('base64')
const size = 1200
const parts = []
for (let o = 0; o < b64.length; o += size) parts.push(b64.slice(o, o + size))

const fn = `async () => {
  const parts = ${JSON.stringify(parts)};
  const script = atob(parts.join(""));
  const metadata = {
    main_module: "worker.js",
    compatibility_date: "2026-08-31",
    compatibility_flags: ["nodejs_compat"],
    bindings: [
      { type: "d1", name: "DB", id: "9bd95dca-6c2f-439a-a4d7-53c44a82118b" },
      {
        type: "plain_text",
        name: "CORS_ORIGINS",
        text: "http://localhost:5173,http://127.0.0.1:5173,https://exar.armindashti.workers.dev",
      },
    ],
  };
  const b = "----formboundary" + Date.now();
  const body = [
    "--" + b,
    'Content-Disposition: form-data; name="metadata"; filename="metadata.json"',
    "Content-Type: application/json",
    "",
    JSON.stringify(metadata),
    "--" + b,
    'Content-Disposition: form-data; name="worker.js"; filename="worker.js"',
    "Content-Type: application/javascript+module",
    "",
    script,
    "--" + b + "--",
    "",
  ].join("\\r\\n");
  const put = await cloudflare.request({
    method: "PUT",
    path: "/accounts/" + accountId + "/workers/scripts/exar-api",
    body,
    contentType: "multipart/form-data; boundary=" + b,
    rawBody: true,
  });
  const sub = await cloudflare.request({
    method: "POST",
    path: "/accounts/" + accountId + "/workers/scripts/exar-api/subdomain",
    body: { enabled: true },
  });
  return {
    putSuccess: put.success,
    putErrors: put.errors,
    putStatus: put.status,
    putResult: put.result && { id: put.result.id, etag: put.result.etag },
    subSuccess: sub.success,
    subErrors: sub.errors,
  };
}`

// Template has \\r\\n -> write real \r\n join string for the executed code
const out = fn.replace('.join("\\r\\n")', '.join("\\r\\n")')
fs.writeFileSync(new URL('./deploy-parts.js', import.meta.url), out)
console.log('bytes', Buffer.byteLength(out), 'parts', parts.length)
