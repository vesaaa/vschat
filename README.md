# VsChat

**English** | [简体中文](README.zh-CN.md)

A small, self-hosted, multi-user AI chat. Plug in your keys for **ChatGPT, Claude, Gemini, Grok, DeepSeek, Qwen, Kimi or GLM** — or any OpenAI-compatible API — and give each person their own account and chat history.

- **Tiny** — one Go binary + one SQLite file. ~25 MB image; uses a few MB of RAM idle and ~25 MB in use. No Node, no Python, no database server.
- **Multi-user** — admin creates accounts; chats are private per user; monthly token quotas (with a default for new users), a warning bar at 80% and a clear stop at 100%; usage dashboard.
- **Admin visibility** — access log (sign-ins, failed sign-ins, chats) with client IP, country and device, works behind Cloudflare; each user's last activity; providers' remaining rate limits and reset times, plus account balance for DeepSeek and Kimi.
- **Shared + personal models** — admins decide which shared providers each user can see; users can optionally bring their own API keys, visible only to them and not counted against their limit.
- **Built-in providers** — pick a provider, paste a key, done: the current mainstream models are built in and update with new releases. Claude goes through the official SDK with thinking shown; reasoning from DeepSeek, Qwen, Kimi and GLM is shown too. Any other OpenAI-compatible endpoint (Ollama, OpenRouter, vLLM, …) works as a custom provider. Switch models mid-chat.
- **Nice to use** — streaming, Markdown + code highlighting, light/dark, mobile layout, 中文 / English.
- **Long pastes become attachments** — paste a big chunk of code/log and it turns into a file card instead of flooding the input. Text files can be attached or dropped in too.

## Quick start (Docker)

```bash
docker run -d --name vschat -p 8080:8080 \
  -v vschat-data:/data \
  -e VSCHAT_ADMIN_PASSWORD=change-me \
  -e VSCHAT_SECRET=a-long-random-string-you-keep \
  vsbox/vschat:latest
```

The same image is on GitHub as `ghcr.io/vesaaa/vschat`. Tags: `latest` (newest release), `0.3.0`-style versions, `edge` (main branch).

Open http://localhost:8080, sign in as `admin`, then go to **Admin → Providers** and paste your API keys.

`VSCHAT_SECRET` encrypts stored API keys — pick one once (e.g. `openssl rand -hex 32`) and keep using the same value; with a different one the saved keys can't be decrypted.

If you don't set `VSCHAT_ADMIN_PASSWORD`, a random one is printed in the log on first start (`docker logs vschat`).

### docker compose

```yaml
services:
  vschat:
    image: vsbox/vschat:latest
    restart: unless-stopped
    ports: ["8080:8080"]
    volumes:
      - ./data:/data            # everything VsChat stores lives here
    environment:
      VSCHAT_ADMIN_PASSWORD: change-me
      VSCHAT_SECRET: a-long-random-string-you-keep
      # PUID: 1000              # run as your host user so ./data stays yours
      # PGID: 1000
      # Optional: pre-fill keys on first start (can also be set in the admin UI)
      # OPENAI_API_KEY: sk-...
      # ANTHROPIC_API_KEY: sk-ant-...
      # DEEPSEEK_API_KEY: sk-...
```

A named volume (`vschat-data:/data`) works just as well. Secrets can come from files instead — `VSCHAT_SECRET_FILE: /run/secrets/vschat_secret` with Docker secrets.

## Configuration

| Variable | Default | |
|---|---|---|
| `VSCHAT_ADDR` | `:8080` | Listen address |
| `VSCHAT_DATA` | `/data` (Docker), `./data` | Where `vschat.db` lives — back up this one file |
| `VSCHAT_ADMIN_USER` | `admin` | First admin, created only when the database is empty |
| `VSCHAT_ADMIN_PASSWORD` | random | Password for that admin |
| `OPENAI_API_KEY` `ANTHROPIC_API_KEY` `GEMINI_API_KEY` `XAI_API_KEY` `DEEPSEEK_API_KEY` `DASHSCOPE_API_KEY` `MOONSHOT_API_KEY` `ZHIPU_API_KEY` | | On first start, adds a provider for each key given |
| `VSCHAT_SECRET` | | Encrypts stored API keys (AES-GCM). Set it before adding keys; existing plain keys are encrypted on start. Keep it safe — keys can't be read without it |
| `PUID` / `PGID` | `10001` | Docker: user the server runs as. The container fixes `/data` ownership on start, so bind mounts work |
| `VSCHAT_TRUST_PROXY` | `auto` | Which forwarding headers to believe for the client IP. `auto` trusts them only when the connection comes from a Cloudflare edge, localhost or a private network (e.g. your nginx); `1` always (origin reachable only through your proxy); `0` never |

`VSCHAT_SECRET`, `VSCHAT_ADMIN_PASSWORD` and the `*_API_KEY` variables also accept a `_FILE` suffix pointing at a file to read (Docker / Kubernetes secrets).

There is no config file: everything else (providers, models, users, limits, settings) lives in the database and is managed in the admin panel. `/data` is the only directory to persist and back up.

### Providers

Built in (see [`catalog.go`](catalog.go)); Chinese providers default to their mainland endpoints, the rest to the official global ones:

| Provider | Endpoint | Built-in models |
|---|---|---|
| OpenAI | api.openai.com | gpt-6.1-sol, gpt-6-astra, gpt-6.1-terra, gpt-6-luna |
| Claude | api.anthropic.com (official SDK) | claude-opus-5-5, claude-sonnet-5-5, claude-haiku-5-5 |
| Gemini | generativelanguage.googleapis.com (OpenAI-compatible) | gemini-4-argon, gemini-3.8-flash, gemini-3.5-flash-lite |
| Grok | api.x.ai | grok-4.7, grok-4.6 |
| DeepSeek | api.deepseek.com | deepseek-v4-pro, deepseek-flash |
| 通义千问 Qwen | dashscope.aliyuncs.com | qwen3.8-max, qwen3.7-plus, qwen3.8-flash |
| Kimi | api.moonshot.cn | kimi-k3, kimi-k2.6 |
| 智谱 GLM | open.bigmodel.cn | glm-5.3, glm-5.3-flash |

A provider added from the list **follows the built-in model list** by default, so a new VsChat release brings new models with no clicks. Untick it to pick models yourself, add extra model IDs, or use **Fetch from provider**. The address can be changed under *Advanced* (e.g. Kimi's global `api.moonshot.ai`). **Custom** providers take any OpenAI-compatible or Anthropic-compatible endpoint.

### Who can use what

- **Shared providers** (Admin → Providers) are visible to *everyone* or to *selected users* — e.g. a "Claude Opus" provider only for a few people. Add the same API several times to split models into groups.
- **Personal providers** — *Admin → Settings → Users' own API keys*:
  - **Not allowed** (default) — users only see shared models.
  - **Official providers only** — users pick one of the built-in providers and paste their own key; the API address is fixed to the official endpoint.
  - **Allow custom API URLs** — users may also point at any public OpenAI-compatible or Anthropic endpoint (relays, OpenRouter, …).

  Users then get **My models** in their menu. Those providers are private to the user, show up as "Mine" in the model picker, and their tokens are tracked separately (not counted against the monthly limit). Switching to a stricter mode disables personal providers that no longer fit; they're kept and marked.
- In custom mode, personal providers can't reach `localhost` or private/LAN addresses unless you turn on *Let personal providers reach private networks* (e.g. for a LAN Ollama), so users can't use the server to probe your network.

## Client IP behind Cloudflare / a reverse proxy

VsChat records the visitor's IP in the access log and uses it for sign-in rate limiting. With the default `VSCHAT_TRUST_PROXY=auto` it reads `CF-Connecting-IP`, `True-Client-IP`, the first `X-Forwarded-For` hop, then `X-Real-IP` — but only when the request arrives from a Cloudflare edge address, localhost or a private network, so clients hitting the port directly can't fake their IP. Behind Cloudflare the visitor's country (`CF-IPCountry`) is recorded too.

## Behind a reverse proxy

Streaming uses server-sent events; turn off response buffering. nginx:

```nginx
location / {
    proxy_pass http://127.0.0.1:8080;
    proxy_http_version 1.1;
    proxy_buffering off;
    proxy_read_timeout 1h;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
}
```

No cache rules are needed: scripts and styles are served under content-hashed URLs (`/s/<hash>/app.js`) with long-lived `immutable` caching, the page itself is always revalidated, and API responses are `no-store` — so after an upgrade browsers and Cloudflare pick up the new version on the next load.

Static files are pre-compressed with gzip.

Serve it over HTTPS if it's reachable from the internet — the session cookie is marked `Secure` automatically when it is.

## Build from source

```bash
go build -o vschat .
./vschat            # http://localhost:8080
```

Requires Go 1.24+. The web UI is plain HTML/CSS/JS embedded in the binary — no build step.

## Notes

- API keys never leave the server. They're stored in the SQLite file — encrypted when `VSCHAT_SECRET` is set. Protect the data volume either way.
- Each user's usage is counted from the token counts the providers report.
- Attachments are text only for now (code, logs, Markdown, CSV…, up to 1 MB each).

## License

Copyright (C) 2026 vesaaa

VsChat is free software under the [GNU Affero General Public License v3.0](LICENSE) (AGPL-3.0), from v0.3.5 on. You may use, modify and redistribute it; if you run a modified version as a network service, you must offer its users the corresponding source code (AGPL §13). Releases up to v0.3.4 were published under MIT.
