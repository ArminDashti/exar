async () => {
  const completionJwt = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJodHRwczovL2FwaS53b3JrZXJzLmNsb3VkZmxhcmUuY29tIiwiYXVkIjoiZXdjIiwiZXhwIjoxNzg4NDQyMTQ4LCJpYXQiOjE3ODg0Mzg1NDgsIm5iZiI6MTc4ODQzODU0OCwibWFuaWZlc3RfaWQiOiIyZTg0MjVkMS00YjBlLTRlYTItODYxMC1hZTBhM2EyNjBhNDIiLCJhY2NvdW50X2lkIjo5NTA4MzAzNn0.D2Fw4oSXoOJy0ddG2CsUsORzhZTZB1aC8QRSTktMbb8";
  const worker = "/**\r\n * Minimal Worker entry for Cloudflare Static Assets + Content-Type fix.\r\n */\r\nconst EXT_MIME = {\r\n  \".html\": \"text/html; charset=utf-8\",\r\n  \".js\": \"text/javascript; charset=utf-8\",\r\n  \".mjs\": \"text/javascript; charset=utf-8\",\r\n  \".css\": \"text/css; charset=utf-8\",\r\n  \".json\": \"application/json; charset=utf-8\",\r\n  \".webmanifest\": \"application/manifest+json; charset=utf-8\",\r\n  \".svg\": \"image/svg+xml\",\r\n  \".png\": \"image/png\",\r\n  \".jpg\": \"image/jpeg\",\r\n  \".jpeg\": \"image/jpeg\",\r\n  \".gif\": \"image/gif\",\r\n  \".webp\": \"image/webp\",\r\n  \".ico\": \"image/x-icon\",\r\n  \".woff\": \"font/woff\",\r\n  \".woff2\": \"font/woff2\",\r\n  \".ttf\": \"font/ttf\",\r\n  \".map\": \"application/json; charset=utf-8\",\r\n  \".txt\": \"text/plain; charset=utf-8\",\r\n};\r\n\r\nfunction mimeForPath(pathname) {\r\n  const base = pathname.split(\"?\")[0];\r\n  const dot = base.lastIndexOf(\".\");\r\n  if (dot > base.lastIndexOf(\"/\")) {\r\n    const ext = base.slice(dot).toLowerCase();\r\n    if (EXT_MIME[ext]) return EXT_MIME[ext];\r\n  }\r\n  return \"text/html; charset=utf-8\";\r\n}\r\n\r\nexport default {\r\n  async fetch(request, env) {\r\n    const response = await env.ASSETS.fetch(request);\r\n    const mime = mimeForPath(new URL(request.url).pathname);\r\n    const headers = new Headers(response.headers);\r\n    headers.set(\"Content-Type\", mime);\r\n    return new Response(response.body, {\r\n      status: response.status,\r\n      statusText: response.statusText,\r\n      headers,\r\n    });\r\n  },\r\n};\r\n";
  const metadata = {
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
  ].join("\r\n");
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
