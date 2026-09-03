/**
 * Minimal Worker entry for Cloudflare Static Assets + Content-Type fix.
 */
const EXT_MIME = {
  ".html": "text/html; charset=utf-8",
  ".js": "text/javascript; charset=utf-8",
  ".mjs": "text/javascript; charset=utf-8",
  ".css": "text/css; charset=utf-8",
  ".json": "application/json; charset=utf-8",
  ".webmanifest": "application/manifest+json; charset=utf-8",
  ".svg": "image/svg+xml",
  ".png": "image/png",
  ".jpg": "image/jpeg",
  ".jpeg": "image/jpeg",
  ".gif": "image/gif",
  ".webp": "image/webp",
  ".ico": "image/x-icon",
  ".woff": "font/woff",
  ".woff2": "font/woff2",
  ".ttf": "font/ttf",
  ".map": "application/json; charset=utf-8",
  ".txt": "text/plain; charset=utf-8",
};

function mimeForPath(pathname) {
  const base = pathname.split("?")[0];
  const dot = base.lastIndexOf(".");
  if (dot > base.lastIndexOf("/")) {
    const ext = base.slice(dot).toLowerCase();
    if (EXT_MIME[ext]) return EXT_MIME[ext];
  }
  return "text/html; charset=utf-8";
}

export default {
  async fetch(request, env) {
    const response = await env.ASSETS.fetch(request);
    const mime = mimeForPath(new URL(request.url).pathname);
    const headers = new Headers(response.headers);
    headers.set("Content-Type", mime);
    return new Response(response.body, {
      status: response.status,
      statusText: response.statusText,
      headers,
    });
  },
};
