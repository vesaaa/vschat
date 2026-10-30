# VsChat

A small, self-hosted, multi-user AI chat. Plug in your **OpenAI (ChatGPT)**, **Anthropic (Claude)** and **DeepSeek** keys — or any OpenAI-compatible API — and give each person their own account and chat history.

一个小巧的自托管多用户 AI 对话。填入 ChatGPT / Claude / DeepSeek（或任何 OpenAI 兼容接口）的 Key，每个用户有自己的账号和对话记录。

- **Tiny** — one Go binary + one SQLite file. ~25 MB image; uses a few MB of RAM idle and ~25 MB in use. No Node, no Python, no database server.
- **Multi-user** — admin creates accounts; chats are private per user; optional monthly token limit per user; usage dashboard.
- **Shared + personal models** — admins decide which shared providers each user can see; users can optionally bring their own API keys, visible only to them and not counted against their limit.
- **Multi-provider** — OpenAI, Claude (official SDK, with thinking shown), DeepSeek (reasoning shown), plus any OpenAI-compatible endpoint (Ollama, OpenRouter, vLLM, …). Switch models mid-chat.
- **Nice to use** — streaming, Markdown + code highlighting, light/dark, mobile layout, 中文 / English.
- **Long pastes become attachments** — paste a big chunk of code/log and it turns into a file card instead of flooding the input. Text files can be attached or dropped in too.

## Quick start (Docker)

```bash
docker run -d --name vschat -p 8080:8080 \
  -v vschat-data:/data \
  -e VSCHAT_ADMIN_PASSWORD=change-me \
  -e VSCHAT_SECRET=a-long-random-string-you-keep \
  ghcr.io/vesaaa/vschat:latest
```

Open http://localhost:8080, sign in as `admin`, then go to **Admin → Providers** and paste your API keys.

`VSCHAT_SECRET` encrypts stored API keys — pick one once (e.g. `openssl rand -hex 32`) and keep using the same value; with a different one the saved keys can't be decrypted.

If you don't set `VSCHAT_ADMIN_PASSWORD`, a random one is printed in the log on first start (`docker logs vschat`).

### docker compose

```yaml
services:
  vschat:
    image: ghcr.io/vesaaa/vschat:latest
    restart: unless-stopped
    ports: ["8080:8080"]
    volumes: ["vschat-data:/data"]
    environment:
      VSCHAT_ADMIN_PASSWORD: change-me
      # Optional: pre-fill keys on first start (can also be set in the admin UI)
      # OPENAI_API_KEY: sk-...
      # ANTHROPIC_API_KEY: sk-ant-...
      # DEEPSEEK_API_KEY: sk-...
volumes:
  vschat-data:
```

## Configuration

| Variable | Default | |
|---|---|---|
| `VSCHAT_ADDR` | `:8080` | Listen address |
| `VSCHAT_DATA` | `/data` (Docker), `./data` | Where `vschat.db` lives — back up this one file |
| `VSCHAT_ADMIN_USER` | `admin` | First admin, created only when the database is empty |
| `VSCHAT_ADMIN_PASSWORD` | random | Password for that admin |
| `OPENAI_API_KEY` / `ANTHROPIC_API_KEY` / `DEEPSEEK_API_KEY` | | Seed provider keys on first start only |
| `VSCHAT_SECRET` | | Encrypts stored API keys (AES-GCM). Set it before adding keys; existing plain keys are encrypted on start. Keep it safe — keys can't be read without it |
| `VSCHAT_TRUST_PROXY` | | Set to `1` behind a reverse proxy so login rate-limiting uses `X-Forwarded-For` |

Everything else (providers, models, users, limits) is managed in the admin panel.

### Providers

- **OpenAI-compatible** — OpenAI, DeepSeek, and anything else speaking `/chat/completions` (set its base URL, e.g. `https://openrouter.ai/api/v1`, `http://ollama:11434/v1`).
- **Anthropic** — Claude via the official Go SDK. Thinking summaries are shown in a collapsible "Thoughts" block, and Claude's turns are replayed unchanged so multi-turn conversations keep their reasoning and prompt cache. On `api.anthropic.com`, newer models opt into server-side refusal fallback.

Use **Fetch from provider** in the provider editor to pick models from the live list.

### Who can use what

- **Shared providers** (Admin → Providers) are visible to *everyone* or to *selected users* — e.g. a "Claude Opus" provider only for a few people. Add the same API several times to split models into groups.
- **Personal providers** — turn on *Admin → Settings → Let users add their own API keys*. Users then get **My models** in their menu. Those providers are private to the user, show up as "Mine" in the model picker, and their tokens are tracked separately (not counted against the monthly limit).
- Personal providers can't reach `localhost` or private/LAN addresses by default, so users can't use the server to probe your network. Turn on *Let personal providers reach private networks* if they need e.g. a LAN Ollama.

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

MIT
