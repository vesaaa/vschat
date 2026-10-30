'use strict';

/* ================= i18n ================= */
const I18N = {
  zh: {
    loginSub: '登录以继续', username: '用户名', password: '密码', signIn: '登录',
    newChat: '新对话', search: '搜索对话', placeholder: '输入消息…',
    hint: 'Enter 发送 · Shift + Enter 换行', attach: '添加文件', dropHint: '松开以添加文件',
    today: '今天', yesterday: '昨天', week: '近 7 天', earlier: '更早', noChats: '还没有对话',
    welcome: name => `${name}，今天想聊点什么？`, welcomeNoModel: '还没有可用的模型',
    welcomeNoModelAdmin: '还没有可用的模型，请在「管理后台 → 模型服务」中填入 API Key',
    noModel: '未配置模型', pickModel: '请先选择模型', searchModel: '搜索模型',
    copy: '复制', copied: '已复制', regenerate: '重新生成', rename: '重命名', delete: '删除',
    thinking: '思考过程', thinkingNow: '思考中', thoughtFor: d => `思考了 ${d}`, retry: '重试',
    secs: n => n < 60 ? `${n} 秒` : `${Math.floor(n / 60)} 分 ${n % 60} 秒`,
    stopped: '已停止生成', refusal: '模型拒绝回答了这个请求', maxTokens: '回答达到长度上限，已被截断',
    errorPrefix: '请求失败：',
    pastedText: '粘贴的文本', lines: n => `${n} 行`, preview: '预览', insertBack: '放回输入框',
    binaryFile: name => `暂不支持该文件类型：${name}`, fileTooBig: name => `文件过大（上限 1 MB）：${name}`,
    tooMany: '附件最多 10 个', imageNotSupported: '暂不支持粘贴图片',
    renameChat: '重命名对话', deleteChat: '删除这个对话？', deleteChatSub: '删除后无法恢复。',
    cancel: '取消', save: '保存', confirm: '确定', close: '关闭', back: '返回',
    theme: '主题', themeSystem: '系统', themeLight: '浅色', themeDark: '深色', language: '语言',
    changePassword: '修改密码', admin: '管理后台', logout: '退出登录',
    oldPassword: '当前密码', newPassword: '新密码', newPassword2: '确认新密码', pwMismatch: '两次输入的密码不一致',
    pwChanged: '密码已修改', monthUsage: '本月用量',
    users: '用户', providers: '模型服务', usage: '用量',
    addUser: '添加用户', editUser: '编辑用户', role: '角色', status: '状态', adminRole: '管理员', userRole: '用户',
    active: '正常', disabled: '已停用', limit: '每月 token 上限', limitHint: '0 表示不限', unlimited: '不限',
    pwKeep: '留空则不修改', isAdmin: '设为管理员', isDisabled: '停用账号', deleteUser: name => `删除用户「${name}」？`,
    deleteUserSub: '该用户的所有对话将一并删除。',
    addProvider: '添加服务', editProvider: '编辑服务', name: '名称', kind: '接口类型',
    kindOpenAI: 'OpenAI 兼容', kindAnthropic: 'Anthropic (Claude)', baseURL: 'API 地址', apiKey: 'API Key',
    keySet: hint => `已设置 ${hint}，留空保持不变`, keyNotSet: '未设置', models: '模型列表', modelsHint: '每行一个模型 ID',
    fetchModels: '从服务商获取', fetching: '获取中…', enabled: '启用', deleteProvider: name => `删除服务「${name}」？`,
    modelCount: n => `${n} 个模型`, noKey: '缺少 Key', saved: '已保存',
    month: '月份', requests: '请求数', input: '输入', output: '输出', total: '合计', ownKeys: '自带 Key',
    quotaWarn: (pct, d) => `本月额度已用 ${pct}%，${d} 重置`, quotaOut: d => `本月额度已用完，${d} 恢复`, quotaOwnOk: '使用自己的 Key 不受影响',
    remaining: n => `剩余 ${n}`, resetOn: d => `${d} 重置`, lastActive: '最近活跃', never: '从未登录',
    ago: (n, u) => `${n} ${u}前`, unitS: '秒', unitM: '分钟', unitH: '小时', unitD: '天', justNow: '刚刚',
    access: '访问记录', allUsers: '全部用户', time: '时间', event: '事件', ipAddr: 'IP', device: '设备', detailCol: '详情', loadMore: '加载更多',
    evLogin: '登录', evLoginFailed: '登录失败', evChat: '对话',
    rateReq: '请求余量', rateTok: 'Token 余量', resetIn: s => `${s}后重置`, updated: s => `更新于 ${s}`, noRateYet: '发起对话后显示限流余量',
    checkBalance: '查询余额', balance: '余额', defaultLimit: '新用户默认每月 token 上限', defaultLimitSub: '新建用户时自动填入，0 表示不限。',
    visibility: '谁可以使用', everyone: '所有用户', selectedUsers: '指定用户', usersCount: n => `${n} 位用户`,
    pickUsers: '勾选可以使用的用户', settings: '设置', mine: '我的',
    myModels: '我的模型', myModelsSub: '添加你自己的 API Key，只有你能看到和使用，用量不计入额度。', noOwn: '还没有添加自己的模型服务',
    keyMode: '用户自己的 API Key', keyModeSub: '用户可以在「我的模型」中接入自己的服务，只对本人可见，用量不计入额度。',
    modeOff: '不允许', modeOfficial: '仅官方服务（OpenAI / Claude / DeepSeek）', modeCustom: '允许自定义 API 地址',
    providerPreset: '服务商', customOpenAI: '自定义（OpenAI 兼容接口）', customAnthropic: '自定义（Anthropic 接口）',
    getKey: '获取 API Key', autoModels: '使用内置模型列表（随版本自动更新）', moreModels: '其他模型', moreModelsHint: '每行一个，可选',
    advanced: '高级', cnRegion: '国内', pickOneModel: '请至少选择一个模型', blockedMode: '当前设置下不可用', privateNeedsCustom: '仅在「允许自定义 API 地址」时生效。',
    allowPrivate: '允许个人服务访问内网地址', allowPrivateSub: '默认禁止，防止用户借服务器访问本机或内网。需要接入内网的 Ollama 等服务时再开启。',
    keysEncrypted: 'API Key 已加密存储（VSCHAT_SECRET 已设置）', keysPlain: 'API Key 以明文存储。设置环境变量 VSCHAT_SECRET 可加密保存。',
  },
  en: {
    loginSub: 'Sign in to continue', username: 'Username', password: 'Password', signIn: 'Sign in',
    newChat: 'New chat', search: 'Search chats', placeholder: 'Message…',
    hint: 'Enter to send · Shift + Enter for new line', attach: 'Add files', dropHint: 'Drop files to attach',
    today: 'Today', yesterday: 'Yesterday', week: 'Previous 7 days', earlier: 'Earlier', noChats: 'No chats yet',
    welcome: name => `What's on your mind, ${name}?`, welcomeNoModel: 'No models available yet',
    welcomeNoModelAdmin: 'No models yet — add an API key in Admin → Providers',
    noModel: 'No model', pickModel: 'Pick a model first', searchModel: 'Search models',
    copy: 'Copy', copied: 'Copied', regenerate: 'Regenerate', rename: 'Rename', delete: 'Delete',
    thinking: 'Thoughts', thinkingNow: 'Thinking', thoughtFor: d => `Thought for ${d}`, retry: 'Retry',
    secs: n => n < 60 ? `${n}s` : `${Math.floor(n / 60)}m ${n % 60}s`,
    stopped: 'Generation stopped', refusal: 'The model declined this request', maxTokens: 'Reply hit the length limit and was cut off',
    errorPrefix: 'Request failed: ',
    pastedText: 'Pasted text', lines: n => `${n} lines`, preview: 'Preview', insertBack: 'Put back in message',
    binaryFile: name => `Unsupported file type: ${name}`, fileTooBig: name => `File too large (1 MB max): ${name}`,
    tooMany: 'Up to 10 attachments', imageNotSupported: 'Pasting images is not supported yet',
    renameChat: 'Rename chat', deleteChat: 'Delete this chat?', deleteChatSub: 'This cannot be undone.',
    cancel: 'Cancel', save: 'Save', confirm: 'OK', close: 'Close', back: 'Back',
    theme: 'Theme', themeSystem: 'Auto', themeLight: 'Light', themeDark: 'Dark', language: 'Language',
    changePassword: 'Change password', admin: 'Admin', logout: 'Sign out',
    oldPassword: 'Current password', newPassword: 'New password', newPassword2: 'Confirm new password', pwMismatch: 'Passwords do not match',
    pwChanged: 'Password changed', monthUsage: 'This month',
    users: 'Users', providers: 'Providers', usage: 'Usage',
    addUser: 'Add user', editUser: 'Edit user', role: 'Role', status: 'Status', adminRole: 'Admin', userRole: 'User',
    active: 'Active', disabled: 'Disabled', limit: 'Monthly token limit', limitHint: '0 means unlimited', unlimited: 'Unlimited',
    pwKeep: 'Leave empty to keep', isAdmin: 'Administrator', isDisabled: 'Disable account', deleteUser: name => `Delete user "${name}"?`,
    deleteUserSub: 'All of their chats will be deleted too.',
    addProvider: 'Add provider', editProvider: 'Edit provider', name: 'Name', kind: 'API type',
    kindOpenAI: 'OpenAI-compatible', kindAnthropic: 'Anthropic (Claude)', baseURL: 'Base URL', apiKey: 'API key',
    keySet: hint => `Set (${hint}) — leave empty to keep`, keyNotSet: 'Not set', models: 'Models', modelsHint: 'One model ID per line',
    fetchModels: 'Fetch from provider', fetching: 'Fetching…', enabled: 'Enabled', deleteProvider: name => `Delete provider "${name}"?`,
    modelCount: n => `${n} models`, noKey: 'No key', saved: 'Saved',
    month: 'Month', requests: 'Requests', input: 'Input', output: 'Output', total: 'Total', ownKeys: 'Own keys',
    quotaWarn: (pct, d) => `${pct}% of this month's quota used · resets ${d}`, quotaOut: d => `Monthly quota used up · resets ${d}`, quotaOwnOk: 'Your own keys still work',
    remaining: n => `${n} left`, resetOn: d => `resets ${d}`, lastActive: 'Last active', never: 'Never',
    ago: (n, u) => `${n}${u} ago`, unitS: 's', unitM: 'm', unitH: 'h', unitD: 'd', justNow: 'just now',
    access: 'Access log', allUsers: 'All users', time: 'Time', event: 'Event', ipAddr: 'IP', device: 'Device', detailCol: 'Details', loadMore: 'Load more',
    evLogin: 'Sign in', evLoginFailed: 'Failed sign-in', evChat: 'Chat',
    rateReq: 'Requests left', rateTok: 'Tokens left', resetIn: s => `resets in ${s}`, updated: s => `updated ${s}`, noRateYet: 'Rate limits show up after the first chat',
    checkBalance: 'Check balance', balance: 'Balance', defaultLimit: 'Default monthly token limit for new users', defaultLimitSub: 'Filled in when creating a user. 0 means unlimited.',
    visibility: 'Who can use it', everyone: 'Everyone', selectedUsers: 'Selected users', usersCount: n => `${n} users`,
    pickUsers: 'Tick the users who may use it', settings: 'Settings', mine: 'Mine',
    myModels: 'My models', myModelsSub: 'Add your own API keys. Only you can see and use them, and they don’t count toward your limit.', noOwn: 'No personal providers yet',
    keyMode: 'Users’ own API keys', keyModeSub: 'Users can connect their own providers under “My models”, visible only to them and not counted against their limit.',
    modeOff: 'Not allowed', modeOfficial: 'Official providers only (OpenAI / Claude / DeepSeek)', modeCustom: 'Allow custom API URLs',
    providerPreset: 'Provider', customOpenAI: 'Custom (OpenAI-compatible)', customAnthropic: 'Custom (Anthropic API)',
    getKey: 'Get an API key', autoModels: 'Use the built-in model list (updates with new releases)', moreModels: 'More models', moreModelsHint: 'One per line, optional',
    advanced: 'Advanced', cnRegion: 'China', pickOneModel: 'Pick at least one model', blockedMode: 'Not allowed by current settings', privateNeedsCustom: 'Only applies when custom API URLs are allowed.',
    allowPrivate: 'Let personal providers reach private networks', allowPrivateSub: 'Off by default so users can’t use the server to reach localhost or your LAN. Turn on for a LAN Ollama and the like.',
    keysEncrypted: 'API keys are encrypted at rest (VSCHAT_SECRET is set)', keysPlain: 'API keys are stored in plain text. Set VSCHAT_SECRET to encrypt them.',
  },
};
const store = {
  get(k, d) { try { return localStorage.getItem('vschat.' + k) ?? d; } catch { return d; } },
  set(k, v) { try { localStorage.setItem('vschat.' + k, v); } catch {} },
};
let lang = store.get('lang', navigator.language?.startsWith('zh') ? 'zh' : 'en');
function t(key, ...args) {
  const v = (I18N[lang] || I18N.zh)[key] ?? I18N.zh[key] ?? key;
  return typeof v === 'function' ? v(...args) : v;
}

/* ================= icons ================= */
const ICONS = {
  plus: '<path d="M12 5v14M5 12h14"/>',
  search: '<circle cx="11" cy="11" r="7"/><path d="m20 20-3.5-3.5"/>',
  x: '<path d="M18 6 6 18M6 6l12 12"/>',
  menu: '<path d="M4 6h16M4 12h16M4 18h16"/>',
  more: '<path d="M5 12h.01M12 12h.01M19 12h.01" stroke-width="3"/>',
  chevron: '<path d="m6 9 6 6 6-6"/>',
  right: '<path d="m9 6 6 6-6 6"/>',
  edit: '<path d="M12 20h9"/><path d="M16.5 3.5a2.1 2.1 0 0 1 3 3L7 19l-4 1 1-4z"/>',
  down: '<path d="M12 5v14M5 12l7 7 7-7"/>',
  clip: '<path d="m21.4 11.6-8.9 8.9a5.5 5.5 0 0 1-7.8-7.8l8.9-8.9a3.7 3.7 0 0 1 5.2 5.2l-8.9 8.9a1.8 1.8 0 0 1-2.6-2.6l8.2-8.2"/>',
  send: '<path d="M12 19V5M5 12l7-7 7 7"/>',
  stop: '<rect x="6" y="6" width="12" height="12" rx="2.5"/>',
  copy: '<rect x="9" y="9" width="11" height="11" rx="2"/><path d="M5 15H4a1 1 0 0 1-1-1V4a1 1 0 0 1 1-1h10a1 1 0 0 1 1 1v1"/>',
  check: '<path d="M20 6 9 17l-5-5"/>',
  refresh: '<path d="M3 12a9 9 0 0 1 15.5-6.2L21 8"/><path d="M21 3v5h-5"/><path d="M21 12a9 9 0 0 1-15.5 6.2L3 16"/><path d="M3 21v-5h5"/>',
  trash: '<path d="M3 6h18M8 6V4h8v2M19 6l-1 14H6L5 6"/>',
  file: '<path d="M14 3H6a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V9z"/><path d="M14 3v6h6"/>',
  paste: '<rect x="8" y="2" width="8" height="4" rx="1"/><path d="M16 4h2a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2h2M8 12h8M8 16h5"/>',
  sun: '<circle cx="12" cy="12" r="4"/><path d="M12 2v2M12 20v2M4.9 4.9l1.4 1.4M17.7 17.7l1.4 1.4M2 12h2M20 12h2M4.9 19.1l1.4-1.4M17.7 6.3l1.4-1.4"/>',
  globe: '<circle cx="12" cy="12" r="9"/><path d="M3 12h18M12 3a14 14 0 0 1 0 18M12 3a14 14 0 0 0 0 18"/>',
  key: '<circle cx="7.5" cy="15.5" r="3.5"/><path d="m10 13 9-9M16 7l3 3M14 9l2 2"/>',
  shield: '<path d="M12 3l8 3v6c0 5-3.5 8-8 9-4.5-1-8-4-8-9V6z"/>',
  logout: '<path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4M16 17l5-5-5-5M21 12H9"/>',
  spark: '<path d="M12 3l1.9 5.1L19 10l-5.1 1.9L12 17l-1.9-5.1L5 10l5.1-1.9z"/><path d="M19 15l.8 2.2L22 18l-2.2.8L19 21l-.8-2.2L16 18l2.2-.8z"/>',
  alert: '<circle cx="12" cy="12" r="9"/><path d="M12 8v4M12 16h.01"/>',
  clock: '<circle cx="12" cy="12" r="9"/><path d="M12 7v5l3 2"/>',
  cube: '<path d="M21 8l-9-5-9 5v8l9 5 9-5z"/><path d="M3 8l9 5 9-5M12 13v8"/>',
  sliders: '<path d="M4 6h10M18 6h2M4 12h4M12 12h8M4 18h12M20 18h0"/><circle cx="16" cy="6" r="2"/><circle cx="10" cy="12" r="2"/><circle cx="18" cy="18" r="2"/>',
  bulb: '<path d="M9 18h6M10 21h4M12 3a6 6 0 0 0-4 10.5c.7.7 1 1.5 1 2.5h6c0-1 .3-1.8 1-2.5A6 6 0 0 0 12 3z"/>',
};
function icon(name) {
  return `<svg class="i" viewBox="0 0 24 24" aria-hidden="true">${ICONS[name] || ''}</svg>`;
}

/* ================= helpers ================= */
const $ = sel => document.querySelector(sel);
const esc = s => String(s ?? '').replace(/[&<>"']/g, c => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c]));
function h(html) {
  const tpl = document.createElement('template');
  tpl.innerHTML = html.trim();
  return tpl.content.firstElementChild;
}
function fmtTokens(n) {
  n = n || 0;
  if (n >= 1e6) return (n / 1e6).toFixed(n >= 1e7 ? 1 : 2).replace(/\.?0+$/, '') + 'M';
  if (n >= 1e3) return (n / 1e3).toFixed(n >= 1e4 ? 0 : 1).replace(/\.0$/, '') + 'K';
  return String(n);
}
function fmtSize(n) { return n >= 1024 * 1024 ? (n / 1048576).toFixed(1) + ' MB' : n >= 1024 ? Math.round(n / 1024) + ' KB' : n + ' B'; }
function lineCount(s) { return s ? s.split('\n').length : 0; }
const fmtDate = ts => new Date(ts * 1000).toLocaleDateString(lang === 'zh' ? 'zh-CN' : 'en', { month: 'short', day: 'numeric' });
const fmtDateTime = ts => new Date(ts * 1000).toLocaleString(lang === 'zh' ? 'zh-CN' : 'en', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit', second: '2-digit' });
function fmtSpan(sec) {
  sec = Math.max(0, Math.round(sec));
  if (sec < 60) return `${sec}${lang === 'zh' ? ' 秒' : 's'}`;
  if (sec < 3600) return `${Math.round(sec / 60)}${lang === 'zh' ? ' 分钟' : 'm'}`;
  if (sec < 86400) return `${Math.round(sec / 3600)}${lang === 'zh' ? ' 小时' : 'h'}`;
  return `${Math.round(sec / 86400)}${lang === 'zh' ? ' 天' : 'd'}`;
}
function fmtAgo(ts) {
  if (!ts) return t('never');
  const d = Date.now() / 1000 - ts;
  return d < 30 ? t('justNow') : t('ago', ...(d < 60 ? [Math.round(d), t('unitS')] : d < 3600 ? [Math.round(d / 60), t('unitM')] : d < 86400 ? [Math.round(d / 3600), t('unitH')] : [Math.round(d / 86400), t('unitD')]));
}
// "Chrome · Windows" from a user agent string.
function uaSummary(ua) {
  if (!ua) return '';
  const b = /Edg\//.test(ua) ? 'Edge' : /OPR\//.test(ua) ? 'Opera' : /Firefox\//.test(ua) ? 'Firefox' : /Chrome\//.test(ua) ? 'Chrome' : /Safari\//.test(ua) ? 'Safari' : /curl|python|Go-http|okhttp/i.test(ua) ? ua.split('/')[0] : '';
  const o = /iPhone|iPad/.test(ua) ? 'iOS' : /Android/.test(ua) ? 'Android' : /Windows/.test(ua) ? 'Windows' : /Mac OS X/.test(ua) ? 'macOS' : /Linux/.test(ua) ? 'Linux' : '';
  return [b, o].filter(Boolean).join(' · ') || ua.slice(0, 40);
}
const flag = cc => cc ? String.fromCodePoint(...[...cc].map(c => 0x1f1a5 + c.charCodeAt(0))) + ' ' : '';

function toast(msg, err = false) {
  const el = h(`<div class="toast${err ? ' err' : ''}">${esc(msg)}</div>`);
  $('#toasts').append(el);
  setTimeout(() => el.remove(), err ? 4200 : 2200);
}

async function copyText(text) {
  try {
    await navigator.clipboard.writeText(text);
  } catch {
    const ta = h('<textarea style="position:fixed;opacity:0"></textarea>');
    ta.value = text; document.body.append(ta); ta.select();
    document.execCommand('copy'); ta.remove();
  }
}

class HttpError extends Error { constructor(status, msg) { super(msg); this.status = status; } }
async function api(method, path, body) {
  const res = await fetch('/api' + path, {
    method, credentials: 'same-origin',
    headers: body !== undefined ? { 'Content-Type': 'application/json' } : {},
    body: body !== undefined ? JSON.stringify(body) : undefined,
  });
  if (res.status === 401 && path !== '/login') { showLogin(); throw new HttpError(401, 'unauthorized'); }
  if (!res.ok) {
    let msg = res.statusText;
    try { msg = (await res.json()).error || msg; } catch {}
    throw new HttpError(res.status, msg);
  }
  return res.status === 204 ? null : res.json();
}

function applyI18n(root = document) {
  root.querySelectorAll('[data-i18n]').forEach(el => { el.textContent = t(el.dataset.i18n); });
  root.querySelectorAll('[data-i18n-ph]').forEach(el => { el.placeholder = t(el.dataset.i18nPh); });
  root.querySelectorAll('[data-i18n-title]').forEach(el => { el.title = t(el.dataset.i18nTitle); });
  document.documentElement.lang = lang === 'zh' ? 'zh-CN' : 'en';
}
function applyIcons(root = document) {
  root.querySelectorAll('[data-icon]').forEach(el => {
    if (el.dataset.iconDone) return;
    el.insertAdjacentHTML('afterbegin', icon(el.dataset.icon));
    el.dataset.iconDone = '1';
  });
}

/* ================= theme ================= */
function setTheme(mode) {
  store.set('theme', mode);
  if (mode === 'light' || mode === 'dark') document.documentElement.dataset.theme = mode;
  else delete document.documentElement.dataset.theme;
}

/* ================= markdown ================= */
marked.setOptions({ gfm: true, breaks: true });
DOMPurify.addHook('afterSanitizeAttributes', node => {
  if (node.tagName === 'A') { node.setAttribute('target', '_blank'); node.setAttribute('rel', 'noopener noreferrer'); }
});

function renderMarkdown(el, text, live) {
  el.innerHTML = DOMPurify.sanitize(marked.parse(text || ''));
  el.querySelectorAll('pre > code').forEach(code => {
    const pre = code.parentElement;
    const lang = (code.className.match(/language-([\w+#-]+)/) || [])[1] || '';
    if (lang && hljs.getLanguage(lang)) {
      code.innerHTML = hljs.highlight(code.textContent, { language: lang, ignoreIllegals: true }).value;
    } else if (!live && code.textContent.length < 20000) {
      hljs.highlightElement(code);
    }
    const wrap = h(`<div class="code"><div class="code-head"><span>${esc(lang || 'text')}</span>
      <button type="button">${icon('copy')}<span>${t('copy')}</span></button></div></div>`);
    pre.replaceWith(wrap);
    wrap.append(pre);
    wrap.querySelector('button').onclick = async e => {
      await copyText(code.textContent);
      const b = e.currentTarget; b.innerHTML = `${icon('check')}<span>${t('copied')}</span>`;
      setTimeout(() => { b.innerHTML = `${icon('copy')}<span>${t('copy')}</span>`; }, 1500);
    };
  });
}

/* ================= state ================= */
const S = {
  me: null, monthTokens: 0, models: [], chats: [], chatId: null, messages: [],
  model: store.get('model', ''), attachments: [], controller: null, search: '', userKeys: 'off', catalog: [],
};
const modelById = id => S.models.find(m => m.id === id);

/* ================= auth ================= */
function showLogin() {
  $('#app').hidden = true;
  $('#login').hidden = false;
  $('#login-form [name=username]').focus();
}

$('#login-form').addEventListener('submit', async e => {
  e.preventDefault();
  const f = e.currentTarget;
  $('#login-error').textContent = '';
  try {
    await api('POST', '/login', { username: f.username.value, password: f.password.value });
    f.password.value = '';
    await boot();
  } catch (err) {
    $('#login-error').textContent = err.message;
  }
});

async function boot() {
  let me;
  try { me = await api('GET', '/me'); } catch { return; }
  S.me = me.user; S.monthTokens = me.month_tokens; S.resetAt = me.reset_at; S.userKeys = me.user_keys;
  $('#login').hidden = true;
  $('#app').hidden = false;
  $('#user-name').textContent = S.me.username;
  $('#user-avatar').textContent = S.me.username.slice(0, 1).toUpperCase();
  await Promise.all([loadModels(), loadChats(), api('GET', '/catalog').then(c => { S.catalog = c; })]);
  const m = location.hash.match(/^#\/c\/(\d+)/);
  if (m) await openChat(+m[1]); else newChat();
}

/* ================= models ================= */
async function loadModels() {
  S.models = await api('GET', '/models');
  if (!modelById(S.model)) S.model = S.models[0]?.id || '';
  renderModelLabel();
  renderQuota();
}

function renderModelLabel() {
  const m = modelById(S.model);
  $('#model-label').innerHTML = m ? `${esc(m.model)} <span class="sub">· ${esc(m.own ? t('mine') + ' · ' + m.provider : m.provider)}</span>` : esc(t('noModel'));
}

// A slim bar above the composer once 80% of the monthly quota is used. Only
// shared models count, so it hides while a model on the user's own key is picked.
function renderQuota() {
  const box = $('#quota');
  const limit = S.me?.token_limit || 0;
  const pct = limit ? Math.floor(S.monthTokens / limit * 100) : 0;
  if (!limit || pct < 80 || modelById(S.model)?.own) { box.hidden = true; return; }
  const out = pct >= 100;
  box.hidden = false;
  box.className = 'quota' + (out ? ' out' : '');
  box.innerHTML = `<span class="quota-bar"><i style="width:${Math.min(100, pct)}%"></i></span>
    <span>${esc(out ? t('quotaOut', fmtDate(S.resetAt)) : t('quotaWarn', pct, fmtDate(S.resetAt)))}${out && S.userKeys !== 'off' ? ' · ' + esc(t('quotaOwnOk')) : ''}</span>`;
}

function setModel(id) {
  S.model = id;
  store.set('model', id);
  renderModelLabel();
  renderQuota();
  if (S.chatId) api('PATCH', `/chats/${S.chatId}`, { model: id }).catch(() => {});
}

const KIND_COLORS = { anthropic: '#d97757', openai: '#10a37f' };
function providerColor(name, kind, preset) {
  const pr = S.catalog.find(c => c.id === preset);
  if (pr) return pr.color;
  if (/deepseek/i.test(name)) return '#4d6bfe';
  return KIND_COLORS[kind] || '#6366f1';
}

function openModelMenu() {
  const btn = $('#model-btn');
  const render = q => {
    q = q.toLowerCase();
    let html = S.models.length > 8 ? `<input class="menu-search" placeholder="${esc(t('searchModel'))}" value="${esc(q)}">` : '';
    let group = null;
    for (const m of S.models) {
      if (q && !(m.model + ' ' + m.provider).toLowerCase().includes(q)) continue;
      const g = m.own ? `${t('mine')} · ${m.provider}` : m.provider;
      if (g !== group) { group = g; html += `<div class="menu-label">${esc(group)}</div>`; }
      html += `<button class="menu-item" data-id="${esc(m.id)}"><span class="kind-dot" style="background:${providerColor(m.provider, m.kind, m.preset)}"></span>
        <span>${esc(m.model)}</span>${m.id === S.model ? `<span class="check-mark">${icon('check')}</span>` : ''}</button>`;
    }
    if (!S.models.length) html += `<div class="empty-state">${esc(S.me.is_admin ? t('welcomeNoModelAdmin') : t('welcomeNoModel'))}</div>`;
    return html;
  };
  const pop = openPopover(btn, render(''), { minWidth: 260, above: true });
  pop.addEventListener('click', e => {
    const b = e.target.closest('[data-id]');
    if (b) { setModel(b.dataset.id); closePopover(); }
  });
  pop.addEventListener('input', e => {
    if (!e.target.matches('.menu-search')) return;
    const q = e.target.value;
    pop.innerHTML = render(q);
    const inp = pop.querySelector('.menu-search'); inp.focus(); inp.setSelectionRange(q.length, q.length);
  });
  pop.querySelector('.menu-search')?.focus();
}
$('#model-btn').onclick = openModelMenu;

/* ================= popover ================= */
let popAnchor = null;
function openPopover(anchor, html, { minWidth = 220, above = false } = {}) {
  const pop = $('#popover');
  const fresh = pop.cloneNode(false); // drop old listeners
  pop.replaceWith(fresh);
  fresh.innerHTML = html;
  fresh.style.minWidth = minWidth + 'px';
  fresh.hidden = false;
  const r = anchor.getBoundingClientRect();
  const pw = fresh.offsetWidth, ph = fresh.offsetHeight;
  let left = Math.min(Math.max(8, r.left), window.innerWidth - pw - 8);
  let top = above ? r.top - ph - 6 : r.bottom + 6;
  if (top + ph > window.innerHeight - 8) top = Math.max(8, r.top - ph - 6);
  fresh.style.left = left + 'px';
  fresh.style.top = top + 'px';
  popAnchor = anchor;
  return fresh;
}
function closePopover() { $('#popover').hidden = true; popAnchor = null; }
document.addEventListener('mousedown', e => {
  const pop = $('#popover');
  if (!pop.hidden && !pop.contains(e.target) && !(popAnchor && popAnchor.contains(e.target))) closePopover();
});

/* ================= modal ================= */
function openModal({ title, body, foot, wide = false, onClose }) {
  const dlg = h(`<dialog class="modal${wide ? ' wide' : ''}">
    <div class="modal-head"><h2></h2><button class="icon-btn" data-close>${icon('x')}</button></div>
    <div class="modal-body"></div></dialog>`);
  dlg.querySelector('h2').textContent = title;
  const bodyEl = dlg.querySelector('.modal-body');
  if (typeof body === 'string') bodyEl.innerHTML = body; else if (body) bodyEl.append(body);
  if (foot) {
    const f = h('<div class="modal-foot"></div>');
    foot.forEach(b => f.append(b));
    dlg.append(f);
  }
  dlg.querySelector('[data-close]').onclick = () => dlg.close();
  dlg.addEventListener('close', () => { dlg.remove(); onClose?.(); });
  dlg.addEventListener('mousedown', e => { if (e.target === dlg) dlg.close(); });
  document.body.append(dlg);
  dlg.showModal();
  return dlg;
}
function button(label, cls = '', onclick) {
  const b = h(`<button type="button" class="btn ${cls}"></button>`);
  b.textContent = label;
  if (onclick) b.onclick = onclick;
  return b;
}
function confirmDialog(title, sub, danger = true) {
  return new Promise(resolve => {
    let ok = false;
    const dlg = openModal({
      title, body: sub ? `<p style="margin:0;color:var(--text-2)">${esc(sub)}</p>` : '',
      foot: [button(t('cancel'), '', () => dlg.close()),
             button(danger ? t('delete') : t('confirm'), danger ? 'btn-primary' : 'btn-primary', () => { ok = true; dlg.close(); })],
      onClose: () => resolve(ok),
    });
    if (danger) dlg.querySelector('.btn-primary').style.cssText = 'background:var(--danger);border-color:var(--danger);color:#fff';
  });
}
function promptDialog(title, value) {
  return new Promise(resolve => {
    let result = null;
    const form = h(`<form><label class="field"><input required maxlength="100"></label></form>`);
    const inp = form.querySelector('input'); inp.value = value || '';
    const submit = () => { if (inp.value.trim()) { result = inp.value.trim(); dlg.close(); } };
    form.onsubmit = e => { e.preventDefault(); submit(); };
    const dlg = openModal({ title, body: form, foot: [button(t('cancel'), '', () => dlg.close()), button(t('save'), 'btn-primary', submit)], onClose: () => resolve(result) });
    inp.select();
  });
}

/* ================= chats sidebar ================= */
async function loadChats() {
  S.chats = await api('GET', '/chats');
  renderChats();
}

function chatGroup(ts) {
  const d = new Date(ts * 1000), n = new Date();
  const day = x => new Date(x.getFullYear(), x.getMonth(), x.getDate()).getTime();
  const diff = (day(n) - day(d)) / 864e5;
  return diff < 1 ? 'today' : diff < 2 ? 'yesterday' : diff < 7 ? 'week' : 'earlier';
}

function renderChats() {
  const list = $('#chat-list');
  const q = S.search.toLowerCase();
  const chats = S.chats.filter(c => !q || (c.title || '').toLowerCase().includes(q));
  if (!chats.length) { list.innerHTML = `<div class="chat-empty">${esc(t('noChats'))}</div>`; return; }
  let html = '', group = null;
  for (const c of chats) {
    const g = chatGroup(c.updated_at);
    if (g !== group) { group = g; html += `<div class="chat-group">${esc(t(g))}</div>`; }
    html += `<div class="chat-item${c.id === S.chatId ? ' active' : ''}" data-id="${c.id}">
      <span class="title">${esc(c.title || t('newChat'))}</span>
      <button class="icon-btn" data-menu>${icon('more')}</button></div>`;
  }
  list.innerHTML = html;
}

$('#chat-list').addEventListener('click', e => {
  const item = e.target.closest('.chat-item');
  if (!item) return;
  const id = +item.dataset.id;
  if (e.target.closest('[data-menu]')) {
    const pop = openPopover(e.target.closest('[data-menu]'), `
      <button class="menu-item" data-act="rename">${icon('edit')}<span>${esc(t('rename'))}</span></button>
      <button class="menu-item danger" data-act="delete">${icon('trash')}<span>${esc(t('delete'))}</span></button>`, { minWidth: 160 });
    pop.onclick = async ev => {
      const act = ev.target.closest('[data-act]')?.dataset.act;
      if (!act) return;
      closePopover();
      const chat = S.chats.find(c => c.id === id);
      if (act === 'rename') {
        const title = await promptDialog(t('renameChat'), chat?.title);
        if (title) { await api('PATCH', `/chats/${id}`, { title }); chat.title = title; renderChats(); if (id === S.chatId) $('#chat-title').textContent = title; }
      } else if (await confirmDialog(t('deleteChat'), t('deleteChatSub'))) {
        await api('DELETE', `/chats/${id}`);
        S.chats = S.chats.filter(c => c.id !== id);
        if (id === S.chatId) newChat(); else renderChats();
      }
    };
    return;
  }
  openChat(id);
  closeNav();
});

$('#chat-search').addEventListener('input', e => { S.search = e.target.value; renderChats(); });

function newChat() {
  if (S.controller) S.controller.abort();
  S.chatId = null;
  S.messages = [];
  history.replaceState(null, '', location.pathname);
  $('#chat-title').textContent = '';
  renderChats();
  renderMessages();
  closeNav();
  $('#input').focus();
}
$('#new-chat').onclick = newChat;
$('#new-chat-m').onclick = newChat;

async function openChat(id) {
  if (S.controller) S.controller.abort();
  let data;
  try { data = await api('GET', `/chats/${id}`); } catch { newChat(); return; }
  S.chatId = id;
  S.messages = data.messages;
  if (modelById(data.chat.model)) { S.model = data.chat.model; renderModelLabel(); }
  history.replaceState(null, '', `#/c/${id}`);
  $('#chat-title').textContent = data.chat.title;
  renderChats();
  renderMessages();
  scrollToBottom(true);
}

/* ================= messages ================= */
function attachmentChip(a, { removable = false, idx = 0 } = {}) {
  const paste = a.kind === 'paste' || a.name.startsWith(t('pastedText'));
  const size = new Blob([a.text]).size;
  return `<div class="att" data-idx="${idx}" title="${esc(a.name)}">
    <span class="att-icon${paste ? ' paste' : ''}">${icon(paste ? 'paste' : 'file')}</span>
    <span class="att-info"><span class="att-name">${esc(a.name)}</span>
    <span class="att-meta">${esc(t('lines', lineCount(a.text)))} · ${fmtSize(size)}</span></span>
    ${removable ? `<button type="button" class="icon-btn att-x" data-remove>${icon('x')}</button>` : ''}</div>`;
}

function previewAttachment(a, onInsert) {
  const pre = h('<pre class="preview-pre"></pre>');
  pre.textContent = a.text;
  const foot = [button(t('close'), '', () => dlg.close())];
  if (onInsert) foot.unshift(button(t('insertBack'), '', () => { onInsert(); dlg.close(); }));
  const dlg = openModal({ title: a.name, body: pre, foot, wide: true });
}

function noticeHtml(m, isLast) {
  if (!m.error) return '';
  const known = { stopped: t('stopped'), refusal: t('refusal'), max_tokens: t('maxTokens') };
  const text = known[m.error] || t('errorPrefix') + m.error;
  const cls = known[m.error] ? 'notice' : 'notice err';
  const retry = isLast && m.error !== 'max_tokens' ? `<button class="notice-btn" data-act="regen">${icon('refresh')}<span>${esc(t('retry'))}</span></button>` : '';
  return `<div class="${cls}">${icon('alert')}<span class="grow">${esc(text)}</span>${retry}</div>`;
}

function messageEl(m, i) {
  if (m.role === 'user') {
    const el = h(`<div class="msg msg-user">
      ${m.attachments?.length ? `<div class="att-row">${m.attachments.map((a, j) => attachmentChip(a, { idx: j })).join('')}</div>` : ''}
      ${m.content ? '<div class="bubble"></div>' : ''}
      <div class="msg-actions"><button class="icon-btn" data-act="copy" title="${esc(t('copy'))}">${icon('copy')}</button></div></div>`);
    if (m.content) el.querySelector('.bubble').textContent = m.content;
    el.querySelectorAll('.att').forEach(c => { c.onclick = () => previewAttachment(m.attachments[+c.dataset.idx]); });
    return el;
  }
  const el = h(`<div class="msg msg-ai"><div class="ai-body"></div></div>`);
  fillAssistant(el.querySelector('.ai-body'), m, i === S.messages.length - 1);
  return el;
}

const elapsed = m => Math.max(0, Math.floor((Date.now() - m.startedAt) / 1000));
const fmtDuration = ms => ms >= 60000 ? t('secs', Math.round(ms / 1000)) : (ms / 1000).toFixed(ms < 10000 ? 1 : 0) + 's';

// An assistant turn goes through: thinking (spinner + live timer, streamed
// thoughts if the model shares them) → answering (text streams in, spinner
// below) → done (thoughts fold into "Thought for Ns", actions + meta line).
function fillAssistant(body, m, isLast) {
  const live = !!m.streaming;
  const thinking = live && !m.content;
  const thinkMs = m.think_ms || m.thinkMs || 0;
  let html = '';
  if (thinking) {
    html += `<div class="think live${m.reasoning ? ' has-body' : ''}"><div class="think-head">
      <span class="spin"></span><span class="shimmer">${esc(t('thinkingNow'))}</span><span class="timer">${esc(t('secs', elapsed(m)))}</span></div>
      ${m.reasoning ? '<div class="r-body"></div>' : ''}</div>`;
  } else if (m.reasoning || thinkMs >= 2000) {
    const label = esc(t('thoughtFor', t('secs', Math.max(1, Math.round(thinkMs / 1000)))));
    html += m.reasoning
      ? `<details class="think"${m._open ? ' open' : ''}><summary>${icon('bulb')}<span>${label}</span>${icon('chevron')}</summary><div class="r-body"></div></details>`
      : `<div class="think"><div class="think-head static">${icon('bulb')}<span>${label}</span></div></div>`;
  }
  html += '<div class="md"></div>';
  if (live && m.content) html += '<div class="writing"><span class="spin"></span></div>';
  html += noticeHtml(m, isLast && !live);
  if (!live) {
    const meta = [m.model, m.duration_ms ? fmtDuration(m.duration_ms) : '', m.tokens_out ? `${fmtTokens(m.tokens_in)} → ${fmtTokens(m.tokens_out)} tokens` : ''].filter(Boolean).join(' · ');
    html += `<div class="msg-actions${isLast ? ' show' : ''}">
      ${m.content ? `<button class="icon-btn" data-act="copy" title="${esc(t('copy'))}">${icon('copy')}</button>` : ''}
      ${isLast && !m.error ? `<button class="icon-btn" data-act="regen" title="${esc(t('regenerate'))}">${icon('refresh')}</button>` : ''}
      <span class="meta">${esc(meta)}</span></div>`;
  }
  body.innerHTML = html;
  const rb = body.querySelector('.r-body');
  if (rb) {
    rb.textContent = m.reasoning;
    if (thinking) rb.scrollTop = rb.scrollHeight;
  }
  const det = body.querySelector('details.think');
  if (det) det.ontoggle = () => { m._open = det.open; };
  renderMarkdown(body.querySelector('.md'), m.content, live);
}

// Live timers tick without re-rendering the message.
setInterval(() => {
  const reply = S.messages.at(-1);
  if (!reply?.streaming) return;
  const el = $('#messages').lastElementChild?.querySelector('.timer');
  if (el) el.textContent = t('secs', elapsed(reply));
}, 250);

function renderMessages() {
  const box = $('#messages');
  box.innerHTML = '';
  S.messages.forEach((m, i) => box.append(messageEl(m, i)));
  $('#main').classList.toggle('empty', S.messages.length === 0);
  const name = S.me?.username || '';
  $('#welcome-title').textContent = S.models.length ? t('welcome', name) : (S.me?.is_admin ? t('welcomeNoModelAdmin') : t('welcomeNoModel'));
}

$('#messages').addEventListener('click', e => {
  const b = e.target.closest('[data-act]');
  if (!b) return;
  const idx = [...$('#messages').children].indexOf(b.closest('.msg'));
  const m = S.messages[idx];
  if (!m) return;
  if (b.dataset.act === 'copy') {
    copyText(m.content).then(() => { b.innerHTML = icon('check'); setTimeout(() => { b.innerHTML = icon('copy'); }, 1500); });
  } else if (b.dataset.act === 'regen') {
    send({ regenerate: true });
  }
});

/* ================= scrolling ================= */
const thread = $('#thread');
let stickBottom = true;
function nearBottom() { return thread.scrollHeight - thread.scrollTop - thread.clientHeight < 80; }
function scrollToBottom(force) {
  if (force || stickBottom) thread.scrollTop = thread.scrollHeight;
}
thread.addEventListener('scroll', () => {
  stickBottom = nearBottom();
  $('#to-bottom').hidden = stickBottom || S.messages.length === 0;
});
$('#to-bottom').onclick = () => { stickBottom = true; scrollToBottom(true); };

/* ================= composer ================= */
const input = $('#input');
const PASTE_CHARS = 2000, PASTE_LINES = 40, MAX_FILE = 1024 * 1024, MAX_ATT = 10;

function autosize() {
  input.style.height = 'auto';
  input.style.height = Math.min(input.scrollHeight, window.innerHeight * 0.4) + 'px';
  updateSend();
}
input.addEventListener('input', autosize);

function updateSend() {
  const btn = $('#send-btn');
  const busy = !!S.controller;
  btn.classList.toggle('stop', busy);
  btn.innerHTML = icon(busy ? 'stop' : 'send');
  btn.disabled = !busy && !input.value.trim() && !S.attachments.length;
}

function renderAttachments() {
  const box = $('#attachments');
  box.innerHTML = S.attachments.map((a, i) => attachmentChip(a, { removable: true, idx: i })).join('');
  updateSend();
}
$('#attachments').addEventListener('click', e => {
  const chip = e.target.closest('.att');
  if (!chip) return;
  const i = +chip.dataset.idx;
  if (e.target.closest('[data-remove]')) { S.attachments.splice(i, 1); renderAttachments(); return; }
  const a = S.attachments[i];
  previewAttachment(a, () => {
    S.attachments.splice(i, 1); renderAttachments();
    input.setRangeText(a.text, input.selectionStart, input.selectionEnd, 'end');
    autosize(); input.focus();
  });
});

function addAttachment(a) {
  if (S.attachments.length >= MAX_ATT) { toast(t('tooMany'), true); return false; }
  S.attachments.push(a);
  renderAttachments();
  return true;
}

input.addEventListener('paste', e => {
  const cd = e.clipboardData;
  if (!cd) return;
  const text = cd.getData('text/plain');
  if (!text && [...cd.items].some(it => it.type.startsWith('image/'))) { e.preventDefault(); toast(t('imageNotSupported'), true); return; }
  if (text.length > PASTE_CHARS || lineCount(text) > PASTE_LINES) {
    e.preventDefault();
    const n = S.attachments.filter(a => a.kind === 'paste').length + 1;
    addAttachment({ name: n > 1 ? `${t('pastedText')} ${n}` : t('pastedText'), text, kind: 'paste' });
  }
});

async function addFiles(files) {
  for (const f of files) {
    if (f.size > MAX_FILE) { toast(t('fileTooBig', f.name), true); continue; }
    const text = await f.text();
    if (text.slice(0, 8000).includes('\u0000') || (f.type && /^(image|video|audio)\//.test(f.type))) { toast(t('binaryFile', f.name), true); continue; }
    if (!addAttachment({ name: f.name, text })) break;
  }
}
$('#attach-btn').onclick = () => $('#file-input').click();
$('#file-input').onchange = e => { addFiles([...e.target.files]); e.target.value = ''; };

let dragDepth = 0;
document.addEventListener('dragenter', e => { if (e.dataTransfer?.types.includes('Files') && !$('#app').hidden) { dragDepth++; document.body.classList.add('dragging'); } });
document.addEventListener('dragleave', () => { if (--dragDepth <= 0) { dragDepth = 0; document.body.classList.remove('dragging'); } });
document.addEventListener('dragover', e => { if (e.dataTransfer?.types.includes('Files')) e.preventDefault(); });
document.addEventListener('drop', e => {
  if (!e.dataTransfer?.files.length) return;
  e.preventDefault(); dragDepth = 0; document.body.classList.remove('dragging');
  if (!$('#app').hidden) addFiles([...e.dataTransfer.files]);
});

input.addEventListener('keydown', e => {
  if (e.key === 'Enter' && !e.shiftKey && !e.isComposing && e.keyCode !== 229) {
    if (window.matchMedia('(max-width: 820px)').matches) return; // phones: Enter = newline
    e.preventDefault();
    send();
  }
});
$('#composer').addEventListener('submit', e => {
  e.preventDefault();
  if (S.controller) S.controller.abort(); else send();
});

/* ================= sending ================= */
async function send({ regenerate = false } = {}) {
  if (S.controller) return;
  if (!S.model) { toast(t('pickModel'), true); openModelMenu(); return; }
  const content = regenerate ? '' : input.value;
  const attachments = regenerate ? [] : S.attachments.map(({ name, text }) => ({ name, text }));
  if (!regenerate && !content.trim() && !attachments.length) return;

  const draft = { content: input.value, attachments: S.attachments };
  let userMsg = null;
  if (regenerate) {
    if (S.messages.at(-1)?.role === 'assistant') S.messages.pop();
  } else {
    userMsg = { role: 'user', content, attachments, pending: true };
    S.messages.push(userMsg);
    input.value = ''; S.attachments = []; renderAttachments(); autosize();
  }
  const reply = { role: 'assistant', content: '', reasoning: '', streaming: true, model: modelById(S.model)?.model, startedAt: Date.now() };
  S.messages.push(reply);
  renderMessages();
  stickBottom = true; scrollToBottom(true);

  const controller = new AbortController();
  S.controller = controller;
  updateSend();
  const replyBody = () => $('#messages').lastElementChild?.querySelector('.ai-body');
  let frame = 0;
  const repaint = () => {
    if (frame) return;
    frame = requestAnimationFrame(() => { frame = 0; const b = replyBody(); if (b) fillAssistant(b, reply, true); scrollToBottom(); });
  };

  let started = false;
  try {
    const res = await fetch('/api/send', {
      method: 'POST', credentials: 'same-origin', signal: controller.signal,
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ chat_id: S.chatId || 0, content, attachments, model: S.model, regenerate }),
    });
    if (res.status === 401) { showLogin(); return; }
    if (!res.ok) {
      let msg = res.statusText;
      try {
        const e = await res.json();
        msg = e.error || msg;
        if (e.code === 'quota_exceeded') { msg = t('quotaOut', fmtDate(e.reset_at)); S.monthTokens = Math.max(S.monthTokens, S.me.token_limit); renderQuota(); }
      } catch {}
      throw new HttpError(res.status, msg);
    }
    const reader = res.body.getReader();
    const dec = new TextDecoder();
    let buf = '';
    for (;;) {
      const { value, done } = await reader.read();
      if (done) break;
      buf += dec.decode(value, { stream: true });
      let cut;
      while ((cut = buf.indexOf('\n\n')) >= 0) {
        const chunk = buf.slice(0, cut); buf = buf.slice(cut + 2);
        const line = chunk.split('\n').find(l => l.startsWith('data: '));
        if (!line) continue;
        const ev = JSON.parse(line.slice(6));
        if (ev.type === 'start') {
          started = true;
          onChatStarted(ev.chat);
          if (ev.user_message && userMsg) Object.assign(userMsg, ev.user_message, { pending: false });
        } else if (ev.type === 'text') { if (!reply.content) reply.thinkMs = Date.now() - reply.startedAt; reply.content += ev.t; repaint(); }
        else if (ev.type === 'reasoning') { reply.reasoning += ev.t; repaint(); }
        else if (ev.type === 'done') { Object.assign(reply, ev.message, { streaming: false }); }
      }
    }
  } catch (err) {
    if (err.name === 'AbortError') {
      reply.error = 'stopped';
    } else if (!started) {
      // Nothing was saved: put the draft back so the user can retry.
      S.messages.splice(S.messages.indexOf(reply), 1);
      if (userMsg) { S.messages.splice(S.messages.indexOf(userMsg), 1); input.value = draft.content; S.attachments = draft.attachments; renderAttachments(); autosize(); }
      toast(err.message, true);
    } else {
      reply.error = err.message;
    }
  } finally {
    if (frame) cancelAnimationFrame(frame);
    reply.streaming = false;
    if (S.controller === controller) S.controller = null;
    updateSend();
    if (S.messages.includes(reply) || !started) { renderMessages(); scrollToBottom(); }
    api('GET', '/me').then(me => { S.me = me.user; S.monthTokens = me.month_tokens; S.resetAt = me.reset_at; renderQuota(); }).catch(() => {});
  }
}

function onChatStarted(chat) {
  S.chatId = chat.id;
  history.replaceState(null, '', `#/c/${chat.id}`);
  $('#chat-title').textContent = chat.title;
  S.chats = [chat, ...S.chats.filter(c => c.id !== chat.id)];
  renderChats();
}

/* ================= nav (mobile) ================= */
function closeNav() { $('#app').classList.remove('nav-open'); }
$('#open-sidebar').onclick = () => $('#app').classList.add('nav-open');
$('#close-sidebar').onclick = closeNav;
$('#scrim').onclick = closeNav;

document.addEventListener('keydown', e => {
  if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'k' && !$('#app').hidden) { e.preventDefault(); newChat(); }
  if (e.key === 'Escape') closePopover();
});

/* ================= user menu ================= */
$('#user-btn').onclick = () => {
  const theme = store.get('theme', 'system');
  const limit = S.me.token_limit;
  const pop = openPopover($('#user-btn'), `
    <div class="menu-label">${esc(t('monthUsage'))}: ${fmtTokens(S.monthTokens)}${limit ? ' / ' + fmtTokens(limit) : ''} tokens</div>
    <div class="menu-item">${icon('sun')}<span>${esc(t('theme'))}</span><span class="seg" data-seg="theme">
      ${['system', 'light', 'dark'].map(v => `<button data-v="${v}" class="${theme === v ? 'on' : ''}">${esc(t('theme' + v[0].toUpperCase() + v.slice(1)))}</button>`).join('')}</span></div>
    <div class="menu-item">${icon('globe')}<span>${esc(t('language'))}</span><span class="seg" data-seg="lang">
      <button data-v="zh" class="${lang === 'zh' ? 'on' : ''}">中文</button><button data-v="en" class="${lang === 'en' ? 'on' : ''}">EN</button></span></div>
    <div class="menu-sep"></div>
    ${S.userKeys !== 'off' ? `<button class="menu-item" data-act="mine">${icon('cube')}<span>${esc(t('myModels'))}</span></button>` : ''}
    <button class="menu-item" data-act="password">${icon('key')}<span>${esc(t('changePassword'))}</span></button>
    ${S.me.is_admin ? `<button class="menu-item" data-act="admin">${icon('shield')}<span>${esc(t('admin'))}</span></button>` : ''}
    <div class="menu-sep"></div>
    <button class="menu-item danger" data-act="logout">${icon('logout')}<span>${esc(t('logout'))}</span></button>`, { minWidth: 252, above: true });
  pop.onclick = async e => {
    const seg = e.target.closest('[data-seg] button');
    if (seg) {
      const kind = seg.parentElement.dataset.seg;
      seg.parentElement.querySelectorAll('button').forEach(b => b.classList.toggle('on', b === seg));
      if (kind === 'theme') setTheme(seg.dataset.v);
      else { lang = seg.dataset.v; store.set('lang', lang); applyI18n(); renderModelLabel(); renderChats(); renderMessages(); closePopover(); }
      return;
    }
    const act = e.target.closest('[data-act]')?.dataset.act;
    if (!act) return;
    closePopover();
    if (act === 'logout') { await api('POST', '/logout').catch(() => {}); location.hash = ''; showLogin(); }
    else if (act === 'password') changePasswordDialog();
    else if (act === 'admin') openAdmin();
    else if (act === 'mine') openMyModels();
  };
};

function changePasswordDialog() {
  const form = h(`<form style="display:flex;flex-direction:column;gap:14px">
    <label class="field"><span>${esc(t('oldPassword'))}</span><input name="old" type="password" autocomplete="current-password" required></label>
    <label class="field"><span>${esc(t('newPassword'))}</span><input name="new1" type="password" autocomplete="new-password" minlength="6" required></label>
    <label class="field"><span>${esc(t('newPassword2'))}</span><input name="new2" type="password" autocomplete="new-password" minlength="6" required></label>
    <div class="form-error"></div></form>`);
  const submit = async () => {
    const f = form, err = f.querySelector('.form-error');
    if (!f.reportValidity()) return;
    if (f.new1.value !== f.new2.value) { err.textContent = t('pwMismatch'); return; }
    try { await api('POST', '/me/password', { old: f.old.value, new: f.new1.value }); dlg.close(); toast(t('pwChanged')); }
    catch (e) { err.textContent = e.message; }
  };
  form.onsubmit = e => { e.preventDefault(); submit(); };
  const dlg = openModal({ title: t('changePassword'), body: form, foot: [button(t('cancel'), '', () => dlg.close()), button(t('save'), 'btn-primary', submit)] });
}

/* ================= admin ================= */
function openAdmin() {
  const body = h('<div style="display:flex;flex-direction:column;gap:14px"></div>');
  const dlg = openModal({ title: t('admin'), body, wide: true, onClose: () => loadModels().then(renderMessages) });
  const tabs = h(`<div class="tabs">
    <button class="tab" data-tab="users">${esc(t('users'))}</button>
    <button class="tab" data-tab="providers">${esc(t('providers'))}</button>
    <button class="tab" data-tab="usage">${esc(t('usage'))}</button>
    <button class="tab" data-tab="access">${esc(t('access'))}</button>
    <button class="tab" data-tab="settings">${esc(t('settings'))}</button></div>`);
  dlg.querySelector('.modal-head').after(tabs);
  dlg.querySelector('.modal-body').style.paddingTop = '18px';
  const show = (name, arg) => {
    tabs.querySelectorAll('.tab').forEach(b => b.classList.toggle('on', b.dataset.tab === name));
    store.set('adminTab', name);
    ({ users: adminUsers, providers: adminProviders, usage: adminUsage, access: adminAccess, settings: adminSettings })[name](body, arg);
  };
  body.addEventListener('admin-show', e => show(e.detail.tab, e.detail.arg));
  tabs.onclick = e => { const b = e.target.closest('[data-tab]'); if (b) show(b.dataset.tab); };
  show(store.get('adminTab', 'providers'));
}

async function adminUsers(body) {
  const [users, usage] = await Promise.all([api('GET', '/admin/users'), api('GET', '/admin/usage')]);
  const used = Object.fromEntries(usage.rows.map(r => [r.user_id, r.tokens_in + r.tokens_out]));
  body.innerHTML = `<div class="row"><div class="grow"></div><button class="btn btn-primary btn-sm" data-add>${icon('plus')}${esc(t('addUser'))}</button></div>
    <div class="table-wrap"><table class="table"><thead><tr>
      <th>${esc(t('username'))}</th><th>${esc(t('status'))}</th><th>${esc(t('monthUsage'))}</th><th>${esc(t('lastActive'))}</th><th></th></tr></thead><tbody>
    ${users.map(u => {
      const n = used[u.id] || 0;
      const quota = u.token_limit
        ? `${fmtTokens(n)} / ${fmtTokens(u.token_limit)}<div class="sub">${esc(t('remaining', fmtTokens(Math.max(0, u.token_limit - n))))} · ${esc(t('resetOn', fmtDate(usage.reset_at)))}</div>`
        : `${fmtTokens(n)} <span class="sub">/ ${esc(t('unlimited'))}</span>`;
      return `<tr data-id="${u.id}">
      <td><strong>${esc(u.username)}</strong> ${u.is_admin ? `<span class="badge accent">${esc(t('adminRole'))}</span>` : ''}</td>
      <td>${u.disabled ? `<span class="badge danger">${esc(t('disabled'))}</span>` : `<span class="badge ok">${esc(t('active'))}</span>`}</td>
      <td>${quota}</td>
      <td>${esc(fmtAgo(u.last_seen_at))}${u.last_ip ? `<div class="sub mono">${flag(u.last_country)}${esc(u.last_ip)}</div>` : ''}</td>
      <td class="actions"><button class="icon-btn" data-log title="${esc(t('access'))}">${icon('clock')}</button>
        <button class="icon-btn" data-edit title="${esc(t('editUser'))}">${icon('edit')}</button>
        ${u.id !== S.me.id ? `<button class="icon-btn" data-del title="${esc(t('delete'))}">${icon('trash')}</button>` : ''}</td></tr>`;
    }).join('')}
    </tbody></table></div>`;
  body.onchange = null;
  body.onclick = async e => {
    if (e.target.closest('[data-add]')) return userForm(null, () => adminUsers(body));
    const row = e.target.closest('tr[data-id]');
    if (!row) return;
    const u = users.find(x => x.id === +row.dataset.id);
    if (e.target.closest('[data-log]')) body.dispatchEvent(new CustomEvent('admin-show', { detail: { tab: 'access', arg: u.id } }));
    else if (e.target.closest('[data-edit]')) userForm(u, () => adminUsers(body));
    else if (e.target.closest('[data-del]') && await confirmDialog(t('deleteUser', u.username), t('deleteUserSub'))) {
      await api('DELETE', `/admin/users/${u.id}`); adminUsers(body);
    }
  };
}

async function adminAccess(body, userID) {
  const users = await api('GET', '/admin/users');
  let uid = userID || 0, before = 0;
  const evName = { login: t('evLogin'), login_failed: t('evLoginFailed'), chat: t('evChat') };
  body.innerHTML = `<div class="row"><label class="field" style="flex:none;min-width:200px"><select data-user>
      <option value="0">${esc(t('allUsers'))}</option>${users.map(u => `<option value="${u.id}" ${u.id === uid ? 'selected' : ''}>${esc(u.username)}</option>`).join('')}</select></label></div>
    <div class="table-wrap"><table class="table log"><thead><tr><th>${esc(t('time'))}</th><th>${esc(t('username'))}</th><th>${esc(t('event'))}</th>
      <th>${esc(t('ipAddr'))}</th><th>${esc(t('device'))}</th><th>${esc(t('detailCol'))}</th></tr></thead><tbody></tbody></table></div>
    <div class="row" data-more hidden><div class="grow"></div><button class="btn btn-sm" data-load>${esc(t('loadMore'))}</button><div class="grow"></div></div>`;
  const tbody = body.querySelector('tbody');
  const load = async (reset) => {
    if (reset) { tbody.innerHTML = ''; before = 0; }
    const rows = await api('GET', `/admin/access?user=${uid}&before=${before}`);
    tbody.insertAdjacentHTML('beforeend', rows.map(r => `<tr class="${r.ok ? '' : 'bad'}">
      <td class="nowrap">${esc(fmtDateTime(r.created_at))}</td><td>${esc(r.username || '—')}</td>
      <td><span class="badge${r.ok ? '' : ' danger'}">${esc(evName[r.event] || r.event)}</span></td>
      <td class="mono nowrap">${flag(r.country)}${esc(r.ip)}</td><td class="nowrap" title="${esc(r.user_agent)}">${esc(uaSummary(r.user_agent))}</td>
      <td class="sub">${esc(r.detail)}</td></tr>`).join(''));
    if (!tbody.children.length) tbody.innerHTML = `<tr><td colspan="6" class="empty-state">—</td></tr>`;
    before = rows.at(-1)?.id || before;
    body.querySelector('[data-more]').hidden = rows.length < 100;
  };
  body.onclick = e => { if (e.target.closest('[data-load]')) load(false); };
  body.onchange = e => { if (e.target.matches('[data-user]')) { uid = +e.target.value; load(true); } };
  await load(true);
}

async function userForm(u, done) {
  const defaults = u ? null : await api('GET', '/admin/settings');
  const form = h(`<form style="display:flex;flex-direction:column;gap:14px">
    <label class="field"><span>${esc(t('username'))}</span><input name="username" required ${u ? 'disabled' : ''}></label>
    <label class="field"><span>${esc(t('password'))}</span><input name="password" type="password" autocomplete="new-password" ${u ? '' : 'required minlength="6"'} placeholder="${u ? esc(t('pwKeep')) : ''}"></label>
    <label class="field"><span>${esc(t('limit'))} <small>${esc(t('limitHint'))}</small></span><input name="limit" type="number" min="0" step="1000"></label>
    <label class="check"><input type="checkbox" name="admin"><span>${esc(t('isAdmin'))}</span></label>
    ${u ? `<label class="check"><input type="checkbox" name="disabled"><span>${esc(t('isDisabled'))}</span></label>` : ''}
    <div class="form-error"></div></form>`);
  form.username.value = u?.username || '';
  form.limit.value = u ? u.token_limit || 0 : defaults.default_token_limit || 0;
  form.admin.checked = !!u?.is_admin;
  if (u) form.disabled.checked = u.disabled;
  const submit = async () => {
    if (!form.reportValidity()) return;
    const body = { password: form.password.value, is_admin: form.admin.checked, token_limit: +form.limit.value || 0 };
    try {
      if (u) { body.disabled = form.disabled.checked; await api('PATCH', `/admin/users/${u.id}`, body); }
      else { body.username = form.username.value; await api('POST', '/admin/users', body); }
      dlg.close(); done();
    } catch (e) { form.querySelector('.form-error').textContent = e.message; }
  };
  form.onsubmit = e => { e.preventDefault(); submit(); };
  const dlg = openModal({ title: u ? t('editUser') : t('addUser'), body: form, foot: [button(t('cancel'), '', () => dlg.close()), button(t('save'), 'btn-primary', submit)] });
}

// Provider lists come in two flavours: shared ones in the admin panel and the
// user's own under "My models". They differ only in API base and sharing fields.
const SHARED = { base: '/admin/providers', shared: true };
const OWN = { base: '/my/providers', shared: false };

function rateLine(r) {
  if (!r) return `<span class="sub">${esc(t('noRateYet'))}</span>`;
  const now = Date.now() / 1000;
  const part = (label, left, limit, reset, fmt) => limit
    ? `<span>${esc(label)} <strong>${fmt(left)}</strong>/${fmt(limit)}${reset > now ? ` · ${esc(t('resetIn', fmtSpan(reset - now)))}` : ''}</span>` : '';
  const exact = n => (n || 0).toLocaleString();
  return [part(t('rateReq'), r.req_remaining, r.req_limit, r.req_reset, exact), part(t('rateTok'), r.tok_remaining, r.tok_limit, r.tok_reset, fmtTokens),
    `<span class="sub">${esc(t('updated', fmtAgo(r.updated_at)))}</span>`].filter(Boolean).join('<span class="dot">·</span>');
}

function providerCards(ps, scope) {
  if (!ps.length) return scope.shared ? '' : `<div class="empty-state">${esc(t('noOwn'))}</div>`;
  return `<div class="prov-list">${ps.map(p => `<div class="prov" data-id="${p.id}">
    <div class="prov-logo" style="background:${providerColor(p.name, p.kind, p.preset)}">${esc(p.name.slice(0, 1).toUpperCase())}</div>
    <div class="prov-info">
      <div class="prov-name">${esc(p.name)}<span class="badge">${esc(p.kind === 'anthropic' ? 'Anthropic' : 'OpenAI API')}</span>
        ${scope.shared && p.visibility === 'users' ? `<span class="badge accent">${esc(t('usersCount', p.user_ids.length))}</span>` : ''}
        ${p.blocked ? `<span class="badge danger">${esc(t('blockedMode'))}</span>` : ''}
        ${p.key_error ? `<span class="badge danger" title="${esc(p.key_error)}">${esc(t('noKey'))}</span>` : p.has_key ? '' : `<span class="badge danger">${esc(t('noKey'))}</span>`}</div>
      <div class="prov-sub">${esc(t('modelCount', p.models.length))} · ${esc(p.base_url || (p.kind === 'anthropic' ? 'api.anthropic.com' : 'api.openai.com'))}</div>
      ${scope.shared && p.has_key ? `<div class="prov-rate">${rateLine(p.rate)}${p.balance_supported ? ` <button type="button" class="link-btn" data-balance>${esc(t('checkBalance'))}</button><span data-balance-out></span>` : ''}</div>` : ''}
    </div>
    <label class="switch" title="${esc(t('enabled'))}"><input type="checkbox" data-toggle ${p.enabled ? 'checked' : ''}><span></span></label>
    <button class="icon-btn" data-edit>${icon('edit')}</button></div>`).join('')}</div>`;
}

async function providerList(body, scope) {
  const [ps, users] = await Promise.all([api('GET', scope.base), scope.shared ? api('GET', '/admin/users') : []]);
  const reload = () => providerList(body, scope);
  body.innerHTML = `${scope.shared ? '' : `<p style="margin:0;color:var(--text-2);font-size:14px">${esc(t('myModelsSub'))}</p>`}
    <div class="row"><div class="grow"></div><button class="btn btn-primary btn-sm" data-add>${icon('plus')}${esc(t('addProvider'))}</button></div>
    ${providerCards(ps, scope)}`;
  body.onclick = async e => {
    if (e.target.closest('[data-add]')) return providerForm(null, scope, users, reload);
    const card = e.target.closest('.prov');
    const bal = e.target.closest('[data-balance]');
    if (card && bal) {
      const out = card.querySelector('[data-balance-out]');
      bal.disabled = true; out.textContent = '…';
      try {
        const st = await api('GET', `/admin/providers/${card.dataset.id}/status`);
        out.innerHTML = st.balance ? ' ' + st.balance.map(b => `${esc(t('balance'))} <strong>${esc(b.amount)}</strong> ${esc(b.currency)}`).join(' · ')
          : ` <span class="err-text">${esc(st.balance_error || '—')}</span>`;
      } catch (ex) { out.innerHTML = ` <span class="err-text">${esc(ex.message)}</span>`; }
      bal.disabled = false;
      return;
    }
    if (card && e.target.closest('[data-edit]')) providerForm(ps.find(p => p.id === +card.dataset.id), scope, users, reload);
  };
  body.onchange = async e => {
    if (!e.target.matches('[data-toggle]')) return;
    const p = ps.find(x => x.id === +e.target.closest('.prov').dataset.id);
    p.enabled = e.target.checked;
    await api('PUT', `${scope.base}/${p.id}`, { name: p.name, kind: p.kind, base_url: p.base_url, models: p.models,
      enabled: p.enabled, sort: p.sort, visibility: p.visibility, user_ids: p.user_ids, preset: p.preset, auto_models: p.auto_models });
  };
}

const adminProviders = body => providerList(body, SHARED);

function openMyModels() {
  const body = h('<div style="display:flex;flex-direction:column;gap:14px"></div>');
  openModal({ title: t('myModels'), body, wide: true, onClose: () => loadModels().then(renderMessages) });
  providerList(body, OWN);
}

function providerForm(p, scope, users, done) {
  // Pick a provider from the built-in catalog (or a custom endpoint), paste a key,
  // and by default follow the catalog's model list. In "official" mode personal
  // providers are catalog-only and the address can't be changed.
  const official = !scope.shared && S.userKeys === 'official';
  const presetOpts = S.catalog.map(pr => `<option value="${esc(pr.id)}">${esc(pr.name)}${pr.region === 'cn' ? ` · ${esc(t('cnRegion'))}` : ''}</option>`).join('');
  const form = h(`<form style="display:flex;flex-direction:column;gap:14px">
    <div class="row">
      <label class="field"><span>${esc(t('providerPreset'))}</span><select name="preset">${presetOpts}
        ${official ? '' : `<option value="custom:openai">${esc(t('customOpenAI'))}</option><option value="custom:anthropic">${esc(t('customAnthropic'))}</option>`}</select></label>
      <label class="field"><span>${esc(t('name'))}</span><input name="name" required></label>
    </div>
    <label class="field"><span class="row" style="gap:8px">${esc(t('apiKey'))}<a class="key-link" target="_blank" rel="noopener" hidden>${esc(t('getKey'))} ↗</a></span>
      <input name="api_key" type="password" autocomplete="off" placeholder="${p?.has_key ? esc(t('keySet', p.key_hint || '••••')) : esc(t('keyNotSet'))}"></label>
    <div class="field"><span>${esc(t('models'))}</span>
      <label class="check" data-auto-row><input type="checkbox" name="auto_models"><span>${esc(t('autoModels'))}</span></label>
      <div class="model-picks" data-picks></div>
      <div data-manual style="display:flex;flex-direction:column;gap:8px">
        <label class="field"><span>${esc(t('moreModels'))} <small>${esc(t('moreModelsHint'))}</small></span><textarea name="models" rows="3"></textarea></label>
        <div><button type="button" class="btn btn-sm" data-fetch>${icon('refresh')}<span>${esc(t('fetchModels'))}</span></button></div>
      </div>
    </div>
    ${official ? '<input type="hidden" name="base_url"><input type="hidden" name="kind">' : `<details class="adv"><summary>${icon('right')}<span>${esc(t('advanced'))}</span></summary>
      <div style="display:flex;flex-direction:column;gap:14px;margin-top:12px">
        <label class="field" data-kind-row><span>${esc(t('kind'))}</span><select name="kind">
          <option value="openai">${esc(t('kindOpenAI'))}</option><option value="anthropic">${esc(t('kindAnthropic'))}</option></select></label>
        <label class="field"><span>${esc(t('baseURL'))}</span><input name="base_url"></label>
      </div></details>`}
    ${scope.shared ? `<label class="field"><span>${esc(t('visibility'))}</span><select name="visibility">
        <option value="all">${esc(t('everyone'))}</option><option value="users">${esc(t('selectedUsers'))}</option></select></label>
      <div class="model-picks" data-users hidden></div>` : ''}
    <label class="check"><input type="checkbox" name="enabled"><span>${esc(t('enabled'))}</span></label>
    <div class="form-error"></div></form>`);

  const picks = form.querySelector('[data-picks]'), manual = form.querySelector('[data-manual]');
  const autoRow = form.querySelector('[data-auto-row]'), keyLink = form.querySelector('.key-link');
  const preset = () => S.catalog.find(c => c.id === form.preset.value);
  const lines = text => text.split('\n').map(x => x.trim()).filter(Boolean);
  let selected = new Set(p && !p.auto_models ? p.models : []);
  let lastName = '';

  const renderPicks = () => {
    const pr = preset();
    const auto = pr && form.auto_models.checked;
    autoRow.hidden = !pr;
    manual.hidden = auto;
    const list = pr ? pr.models : [];
    picks.hidden = !list.length;
    picks.innerHTML = list.map(m => `<label class="${auto ? 'locked' : ''}"><input type="checkbox" value="${esc(m)}" ${auto || selected.has(m) ? 'checked' : ''} ${auto ? 'disabled' : ''}>${esc(m)}</label>`).join('');
  };
  const onPreset = (first) => {
    const pr = preset();
    const custom = form.preset.value.startsWith('custom:');
    if (pr) {
      form.kind.value = pr.kind;
      if (!first || !p) form.base_url.value = pr.base_url;
      if (!form.name.value || form.name.value === lastName) form.name.value = pr.name;
      lastName = pr.name;
      keyLink.href = pr.key_url; keyLink.hidden = !pr.key_url;
      if (!first) { form.auto_models.checked = true; selected = new Set(pr.models); }
    } else {
      form.kind.value = form.preset.value.slice(7);
      if (!first) { form.base_url.value = ''; if (form.name.value === lastName) form.name.value = ''; lastName = ''; }
      keyLink.hidden = true;
    }
    const kindRow = form.querySelector('[data-kind-row]');
    if (kindRow) kindRow.hidden = !custom;
    if (form.base_url.type !== 'hidden') form.base_url.placeholder = form.kind.value === 'anthropic' ? 'https://api.anthropic.com' : 'https://api.openai.com/v1';
    if (custom) form.querySelector('details.adv').open = true;
    renderPicks();
  };

  // initial state
  const pr0 = p && S.catalog.find(c => c.id === p.preset);
  form.preset.value = pr0 ? pr0.id : p ? `custom:${p.kind}` : (S.catalog[0]?.id || 'custom:openai');
  if (official && !pr0) form.preset.value = S.catalog[0]?.id;
  form.name.value = p?.name || '';
  lastName = pr0 ? pr0.name : '';
  form.base_url.value = p?.base_url || '';
  form.auto_models.checked = p ? !!p.auto_models : true;
  form.models.value = (p?.models || []).filter(m => !(pr0 && pr0.models.includes(m))).join('\n');
  if (p?.auto_models && pr0) form.models.value = '';
  form.enabled.checked = p ? p.enabled : true;
  if (!p && preset()) selected = new Set(preset().models);
  onPreset(true);
  form.preset.onchange = () => onPreset(false);
  if (form.kind.tagName === 'SELECT') form.kind.onchange = () => onPreset(false);
  form.auto_models.onchange = () => {
    if (!form.auto_models.checked && preset() && !selected.size) selected = new Set(preset().models);
    renderPicks();
  };
  picks.onchange = e => { if (e.target.checked) selected.add(e.target.value); else selected.delete(e.target.value); };

  const userBox = form.querySelector('[data-users]');
  if (userBox) {
    const allowed = new Set(p?.user_ids || []);
    userBox.innerHTML = `<div style="width:100%;font:12px var(--font);color:var(--muted)">${esc(t('pickUsers'))}</div>` +
      users.map(u => `<label><input type="checkbox" value="${u.id}" ${allowed.has(u.id) ? 'checked' : ''}>${esc(u.username)}</label>`).join('');
    form.visibility.value = p?.visibility || 'all';
    const syncUsers = () => { userBox.hidden = form.visibility.value !== 'users'; };
    form.visibility.onchange = syncUsers; syncUsers();
  }
  const err = form.querySelector('.form-error');
  const values = () => {
    const pr = preset();
    const auto = !!pr && form.auto_models.checked;
    const models = auto ? [] : [...new Set([...(pr ? pr.models.filter(m => selected.has(m)) : []), ...lines(form.models.value)])];
    const v = { preset: pr ? pr.id : '', auto_models: auto, name: form.name.value, kind: form.kind.value, base_url: form.base_url.value,
      models, enabled: form.enabled.checked, sort: p?.sort || 0 };
    if (form.api_key.value) v.api_key = form.api_key.value;
    if (userBox) {
      v.visibility = form.visibility.value;
      v.user_ids = [...userBox.querySelectorAll('input:checked')].map(i => +i.value);
    }
    return v;
  };
  form.querySelector('[data-fetch]').onclick = async e => {
    const b = e.currentTarget; const label = b.querySelector('span');
    b.disabled = true; label.textContent = t('fetching'); err.textContent = '';
    try {
      const list = await api('POST', `${scope.base}/${p?.id || 0}/models`, values());
      const known = new Set(preset()?.models || []);
      const extra = new Set(lines(form.models.value));
      const box = h('<div class="model-picks"></div>');
      box.innerHTML = list.filter(m => !known.has(m)).map(m => `<label><input type="checkbox" value="${esc(m)}" ${extra.has(m) ? 'checked' : ''}>${esc(m)}</label>`).join('');
      box.onchange = ev => {
        const set = new Set(lines(form.models.value));
        ev.target.checked ? set.add(ev.target.value) : set.delete(ev.target.value);
        form.models.value = [...set].join('\n');
      };
      manual.querySelector('.model-picks')?.remove();
      if (box.children.length) manual.append(box);
    } catch (ex) { err.textContent = ex.message; }
    b.disabled = false; label.textContent = t('fetchModels');
  };
  const submit = async () => {
    if (!form.reportValidity()) return;
    const v = values();
    if (!v.auto_models && !v.models.length) { err.textContent = t('pickOneModel'); return; }
    try {
      if (p) await api('PUT', `${scope.base}/${p.id}`, v); else await api('POST', scope.base, v);
      dlg.close(); toast(t('saved')); done();
    } catch (e) { err.textContent = e.message; }
  };
  form.onsubmit = e => { e.preventDefault(); submit(); };
  const foot = [button(t('cancel'), '', () => dlg.close()), button(t('save'), 'btn-primary', submit)];
  if (p) {
    const del = button(t('delete'), 'btn-danger', async () => {
      if (await confirmDialog(t('deleteProvider', p.name))) { await api('DELETE', `${scope.base}/${p.id}`); dlg.close(); done(); }
    });
    del.style.marginRight = 'auto';
    foot.unshift(del);
  }
  const dlg = openModal({ title: p ? t('editProvider') : t('addProvider'), body: form, foot });
}

async function adminSettings(body) {
  const st = await api('GET', '/admin/settings');
  const render = () => {
    body.innerHTML = `<div class="settings">
      <div class="setting"><span class="setting-text"><strong>${esc(t('defaultLimit'))}</strong><small>${esc(t('defaultLimitSub'))}</small></span>
        <input class="setting-select" type="number" min="0" step="10000" data-key="default_token_limit" value="${st.default_token_limit || 0}"></div>
      <div class="setting"><span class="setting-text"><strong>${esc(t('keyMode'))}</strong><small>${esc(t('keyModeSub'))}</small></span>
        <select class="setting-select" data-key="user_keys">
          ${['off', 'official', 'custom'].map(m => `<option value="${m}" ${st.user_keys === m ? 'selected' : ''}>${esc(t('mode' + m[0].toUpperCase() + m.slice(1)))}</option>`).join('')}
        </select></div>
      <label class="setting${st.user_keys === 'custom' ? '' : ' muted'}"><span class="setting-text"><strong>${esc(t('allowPrivate'))}</strong>
        <small>${esc(t('allowPrivateSub'))} ${st.user_keys === 'custom' ? '' : esc(t('privateNeedsCustom'))}</small></span>
        <span class="switch"><input type="checkbox" data-key="allow_private_network" ${st.allow_private_network ? 'checked' : ''} ${st.user_keys === 'custom' ? '' : 'disabled'}><span></span></span></label></div>
      <div class="notice${st.keys_encrypted ? ' ok' : ''}">${icon(st.keys_encrypted ? 'shield' : 'alert')}<span>${esc(st.keys_encrypted ? t('keysEncrypted') : t('keysPlain'))}</span></div>`;
  };
  render();
  body.onclick = null;
  body.onchange = async e => {
    const k = e.target.dataset.key;
    if (!k) return;
    const v = e.target.type === 'checkbox' ? e.target.checked : e.target.type === 'number' ? Math.max(0, +e.target.value || 0) : e.target.value;
    const res = await api('PUT', '/admin/settings', { [k]: v });
    Object.assign(st, res);
    S.userKeys = res.user_keys;
    render();
    toast(t('saved'));
  };
}

async function adminUsage(body, m) {
  const data = await api('GET', '/admin/usage' + (m ? `?month=${encodeURIComponent(m)}` : ''));
  const max = Math.max(1, ...data.rows.map(r => r.token_limit || r.tokens_in + r.tokens_out));
  body.innerHTML = `<div class="row" style="align-items:flex-end"><label class="field" style="flex:none"><span>${esc(t('month'))}</span><input type="month" value="${esc(data.month)}"></label>
    <span class="sub" style="padding-bottom:10px">${esc(t('resetOn', fmtDate(data.reset_at)))}</span></div>
    <div class="table-wrap"><table class="table"><thead><tr><th>${esc(t('username'))}</th><th>${esc(t('requests'))}</th>
      <th>${esc(t('input'))}</th><th>${esc(t('output'))}</th><th>${esc(t('total'))}</th><th>${esc(t('ownKeys'))}</th><th style="width:24%"></th></tr></thead><tbody>
    ${data.rows.map(r => { const tot = r.tokens_in + r.tokens_out; const cap = r.token_limit || max;
      return `<tr><td><strong>${esc(r.username)}</strong></td><td>${r.requests}</td><td>${fmtTokens(r.tokens_in)}</td><td>${fmtTokens(r.tokens_out)}</td>
      <td><strong>${fmtTokens(tot)}</strong>${r.token_limit ? ` <small style="color:var(--muted)">/ ${fmtTokens(r.token_limit)}</small>` : ''}</td>
      <td style="color:var(--muted)">${r.own_tokens ? fmtTokens(r.own_tokens) : '—'}</td>
      <td><div class="bar"><i style="width:${Math.min(100, tot / cap * 100).toFixed(1)}%"></i></div></td></tr>`; }).join('')}
    </tbody></table></div>`;
  body.onclick = null;
  body.onchange = e => { if (e.target.type === 'month' && e.target.value) adminUsage(body, e.target.value); };
}

/* ================= start ================= */
applyIcons();
applyI18n();
updateSend();
boot();
