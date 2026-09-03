const fs = require('fs')
const path = require('path')
const dir = __dirname
const jwt = fs.readFileSync(path.join(dir, 'upload-jwt.txt'), 'utf8').trim()
const worker = fs.readFileSync('C:/Users/armin/GitHub/exar/exar-webui/worker.js', 'utf8')
const code =
  'async () => {\n' +
  '  const completionJwt = ' +
  JSON.stringify(jwt) +
  ';\n' +
  '  const worker = ' +
  JSON.stringify(worker) +
  ';\n' +
  `  const metadata = {
    main_module: "worker.js",
    compatibility_date: "2026-08-31",
    assets: {
      jwt: completionJwt,
      config: {
        not_found_handling: "single-page-application",
        run_worker_first: true,
      },
    },
    bindings: [{ type: "assets", name: "ASSETS" }],
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
    worker,
    "--" + b + "--",
    "",
  ].join("\\r\\n");
  const put = await cloudflare.request({
    method: "PUT",
    path: "/accounts/" + accountId + "/workers/scripts/exar",
    body,
    contentType: "multipart/form-data; boundary=" + b,
    rawBody: true,
  });
  const sub = await cloudflare.request({
    method: "POST",
    path: "/accounts/" + accountId + "/workers/scripts/exar/subdomain",
    body: { enabled: true },
  });
  return {
    putSuccess: put.success,
    putStatus: put.status,
    putErrors: put.errors,
    subSuccess: sub.success,
    subErrors: sub.errors,
  };
}
`
fs.writeFileSync(path.join(dir, 'deploy-webui-code.js'), code)
console.log('wrote deploy-webui-code.js', code.length)
