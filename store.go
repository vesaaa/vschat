package main

import (
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

const schema = `
CREATE TABLE IF NOT EXISTS users (
	id          INTEGER PRIMARY KEY AUTOINCREMENT,
	username    TEXT NOT NULL UNIQUE COLLATE NOCASE,
	pass_hash   TEXT NOT NULL,
	is_admin    INTEGER NOT NULL DEFAULT 0,
	disabled    INTEGER NOT NULL DEFAULT 0,
	token_limit INTEGER NOT NULL DEFAULT 0,
	created_at  INTEGER NOT NULL
);
CREATE TABLE IF NOT EXISTS sessions (
	token_hash TEXT PRIMARY KEY,
	user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	expires_at INTEGER NOT NULL
);
CREATE TABLE IF NOT EXISTS providers (
	id       INTEGER PRIMARY KEY AUTOINCREMENT,
	name     TEXT NOT NULL,
	kind     TEXT NOT NULL,
	base_url TEXT NOT NULL DEFAULT '',
	api_key  TEXT NOT NULL DEFAULT '',
	models   TEXT NOT NULL DEFAULT '',
	enabled  INTEGER NOT NULL DEFAULT 1,
	sort     INTEGER NOT NULL DEFAULT 0
);
CREATE TABLE IF NOT EXISTS provider_users (
	provider_id INTEGER NOT NULL REFERENCES providers(id) ON DELETE CASCADE,
	user_id     INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	PRIMARY KEY (provider_id, user_id)
);
CREATE TABLE IF NOT EXISTS access_log (
	id         INTEGER PRIMARY KEY AUTOINCREMENT,
	created_at INTEGER NOT NULL,
	user_id    INTEGER NOT NULL DEFAULT 0,
	username   TEXT NOT NULL DEFAULT '',
	event      TEXT NOT NULL,
	ip         TEXT NOT NULL DEFAULT '',
	country    TEXT NOT NULL DEFAULT '',
	user_agent TEXT NOT NULL DEFAULT '',
	detail     TEXT NOT NULL DEFAULT '',
	ok         INTEGER NOT NULL DEFAULT 1
);
CREATE INDEX IF NOT EXISTS access_log_time ON access_log(created_at DESC);
CREATE INDEX IF NOT EXISTS access_log_user ON access_log(user_id, created_at DESC);
CREATE TABLE IF NOT EXISTS settings (
	key   TEXT PRIMARY KEY,
	value TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS chats (
	id         INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	title      TEXT NOT NULL DEFAULT '',
	model      TEXT NOT NULL DEFAULT '',
	created_at INTEGER NOT NULL,
	updated_at INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS chats_user ON chats(user_id, updated_at DESC);
CREATE TABLE IF NOT EXISTS messages (
	id          INTEGER PRIMARY KEY AUTOINCREMENT,
	chat_id     INTEGER NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
	role        TEXT NOT NULL,
	content     TEXT NOT NULL DEFAULT '',
	reasoning   TEXT NOT NULL DEFAULT '',
	attachments TEXT NOT NULL DEFAULT '',
	model       TEXT NOT NULL DEFAULT '',
	raw         TEXT NOT NULL DEFAULT '',
	tokens_in   INTEGER NOT NULL DEFAULT 0,
	tokens_out  INTEGER NOT NULL DEFAULT 0,
	error       TEXT NOT NULL DEFAULT '',
	created_at  INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS messages_chat ON messages(chat_id, id);
CREATE TABLE IF NOT EXISTS usage (
	user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	month      TEXT NOT NULL,
	tokens_in  INTEGER NOT NULL DEFAULT 0,
	tokens_out INTEGER NOT NULL DEFAULT 0,
	requests   INTEGER NOT NULL DEFAULT 0,
	PRIMARY KEY (user_id, month)
);
`

// columns added after v0.1; applied to existing databases on start.
var migrations = []struct{ table, column, def string }{
	{"providers", "owner_id", "INTEGER NOT NULL DEFAULT 0"},    // 0 = shared, else the user who owns it
	{"providers", "visibility", "TEXT NOT NULL DEFAULT 'all'"}, // shared providers: "all" or "users"
	{"usage", "own_in", "INTEGER NOT NULL DEFAULT 0"},          // tokens on users' own keys
	{"usage", "own_out", "INTEGER NOT NULL DEFAULT 0"},
	{"providers", "preset", "TEXT NOT NULL DEFAULT ''"},       // catalog preset ID; with empty models it follows the catalog
	{"messages", "think_ms", "INTEGER NOT NULL DEFAULT 0"},    // request start to first answer token
	{"messages", "duration_ms", "INTEGER NOT NULL DEFAULT 0"}, // request start to end
	{"users", "last_seen_at", "INTEGER NOT NULL DEFAULT 0"},
	{"users", "last_ip", "TEXT NOT NULL DEFAULT ''"},
	{"users", "last_country", "TEXT NOT NULL DEFAULT ''"},
}

type Store struct {
	db  *sql.DB
	box *sealer
}

type User struct {
	ID         int64  `json:"id"`
	Username   string `json:"username"`
	PassHash   string `json:"-"`
	IsAdmin    bool   `json:"is_admin"`
	Disabled   bool   `json:"disabled"`
	TokenLimit int64  `json:"token_limit"`
	CreatedAt  int64  `json:"created_at"`
	LastSeenAt int64  `json:"last_seen_at"`
	LastIP     string `json:"last_ip"`
	Country    string `json:"last_country"`
}

type Provider struct {
	ID         int64    `json:"id"`
	Name       string   `json:"name"`
	Kind       string   `json:"kind"` // "openai" (OpenAI-compatible) or "anthropic"
	BaseURL    string   `json:"base_url"`
	APIKey     string   `json:"api_key,omitempty"`
	Models     []string `json:"models"`
	Enabled    bool     `json:"enabled"`
	Sort       int64    `json:"sort"`
	Preset     string   `json:"preset"`
	AutoModels bool     `json:"auto_models"` // models come from the catalog preset
	OwnerID    int64    `json:"owner_id"`    // 0 = shared provider managed by admins
	Visibility string   `json:"visibility"`  // shared only: "all" or "users"
	UserIDs    []int64  `json:"user_ids"`    // who may use it when Visibility is "users"
	KeyErr     error    `json:"-"`           // key could not be decrypted
	guard      bool     // block private-network targets (users' own providers)
}

type Chat struct {
	ID        int64  `json:"id"`
	UserID    int64  `json:"-"`
	Title     string `json:"title"`
	Model     string `json:"model"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

type Attachment struct {
	Name string `json:"name"`
	Text string `json:"text"`
}

type Message struct {
	ID          int64        `json:"id"`
	ChatID      int64        `json:"-"`
	Role        string       `json:"role"`
	Content     string       `json:"content"`
	Reasoning   string       `json:"reasoning,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
	Model       string       `json:"model,omitempty"`
	Raw         string       `json:"-"` // provider-native assistant turn, replayed verbatim
	TokensIn    int64        `json:"tokens_in,omitempty"`
	TokensOut   int64        `json:"tokens_out,omitempty"`
	ThinkMs     int64        `json:"think_ms,omitempty"`
	DurationMs  int64        `json:"duration_ms,omitempty"`
	Error       string       `json:"error,omitempty"`
	CreatedAt   int64        `json:"created_at"`
}

type UsageRow struct {
	UserID    int64  `json:"user_id"`
	Username  string `json:"username"`
	TokensIn  int64  `json:"tokens_in"`
	TokensOut int64  `json:"tokens_out"`
	OwnTokens int64  `json:"own_tokens"` // on the user's own keys, not counted against the limit
	Requests  int64  `json:"requests"`
	Limit     int64  `json:"token_limit"`
}

var errNotFound = errors.New("not found")

func openStore(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path+"?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(schema); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	for _, m := range migrations {
		var n int
		db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info(?) WHERE name=?`, m.table, m.column).Scan(&n)
		if n == 0 {
			if _, err := db.Exec(fmt.Sprintf(`ALTER TABLE %s ADD COLUMN %s %s`, m.table, m.column, m.def)); err != nil {
				return nil, fmt.Errorf("migrate %s.%s: %w", m.table, m.column, err)
			}
		}
	}
	db.Exec(`CREATE INDEX IF NOT EXISTS providers_owner ON providers(owner_id)`)
	return &Store{db: db}, nil
}

// SealKeys encrypts any plain-text API keys once a secret is configured.
func (s *Store) SealKeys() error {
	if s.box == nil {
		return nil
	}
	rows, err := s.db.Query(`SELECT id, api_key FROM providers WHERE api_key<>'' AND api_key NOT LIKE 'enc:%'`)
	if err != nil {
		return err
	}
	keys := map[int64]string{}
	for rows.Next() {
		var id int64
		var k string
		rows.Scan(&id, &k)
		keys[id] = k
	}
	rows.Close()
	for id, k := range keys {
		if _, err := s.db.Exec(`UPDATE providers SET api_key=? WHERE id=?`, s.box.seal(k), id); err != nil {
			return err
		}
	}
	return nil
}

// LinkPresets ties providers from before the catalog to their preset, and lets
// the ones still on an old seeded model list follow the catalog from now on.
func (s *Store) LinkPresets() error {
	rows, err := s.db.Query(`SELECT id, kind, base_url, models FROM providers WHERE preset=''`)
	if err != nil {
		return err
	}
	type link struct {
		id     int64
		preset string
		auto   bool
	}
	var links []link
	for rows.Next() {
		var id int64
		var kind, base, models string
		rows.Scan(&id, &kind, &base, &models)
		if pr := presetFor(kind, base); pr != nil {
			links = append(links, link{id, pr.ID, isLegacySeed(splitModels(models))})
		}
	}
	rows.Close()
	for _, l := range links {
		q := `UPDATE providers SET preset=? WHERE id=?`
		if l.auto {
			q = `UPDATE providers SET preset=?, models='' WHERE id=?`
		}
		if _, err := s.db.Exec(q, l.preset, l.id); err != nil {
			return err
		}
	}
	return nil
}

// --- settings ---

func (s *Store) Setting(key string) string {
	var v string
	s.db.QueryRow(`SELECT value FROM settings WHERE key=?`, key).Scan(&v)
	return v
}

func (s *Store) SetSetting(key, value string) error {
	_, err := s.db.Exec(`INSERT INTO settings(key, value) VALUES(?,?) ON CONFLICT(key) DO UPDATE SET value=excluded.value`, key, value)
	return err
}

func now() int64 { return time.Now().Unix() }

func month() string { return time.Now().Format("2006-01") }

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

// --- users ---

const userCols = `id, username, pass_hash, is_admin, disabled, token_limit, created_at, last_seen_at, last_ip, last_country`

func scanUser(row interface{ Scan(...any) error }) (*User, error) {
	var u User
	err := row.Scan(&u.ID, &u.Username, &u.PassHash, &u.IsAdmin, &u.Disabled, &u.TokenLimit, &u.CreatedAt, &u.LastSeenAt, &u.LastIP, &u.Country)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errNotFound
	}
	return &u, err
}

func (s *Store) CountUsers() (n int, err error) {
	err = s.db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&n)
	return
}

func (s *Store) CreateUser(username, passHash string, admin bool, limit int64) (*User, error) {
	res, err := s.db.Exec(`INSERT INTO users(username, pass_hash, is_admin, token_limit, created_at) VALUES(?,?,?,?,?)`,
		username, passHash, b2i(admin), limit, now())
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return s.UserByID(id)
}

func (s *Store) UserByID(id int64) (*User, error) {
	return scanUser(s.db.QueryRow(`SELECT `+userCols+` FROM users WHERE id=?`, id))
}

func (s *Store) UserByName(name string) (*User, error) {
	return scanUser(s.db.QueryRow(`SELECT `+userCols+` FROM users WHERE username=?`, name))
}

func (s *Store) ListUsers() ([]*User, error) {
	rows, err := s.db.Query(`SELECT ` + userCols + ` FROM users ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*User
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

func (s *Store) UpdateUser(u *User) error {
	_, err := s.db.Exec(`UPDATE users SET is_admin=?, disabled=?, token_limit=? WHERE id=?`,
		b2i(u.IsAdmin), b2i(u.Disabled), u.TokenLimit, u.ID)
	return err
}

func (s *Store) SetPassword(id int64, hash string) error {
	_, err := s.db.Exec(`UPDATE users SET pass_hash=? WHERE id=?`, hash, id)
	return err
}

func (s *Store) DeleteUser(id int64) error {
	if _, err := s.db.Exec(`DELETE FROM providers WHERE owner_id=?`, id); err != nil {
		return err
	}
	_, err := s.db.Exec(`DELETE FROM users WHERE id=?`, id)
	return err
}

func (s *Store) CountAdmins() (n int, err error) {
	err = s.db.QueryRow(`SELECT COUNT(*) FROM users WHERE is_admin=1 AND disabled=0`).Scan(&n)
	return
}

func (s *Store) TouchUser(id int64, ip, country string) {
	s.db.Exec(`UPDATE users SET last_seen_at=?, last_ip=?, last_country=? WHERE id=?`, now(), ip, country, id)
}

// --- access log ---

const accessLogKeep = 20000 // rows kept; older ones are pruned

type AccessEntry struct {
	ID        int64  `json:"id"`
	CreatedAt int64  `json:"created_at"`
	UserID    int64  `json:"user_id"`
	Username  string `json:"username"`
	Event     string `json:"event"` // login, login_failed, chat
	IP        string `json:"ip"`
	Country   string `json:"country"`
	UserAgent string `json:"user_agent"`
	Detail    string `json:"detail"`
	OK        bool   `json:"ok"`
}

func (s *Store) LogAccess(e *AccessEntry) {
	e.CreatedAt = now()
	res, err := s.db.Exec(`INSERT INTO access_log(created_at, user_id, username, event, ip, country, user_agent, detail, ok) VALUES(?,?,?,?,?,?,?,?,?)`,
		e.CreatedAt, e.UserID, e.Username, e.Event, e.IP, e.Country, e.UserAgent, e.Detail, b2i(e.OK))
	if err != nil {
		return
	}
	if id, _ := res.LastInsertId(); id%500 == 0 {
		s.db.Exec(`DELETE FROM access_log WHERE id <= ?`, id-accessLogKeep)
	}
}

// AccessLog lists the newest entries, optionally for one user, older than before (0 = newest).
func (s *Store) AccessLog(userID, before int64, limit int) ([]AccessEntry, error) {
	q := `SELECT id, created_at, user_id, username, event, ip, country, user_agent, detail, ok FROM access_log WHERE 1=1`
	var args []any
	if userID > 0 {
		q += ` AND user_id=?`
		args = append(args, userID)
	}
	if before > 0 {
		q += ` AND id<?`
		args = append(args, before)
	}
	q += ` ORDER BY id DESC LIMIT ?`
	args = append(args, limit)
	rows, err := s.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []AccessEntry{}
	for rows.Next() {
		var e AccessEntry
		if err := rows.Scan(&e.ID, &e.CreatedAt, &e.UserID, &e.Username, &e.Event, &e.IP, &e.Country, &e.UserAgent, &e.Detail, &e.OK); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// --- sessions ---

func (s *Store) CreateSession(tokenHash string, userID int64, ttl time.Duration) error {
	_, err := s.db.Exec(`INSERT INTO sessions(token_hash, user_id, expires_at) VALUES(?,?,?)`,
		tokenHash, userID, time.Now().Add(ttl).Unix())
	return err
}

func (s *Store) SessionUser(tokenHash string) (*User, error) {
	return scanUser(s.db.QueryRow(`SELECT u.id, u.username, u.pass_hash, u.is_admin, u.disabled, u.token_limit, u.created_at,
		u.last_seen_at, u.last_ip, u.last_country
		FROM sessions s JOIN users u ON u.id = s.user_id WHERE s.token_hash=? AND s.expires_at>?`, tokenHash, now()))
}

func (s *Store) DeleteSession(tokenHash string) error {
	_, err := s.db.Exec(`DELETE FROM sessions WHERE token_hash=?`, tokenHash)
	return err
}

func (s *Store) DeleteUserSessions(userID int64, exceptHash string) error {
	_, err := s.db.Exec(`DELETE FROM sessions WHERE user_id=? AND token_hash<>?`, userID, exceptHash)
	return err
}

func (s *Store) PurgeSessions() {
	s.db.Exec(`DELETE FROM sessions WHERE expires_at<=?`, now())
}

// --- providers ---

func splitModels(s string) []string {
	var out []string
	for _, m := range strings.FieldsFunc(s, func(r rune) bool { return r == '\n' || r == ',' }) {
		if m = strings.TrimSpace(m); m != "" {
			out = append(out, m)
		}
	}
	return out
}

const providerCols = `id, name, kind, base_url, api_key, models, enabled, sort, owner_id, visibility, preset`

func (s *Store) queryProviders(where string, args ...any) ([]*Provider, error) {
	rows, err := s.db.Query(`SELECT `+providerCols+` FROM providers `+where+` ORDER BY sort, id`, args...)
	if err != nil {
		return nil, err
	}
	var out []*Provider
	byID := map[int64]*Provider{}
	for rows.Next() {
		var p Provider
		var models string
		if err := rows.Scan(&p.ID, &p.Name, &p.Kind, &p.BaseURL, &p.APIKey, &models, &p.Enabled, &p.Sort, &p.OwnerID, &p.Visibility, &p.Preset); err != nil {
			rows.Close()
			return nil, err
		}
		p.Models = splitModels(models)
		if pr := presetByID(p.Preset); pr != nil && len(p.Models) == 0 {
			p.AutoModels, p.Models = true, slices.Clone(pr.Models)
		}
		p.APIKey, p.KeyErr = s.box.open(p.APIKey)
		p.UserIDs = []int64{}
		out = append(out, &p)
		byID[p.ID] = &p
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	links, err := s.db.Query(`SELECT provider_id, user_id FROM provider_users ORDER BY user_id`)
	if err != nil {
		return nil, err
	}
	defer links.Close()
	for links.Next() {
		var pid, uid int64
		links.Scan(&pid, &uid)
		if p := byID[pid]; p != nil {
			p.UserIDs = append(p.UserIDs, uid)
		}
	}
	return out, links.Err()
}

// SharedProviders are the admin-managed ones.
func (s *Store) SharedProviders() ([]*Provider, error) {
	return s.queryProviders(`WHERE owner_id=0`)
}

// UsableProviders are the shared ones plus the user's own.
func (s *Store) UsableProviders(userID int64) ([]*Provider, error) {
	return s.queryProviders(`WHERE owner_id=0 OR owner_id=?`, userID)
}

func (s *Store) OwnProviders(userID int64) ([]*Provider, error) {
	return s.queryProviders(`WHERE owner_id=?`, userID)
}

func (s *Store) Provider(id int64) (*Provider, error) {
	ps, err := s.queryProviders(`WHERE id=?`, id)
	if err != nil {
		return nil, err
	}
	if len(ps) == 0 {
		return nil, errNotFound
	}
	return ps[0], nil
}

func (s *Store) SaveProvider(p *Provider) error {
	models := strings.Join(p.Models, "\n")
	if p.AutoModels && presetByID(p.Preset) != nil {
		models = ""
	}
	if p.Visibility == "" {
		p.Visibility = "all"
	}
	key := s.box.seal(p.APIKey)
	if p.KeyErr != nil && p.APIKey == "" && p.ID != 0 {
		// Undecryptable key and no replacement: keep what is stored.
		s.db.QueryRow(`SELECT api_key FROM providers WHERE id=?`, p.ID).Scan(&key)
	}
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if p.ID == 0 {
		if p.Sort == 0 {
			tx.QueryRow(`SELECT COALESCE(MAX(sort),0)+1 FROM providers`).Scan(&p.Sort) // new ones go last
		}
		res, err := tx.Exec(`INSERT INTO providers(name, kind, base_url, api_key, models, enabled, sort, owner_id, visibility, preset) VALUES(?,?,?,?,?,?,?,?,?,?)`,
			p.Name, p.Kind, p.BaseURL, key, models, b2i(p.Enabled), p.Sort, p.OwnerID, p.Visibility, p.Preset)
		if err != nil {
			return err
		}
		p.ID, _ = res.LastInsertId()
	} else if _, err := tx.Exec(`UPDATE providers SET name=?, kind=?, base_url=?, api_key=?, models=?, enabled=?, sort=?, visibility=?, preset=? WHERE id=?`,
		p.Name, p.Kind, p.BaseURL, key, models, b2i(p.Enabled), p.Sort, p.Visibility, p.Preset, p.ID); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM provider_users WHERE provider_id=?`, p.ID); err != nil {
		return err
	}
	for _, uid := range p.UserIDs {
		if _, err := tx.Exec(`INSERT OR IGNORE INTO provider_users(provider_id, user_id) SELECT ?, id FROM users WHERE id=?`, p.ID, uid); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) DeleteProvider(id int64) error {
	_, err := s.db.Exec(`DELETE FROM providers WHERE id=?`, id)
	return err
}

func (s *Store) CountProviders() (n int, err error) {
	err = s.db.QueryRow(`SELECT COUNT(*) FROM providers WHERE owner_id=0`).Scan(&n)
	return
}

// --- chats ---

func (s *Store) CreateChat(userID int64, model string) (*Chat, error) {
	t := now()
	res, err := s.db.Exec(`INSERT INTO chats(user_id, model, created_at, updated_at) VALUES(?,?,?,?)`, userID, model, t, t)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return &Chat{ID: id, UserID: userID, Model: model, CreatedAt: t, UpdatedAt: t}, nil
}

func (s *Store) Chat(userID, id int64) (*Chat, error) {
	var c Chat
	err := s.db.QueryRow(`SELECT id, user_id, title, model, created_at, updated_at FROM chats WHERE id=? AND user_id=?`, id, userID).
		Scan(&c.ID, &c.UserID, &c.Title, &c.Model, &c.CreatedAt, &c.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errNotFound
	}
	return &c, err
}

func (s *Store) ListChats(userID int64) ([]*Chat, error) {
	rows, err := s.db.Query(`SELECT id, user_id, title, model, created_at, updated_at FROM chats WHERE user_id=? ORDER BY updated_at DESC, id DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*Chat{}
	for rows.Next() {
		var c Chat
		if err := rows.Scan(&c.ID, &c.UserID, &c.Title, &c.Model, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, &c)
	}
	return out, rows.Err()
}

func (s *Store) UpdateChat(c *Chat) error {
	_, err := s.db.Exec(`UPDATE chats SET title=?, model=?, updated_at=? WHERE id=? AND user_id=?`, c.Title, c.Model, c.UpdatedAt, c.ID, c.UserID)
	return err
}

func (s *Store) DeleteChat(userID, id int64) error {
	_, err := s.db.Exec(`DELETE FROM chats WHERE id=? AND user_id=?`, id, userID)
	return err
}

// --- messages ---

func (s *Store) Messages(chatID int64) ([]*Message, error) {
	rows, err := s.db.Query(`SELECT id, chat_id, role, content, reasoning, attachments, model, raw, tokens_in, tokens_out, think_ms, duration_ms, error, created_at
		FROM messages WHERE chat_id=? ORDER BY id`, chatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*Message{}
	for rows.Next() {
		var m Message
		var att string
		if err := rows.Scan(&m.ID, &m.ChatID, &m.Role, &m.Content, &m.Reasoning, &att, &m.Model, &m.Raw, &m.TokensIn, &m.TokensOut, &m.ThinkMs, &m.DurationMs, &m.Error, &m.CreatedAt); err != nil {
			return nil, err
		}
		m.Attachments = decodeAttachments(att)
		out = append(out, &m)
	}
	return out, rows.Err()
}

func (s *Store) AddMessage(m *Message) error {
	m.CreatedAt = now()
	res, err := s.db.Exec(`INSERT INTO messages(chat_id, role, content, reasoning, attachments, model, raw, tokens_in, tokens_out, think_ms, duration_ms, error, created_at)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		m.ChatID, m.Role, m.Content, m.Reasoning, encodeAttachments(m.Attachments), m.Model, m.Raw, m.TokensIn, m.TokensOut, m.ThinkMs, m.DurationMs, m.Error, m.CreatedAt)
	if err != nil {
		return err
	}
	m.ID, _ = res.LastInsertId()
	return nil
}

func (s *Store) DeleteMessage(id int64) error {
	_, err := s.db.Exec(`DELETE FROM messages WHERE id=?`, id)
	return err
}

// --- usage ---

// AddUsage records a request. Tokens on the user's own key (own=true) are kept
// apart so they never count against the admin-set limit.
func (s *Store) AddUsage(userID, in, out int64, own bool) error {
	var sin, sout, oin, oout int64
	if own {
		oin, oout = in, out
	} else {
		sin, sout = in, out
	}
	_, err := s.db.Exec(`INSERT INTO usage(user_id, month, tokens_in, tokens_out, own_in, own_out, requests) VALUES(?,?,?,?,?,?,1)
		ON CONFLICT(user_id, month) DO UPDATE SET tokens_in=tokens_in+excluded.tokens_in, tokens_out=tokens_out+excluded.tokens_out,
		own_in=own_in+excluded.own_in, own_out=own_out+excluded.own_out, requests=requests+1`,
		userID, month(), sin, sout, oin, oout)
	return err
}

func (s *Store) MonthTokens(userID int64) (n int64, err error) {
	err = s.db.QueryRow(`SELECT COALESCE(SUM(tokens_in+tokens_out),0) FROM usage WHERE user_id=? AND month=?`, userID, month()).Scan(&n)
	return
}

func (s *Store) Usage(m string) ([]UsageRow, error) {
	rows, err := s.db.Query(`SELECT u.id, u.username, COALESCE(g.tokens_in,0), COALESCE(g.tokens_out,0), COALESCE(g.own_in+g.own_out,0), COALESCE(g.requests,0), u.token_limit
		FROM users u LEFT JOIN usage g ON g.user_id=u.id AND g.month=? ORDER BY COALESCE(g.tokens_in+g.tokens_out,0) DESC, u.id`, m)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []UsageRow{}
	for rows.Next() {
		var r UsageRow
		if err := rows.Scan(&r.UserID, &r.Username, &r.TokensIn, &r.TokensOut, &r.OwnTokens, &r.Requests, &r.Limit); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
