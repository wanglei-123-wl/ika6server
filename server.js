const http = require('http');
const fs = require('fs');
const os = require('os');
const path = require('path');
const { URL } = require('url');

const port = process.env.PORT ? Number(process.env.PORT) : 3000;
const host = process.env.HOST || '0.0.0.0';
const frontendDist = path.join(__dirname, 'frontend', 'dist');

function send(res, statusCode, body, headers = {}) {
  res.writeHead(statusCode, {
    'Content-Type': 'text/html; charset=utf-8',
    'Cache-Control': 'no-store',
    ...headers,
  });
  res.end(body);
}

function json(res, statusCode, data) {
  res.writeHead(statusCode, {
    'Content-Type': 'application/json; charset=utf-8',
    'Cache-Control': 'no-store',
  });
  res.end(JSON.stringify(data, null, 2));
}

function sendFile(res, filePath) {
  const ext = path.extname(filePath).toLowerCase();
  const contentTypes = {
    '.css': 'text/css; charset=utf-8',
    '.html': 'text/html; charset=utf-8',
    '.js': 'text/javascript; charset=utf-8',
    '.json': 'application/json; charset=utf-8',
    '.svg': 'image/svg+xml',
  };

  fs.readFile(filePath, (error, content) => {
    if (error) {
      json(res, 404, {
        ok: false,
        error: 'not_found',
      });
      return;
    }

    res.writeHead(200, {
      'Content-Type': contentTypes[ext] || 'application/octet-stream',
      'Cache-Control': 'no-store',
    });
    res.end(content);
  });
}

function escapeHtml(value) {
  return String(value)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;');
}

const samplePosts = [
  { id: 1, title: 'Node 后台起步页', author: 'admin', downloads: 128, updatedAt: '2026-09-05 09:00' },
  { id: 2, title: '源码分享区', author: 'moderator', downloads: 84, updatedAt: '2026-09-04 18:20' },
  { id: 3, title: '上传审核队列', author: 'system', downloads: 12, updatedAt: '2026-09-04 16:10' },
];

function getLanUrls() {
  return Object.values(os.networkInterfaces())
    .flat()
    .filter((item) => item && item.family === 'IPv4' && !item.internal)
    .map((item) => `http://${item.address}:${port}`);
}

function renderHomePage(baseUrl, lanUrls) {
  const rows = samplePosts
    .map(
      (post) => `
        <tr>
          <td>${post.id}</td>
          <td>${escapeHtml(post.title)}</td>
          <td>${escapeHtml(post.author)}</td>
          <td>${post.downloads}</td>
          <td>${escapeHtml(post.updatedAt)}</td>
        </tr>`
    )
    .join('');

  return `<!doctype html>
  <html lang="zh-CN">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>ika6server 后台入口</title>
    <style>
      :root {
        color-scheme: light;
        --bg: #f6f7fb;
        --panel: #ffffff;
        --border: #d9deea;
        --text: #132238;
        --muted: #5f6b85;
        --accent: #2f6fed;
        --accent-2: #1f9d6a;
      }
      * { box-sizing: border-box; }
      body {
        margin: 0;
        font-family: Inter, "Segoe UI", Arial, sans-serif;
        background: var(--bg);
        color: var(--text);
      }
      .shell {
        max-width: 1120px;
        margin: 0 auto;
        padding: 24px;
      }
      .hero {
        background: linear-gradient(180deg, #ffffff 0%, #f9fbff 100%);
        border: 1px solid var(--border);
        border-radius: 14px;
        padding: 24px;
        margin-bottom: 18px;
      }
      .title {
        font-size: 28px;
        margin: 0 0 8px;
      }
      .subtitle {
        margin: 0 0 18px;
        color: var(--muted);
      }
      .grid {
        display: grid;
        grid-template-columns: repeat(3, minmax(0, 1fr));
        gap: 12px;
        margin-bottom: 18px;
      }
      .card {
        background: var(--panel);
        border: 1px solid var(--border);
        border-radius: 12px;
        padding: 16px;
      }
      .label {
        font-size: 12px;
        color: var(--muted);
        text-transform: uppercase;
        letter-spacing: 0.04em;
      }
      .value {
        font-size: 28px;
        font-weight: 700;
        margin-top: 6px;
      }
      .badge {
        display: inline-block;
        padding: 4px 10px;
        border-radius: 999px;
        background: rgba(31, 157, 106, 0.1);
        color: var(--accent-2);
        font-size: 12px;
        margin-left: 8px;
      }
      .section {
        background: var(--panel);
        border: 1px solid var(--border);
        border-radius: 12px;
        overflow: hidden;
      }
      .section-header {
        padding: 16px;
        border-bottom: 1px solid var(--border);
        font-weight: 600;
      }
      table {
        width: 100%;
        border-collapse: collapse;
      }
      th, td {
        padding: 14px 16px;
        border-bottom: 1px solid #edf1f7;
        text-align: left;
      }
      th {
        color: var(--muted);
        font-size: 12px;
        text-transform: uppercase;
        letter-spacing: 0.04em;
      }
      .footer {
        margin-top: 18px;
        color: var(--muted);
        font-size: 13px;
      }
      .url-list {
        display: flex;
        flex-wrap: wrap;
        gap: 8px;
        margin: 0;
        padding: 0;
        list-style: none;
      }
      code {
        background: #eef3ff;
        padding: 2px 6px;
        border-radius: 6px;
      }
      @media (max-width: 840px) {
        .grid { grid-template-columns: 1fr; }
      }
    </style>
  </head>
  <body>
    <div class="shell">
      <div class="hero">
        <h1 class="title">ika6server 后台入口 <span class="badge">running</span></h1>
        <p class="subtitle">这是你现在可以直接在浏览器里打开的本地可视化入口。后面我们会在这个基础上继续加用户、帖子、文件和审核模块。</p>
        <p class="subtitle">本机服务地址：<code>${escapeHtml(baseUrl)}</code></p>
        <ul class="url-list">
          ${lanUrls.map((url) => `<li><code>${escapeHtml(url)}</code></li>`).join('')}
        </ul>
      </div>

      <div class="grid">
        <div class="card">
          <div class="label">Service</div>
          <div class="value">OK</div>
        </div>
        <div class="card">
          <div class="label">API</div>
          <div class="value">Ready</div>
        </div>
        <div class="card">
          <div class="label">Storage</div>
          <div class="value">Stub</div>
        </div>
      </div>

      <div class="section">
        <div class="section-header">示例内容</div>
        <table>
          <thead>
            <tr>
              <th>ID</th>
              <th>标题</th>
              <th>作者</th>
              <th>下载</th>
              <th>更新时间</th>
            </tr>
          </thead>
          <tbody>
            ${rows}
          </tbody>
        </table>
      </div>

      <div class="footer">
        健康检查接口：<code>/api/health</code>，示例数据接口：<code>/api/posts</code>
      </div>
    </div>

    <script>
      fetch('/api/health')
        .then((res) => res.json())
        .then((data) => {
          document.querySelector('.badge').textContent = data.ok ? 'running' : 'degraded';
        })
        .catch(() => {
          document.querySelector('.badge').textContent = 'offline';
        });
    </script>
  </body>
  </html>`;
}

const server = http.createServer((req, res) => {
  const requestUrl = new URL(req.url, `http://${req.headers.host || `${host}:${port}`}`);

  if (requestUrl.pathname.startsWith('/assets/')) {
    sendFile(res, path.join(frontendDist, requestUrl.pathname));
    return;
  }

  if (requestUrl.pathname === '/' || requestUrl.pathname === '/dashboard') {
    const indexPath = path.join(frontendDist, 'index.html');
    if (fs.existsSync(indexPath)) {
      sendFile(res, indexPath);
      return;
    }

    send(res, 200, renderHomePage(`http://127.0.0.1:${port}`, getLanUrls()));
    return;
  }

  if (requestUrl.pathname === '/api/health') {
    json(res, 200, {
      ok: true,
      service: 'ika6server',
      time: new Date().toISOString(),
    });
    return;
  }

  if (requestUrl.pathname === '/api/posts') {
    json(res, 200, {
      ok: true,
      items: samplePosts,
    });
    return;
  }

  json(res, 404, {
    ok: false,
    error: 'not_found',
    path: requestUrl.pathname,
  });
});

server.listen(port, host, () => {
  console.log(`Server is running at http://127.0.0.1:${port}`);
  getLanUrls().forEach((url) => console.log(`LAN access: ${url}`));
});
