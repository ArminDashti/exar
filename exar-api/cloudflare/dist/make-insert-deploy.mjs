import fs from 'fs'

const parts = JSON.parse(
  fs.readFileSync(new URL('./wrapchunks/parts.json', import.meta.url), 'utf8'),
)

// part 0 already inserted; insert 1..end then deploy
const rest = parts.slice(1)
const fn = `async () => {
  const db = "9bd95dca-6c2f-439a-a4d7-53c44a82118b";
  const rest = ${JSON.stringify(rest)};
  for (let i = 0; i < rest.length; i++) {
    const r = await cloudflare.request({
      method: "POST",
      path: "/accounts/" + accountId + "/d1/database/" + db + "/query",
      body: { sql: "INSERT INTO _deploy_chunks (i, c) VALUES (?, ?)", params: [i + 1, rest[i]] },
    });
    if (!r.success) return { step: "insert", i: i + 1, errors: r.errors };
  }
  const sel = await cloudflare.request({
    method: "POST",
    path: "/accounts/" + accountId + "/d1/database/" + db + "/query",
    body: { sql: "SELECT i, c FROM _deploy_chunks ORDER BY i" },
  });
  if (!sel.success) return { step: "select", errors: sel.errors };
  const rows = sel.result?.[0]?.results || sel.result?.results || [];
  const flat = Array.isArray(sel.result) ? (sel.result[0]?.results || []) : rows;
  const b64 = flat.map((r) => r.c).join("");
  const script = atob(b64);
  const metadata = {
    main_module: "worker.js",
    compatibility_date: "2026-08-31",
    compatibility_flags: ["nodejs_compat"],
    bindings: [
      { type: "d1", name: "DB", id: db },
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
  await cloudflare.request({
    method: "POST",
    path: "/accounts/" + accountId + "/d1/database/" + db + "/query",
    body: { sql: "DROP TABLE IF EXISTS _deploy_chunks" },
  });
  return {
    putSuccess: put.success,
    putErrors: put.errors,
    putStatus: put.status,
    scriptLen: script.length,
    chunks: flat.length,
    subSuccess: sub.success,
    subErrors: sub.errors,
  };
}`

const out = fn.replace('.join("\\r\\n")', '.join("\\r\\n")')
fs.writeFileSync(new URL('./insert-deploy.js', import.meta.url), out)
console.log('bytes', Buffer.byteLength(out))
