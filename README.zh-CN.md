# VsChat

[English](README.md) | **简体中文**

一个小巧的自托管多用户 AI 对话。选服务商、填 Key，就能用 **ChatGPT、Claude、Gemini、Grok、DeepSeek、通义千问、Kimi、智谱 GLM**（或任何 OpenAI 兼容接口），每个人有自己的账号和对话记录。

- **小巧**：一个 Go 程序加一个 SQLite 文件。镜像约 25 MB，空闲时只占几 MB 内存，使用中约 25 MB。不需要 Node、Python，也不需要数据库服务。
- **多用户**：管理员建账号，各人的对话互相隔离。可以给每人设每月 token 额度（新用户自动套用默认值），用到 80% 时提醒，用完即停。后台有用量统计。
- **管理可见**：访问记录包括登录、登录失败和每次对话，带真实 IP、国家和设备信息，经过 Cloudflare 也能拿到真实 IP。可以查看每个用户的最近活跃时间，以及各服务商的限流余量和重置时间。DeepSeek、Kimi 还能直接查询账户余额。
- **公共模型 + 个人模型**：管理员决定每个公共服务对哪些用户可见。用户也可以添加自己的 Key（需管理员开启），只对本人可见，用量不计入额度。
- **内置服务商**：选服务商、填 Key 即可。主流模型已经内置好，随版本自动更新。Claude 走官方 SDK，会显示思考过程；DeepSeek、千问、Kimi、GLM 的思考过程也会显示。其他 OpenAI 兼容接口（Ollama、OpenRouter、vLLM 等）可以作为自定义服务接入。同一个对话里可以随时切换模型。
- **好用**：流式输出，模型思考时实时读秒，支持 Markdown 和代码高亮，有深色和浅色主题，手机上也能用，界面有中文和英文。
- **长文本自动转附件**：粘贴一大段代码或日志时，会自动变成一张文件卡片，不会把输入框塞满。也可以上传或直接拖入文本文件。

## 快速开始（Docker）

```bash
docker run -d --name vschat -p 8080:8080 \
  -v vschat-data:/data \
  -e VSCHAT_ADMIN_PASSWORD=change-me \
  -e VSCHAT_SECRET=a-long-random-string-you-keep \
  vsbox/vschat:latest
```

同一个镜像在 GitHub 上也有：`ghcr.io/vesaaa/vschat`。可用的标签有：`latest`（最新正式版）、`0.3.0` 这样的版本号、`edge`（main 分支最新构建）。

打开 http://localhost:8080，用 `admin` 登录，在「管理后台 → 模型服务」里填入 API Key 就能用了。

`VSCHAT_SECRET` 用来加密保存 API Key。定一个值（比如用 `openssl rand -hex 32` 生成）后就不要再改：换了值，之前保存的 Key 就解不开了。

如果没设 `VSCHAT_ADMIN_PASSWORD`，首次启动时会生成一个随机密码打印在日志里，用 `docker logs vschat` 查看。

### docker compose

```yaml
services:
  vschat:
    image: vsbox/vschat:latest
    restart: unless-stopped
    ports: ["8080:8080"]
    volumes:
      - ./data:/data            # VsChat 的所有数据都在这里
    environment:
      VSCHAT_ADMIN_PASSWORD: change-me
      VSCHAT_SECRET: a-long-random-string-you-keep
      # PUID: 1000              # 以宿主机用户身份运行，./data 里的文件归你自己
      # PGID: 1000
      # 可选：首次启动时预填 Key（也可以在管理后台里填）
      # OPENAI_API_KEY: sk-...
      # ANTHROPIC_API_KEY: sk-ant-...
      # DEEPSEEK_API_KEY: sk-...
```

也可以用命名卷（`vschat-data:/data`）。密钥类变量支持从文件读取，可以配合 Docker secrets 使用，例如 `VSCHAT_SECRET_FILE: /run/secrets/vschat_secret`。

## 配置

| 环境变量 | 默认值 | 说明 |
|---|---|---|
| `VSCHAT_ADDR` | `:8080` | 监听地址 |
| `VSCHAT_DATA` | `/data`（Docker）、`./data` | `vschat.db` 所在目录，备份它就够了 |
| `VSCHAT_ADMIN_USER` | `admin` | 首个管理员的用户名，只在数据库为空时创建 |
| `VSCHAT_ADMIN_PASSWORD` | 随机生成 | 首个管理员的密码 |
| `OPENAI_API_KEY` `ANTHROPIC_API_KEY` `GEMINI_API_KEY` `XAI_API_KEY` `DEEPSEEK_API_KEY` `DASHSCOPE_API_KEY` `MOONSHOT_API_KEY` `ZHIPU_API_KEY` | | 首次启动时，每设置一个 Key 就自动添加对应的服务 |
| `VSCHAT_SECRET` | | 用 AES-GCM 加密保存 API Key。最好在填 Key 之前就设好；已有的明文 Key 会在启动时自动加密。务必保管好，丢了就读不出已保存的 Key |
| `PUID` / `PGID` | `10001` | Docker 中运行服务的用户。容器启动时会自动修正 `/data` 的权限，挂载宿主机目录也没问题 |
| `VSCHAT_TRUST_PROXY` | `auto` | 决定在什么情况下相信代理转发的真实 IP 请求头。`auto`：只在请求来自 Cloudflare、本机或内网（比如你的 nginx）时相信；`1`：总是相信（源站只能通过代理访问时用）；`0`：从不相信 |

`VSCHAT_SECRET`、`VSCHAT_ADMIN_PASSWORD` 和各个 `*_API_KEY` 都支持加 `_FILE` 后缀，表示从文件读取（用于 Docker 或 Kubernetes secrets）。

没有配置文件。除环境变量外，服务商、模型、用户、额度和设置都存在数据库里，在管理后台修改。只有 `/data` 一个目录需要持久化和备份。

### 服务商

内置服务商（见 [`catalog.go`](catalog.go)）。国内服务商默认用国内地址，海外服务商默认用官方地址：

| 服务商 | 地址 | 内置模型 |
|---|---|---|
| OpenAI | api.openai.com | gpt-6.1-sol、gpt-6-astra、gpt-6.1-terra、gpt-6-luna |
| Claude | api.anthropic.com（官方 SDK） | claude-opus-5-5、claude-sonnet-5-5、claude-haiku-5-5 |
| Gemini | generativelanguage.googleapis.com（OpenAI 兼容） | gemini-4-argon、gemini-3.8-flash、gemini-3.5-flash-lite |
| Grok | api.x.ai | grok-4.7、grok-4.6 |
| DeepSeek | api.deepseek.com | deepseek-v4-pro、deepseek-flash |
| 通义千问 Qwen | dashscope.aliyuncs.com | qwen3.8-max、qwen3.7-plus、qwen3.8-flash |
| Kimi | api.moonshot.cn | kimi-k3、kimi-k2.6 |
| 智谱 GLM | open.bigmodel.cn | glm-5.3、glm-5.3-flash |

从列表里添加的服务**默认跟随内置模型列表**，VsChat 发新版时会自动带上新模型，不用手动更新。取消勾选后，可以自己挑模型、补充其他模型 ID，或者点「从服务商获取」。API 地址可以在「高级」里修改，比如 Kimi 改成海外的 `api.moonshot.ai`。**自定义**服务支持任何 OpenAI 兼容或 Anthropic 兼容的接口。

### 谁能用哪些模型

- **公共服务**（管理后台 → 模型服务）：可见范围选「所有用户」或「指定用户」，比如单独建一个只给少数人用的「Claude Opus」服务。同一家 API 可以添加多次，把模型分成不同的组。
- **个人服务**：在「管理后台 → 设置 → 用户自己的 API Key」里选：
  - **不允许**（默认）：用户只能用公共模型。
  - **仅官方服务**：用户从内置服务商里选一家，填自己的 Key，API 地址固定为官方地址。
  - **允许自定义 API 地址**：用户还可以填任意公网的 OpenAI 兼容或 Anthropic 接口，比如中转站、OpenRouter。

  开启后，用户菜单里会出现「我的模型」。个人服务只有本人能看到，在模型选择器里归到「我的」分组下，用量单独统计，不计入每月额度。从宽松改成严格时，不符合新规则的个人服务会保留，但会停用并标注出来。
- 自定义地址模式下，个人服务默认不能访问 `localhost` 和内网地址，防止用户借服务器探测你的内网。如果确实要接内网的 Ollama 之类的服务，可以打开「允许个人服务访问内网地址」。

## Cloudflare Tunnel（不需要 nginx）

对外提供服务最简单的方式：由 `cloudflared` 主动连接 Cloudflare，再把请求转给 VsChat，服务器上不用开放任何端口。流式输出在 Tunnel 里可以正常工作，访客的真实 IP 和国家会自动识别。

```yaml
services:
  vschat:
    image: vsbox/vschat:latest
    restart: unless-stopped
    volumes: ["./data:/data"]
    environment:
      VSCHAT_ADMIN_PASSWORD: change-me
      VSCHAT_SECRET: a-long-random-string-you-keep
    # 不写 ports，只让 cloudflared 访问它
  cloudflared:
    image: cloudflare/cloudflared:latest
    restart: unless-stopped
    command: tunnel --no-autoupdate run
    environment:
      TUNNEL_TOKEN: your-tunnel-token   # Zero Trust → Networks → Tunnels
```

在 Tunnel 的 *Public Hostname* 设置里，把你的域名指向 `http://vschat:8080`。如果 cloudflared 直接装在宿主机上，就把端口只绑定到本机（`127.0.0.1:8080:8080`），再指向 `http://localhost:8080`。

静态文件已经预先 gzip 压缩，其余内容 Cloudflare 发给访客时会自动压缩。

## 经过 Cloudflare 或反向代理时的真实 IP

VsChat 会在访问记录里记录访客 IP，登录失败次数限制也按 IP 计算。默认 `VSCHAT_TRUST_PROXY=auto` 时，依次读取 `CF-Connecting-IP`、`True-Client-IP`、`X-Forwarded-For` 的第一跳、`X-Real-IP`。但只有当请求来自 Cloudflare 节点、本机或内网时才相信这些请求头，所以直接访问端口的人无法伪造 IP。经过 Cloudflare 时还会记录访客所在国家（`CF-IPCountry`）。

## 使用反向代理

流式输出用的是 SSE（服务器推送事件），需要关闭代理的响应缓冲。nginx 配置示例：

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

不需要配置缓存规则：
- 脚本和样式表的地址带内容指纹（`/s/<hash>/app.js`），设置为长期缓存（`immutable`）。
- 页面本身每次都会向服务器确认是否有更新。
- 接口一律不缓存（`no-store`）。

所以升级后，浏览器和 Cloudflare 下次加载时都会拿到新版本。

如果从公网访问，请使用 HTTPS。在 HTTPS 下，登录 Cookie 会自动加上 `Secure` 标记。

## 从源码构建

```bash
go build -o vschat .
./vschat            # http://localhost:8080
```

需要 Go 1.24 及以上版本。网页界面是原生 HTML/CSS/JS，直接打包在程序里，不需要额外的构建步骤。

## 说明

- API Key 不会发到浏览器。它保存在 SQLite 文件里，设置了 `VSCHAT_SECRET` 时会加密存储。无论是否加密，都请保护好数据目录。
- 每个用户的用量按服务商返回的 token 数统计。
- 附件目前只支持文本（代码、日志、Markdown、CSV 等），每个不超过 1 MB。

## 许可证

Copyright (C) 2026 vesaaa

从 v0.3.5 起，VsChat 以 [GNU Affero 通用公共许可证 v3.0](LICENSE)（AGPL-3.0）发布：
- **可以做的**：使用、修改、再分发。
- **需要遵守的**：如果把**修改过的版本**作为网络服务提供给别人使用，必须向这些用户提供对应的源代码（AGPL 第 13 条）。用户菜单里的「源代码」链接就是为此准备的，修改后请把它改成指向你自己的仓库。
- v0.3.4 及之前的版本以 MIT 协议发布。
