package main

import (
	"database/sql"
	"errors"
	"fmt"
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

type Store struct{ db *sql.DB }

type User struct {
	ID         int64  `json:"id"`
	Username   string `json:"username"`
	PassHash   string `json:"-"`
	IsAdmin    bool   `json:"is_admin"`
	Disabled   bool   `json:"disabled"`
	TokenLimit int64  `json:"token_limit"`
	CreatedAt  int64  `json:"created_at"`
}

type Provider struct {
	ID      int64    `json:"id"`
	Name    string   `json:"name"`
	Kind    string   `json:"kind"` // "openai" (OpenAI-compatible) or "anthropic"
	BaseURL string   `json:"base_url"`
	APIKey  string   `json:"api_key,omitempty"`
	Models  []string `json:"models"`
	Enabled bool     `json:"enabled"`
	Sort    int64    `json:"sort"`
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
	Error       string       `json:"error,omitempty"`
	CreatedAt   int64        `json:"created_at"`
}

type UsageRow struct {
	UserID    int64  `json:"user_id"`
	Username  string `json:"username"`
	TokensIn  int64  `json:"tokens_in"`
	TokensOut int64  `json:"tokens_out"`
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
	return &Store{db: db}, nil
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

const userCols = `id, username, pass_hash, is_admin, disabled, token_limit, created_at`

func scanUser(row interface{ Scan(...any) error }) (*User, error) {
	var u User
	err := row.Scan(&u.ID, &u.Username, &u.PassHash, &u.IsAdmin, &u.Disabled, &u.TokenLimit, &u.CreatedAt)
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
	_, err := s.db.Exec(`DELETE FROM users WHERE id=?`, id)
	return err
}

func (s *Store) CountAdmins() (n int, err error) {
	err = s.db.QueryRow(`SELECT COUNT(*) FROM users WHERE is_admin=1 AND disabled=0`).Scan(&n)
	return
}

// --- sessions ---

func (s *Store) CreateSession(tokenHash string, userID int64, ttl time.Duration) error {
	_, err := s.db.Exec(`INSERT INTO sessions(token_hash, user_id, expires_at) VALUES(?,?,?)`,
		tokenHash, userID, time.Now().Add(ttl).Unix())
	return err
}

func (s *Store) SessionUser(tokenHash string) (*User, error) {
	return scanUser(s.db.QueryRow(`SELECT u.id, u.username, u.pass_hash, u.is_admin, u.disabled, u.token_limit, u.created_at
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

func (s *Store) ListProviders() ([]*Provider, error) {
	rows, err := s.db.Query(`SELECT id, name, kind, base_url, api_key, models, enabled, sort FROM providers ORDER BY sort, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*Provider
	for rows.Next() {
		var p Provider
		var models string
		if err := rows.Scan(&p.ID, &p.Name, &p.Kind, &p.BaseURL, &p.APIKey, &models, &p.Enabled, &p.Sort); err != nil {
			return nil, err
		}
		p.Models = splitModels(models)
		out = append(out, &p)
	}
	return out, rows.Err()
}

func (s *Store) Provider(id int64) (*Provider, error) {
	ps, err := s.ListProviders()
	if err != nil {
		return nil, err
	}
	for _, p := range ps {
		if p.ID == id {
			return p, nil
		}
	}
	return nil, errNotFound
}

func (s *Store) SaveProvider(p *Provider) error {
	models := strings.Join(p.Models, "\n")
	if p.ID == 0 {
		res, err := s.db.Exec(`INSERT INTO providers(name, kind, base_url, api_key, models, enabled, sort) VALUES(?,?,?,?,?,?,?)`,
			p.Name, p.Kind, p.BaseURL, p.APIKey, models, b2i(p.Enabled), p.Sort)
		if err != nil {
			return err
		}
		p.ID, _ = res.LastInsertId()
		return nil
	}
	_, err := s.db.Exec(`UPDATE providers SET name=?, kind=?, base_url=?, api_key=?, models=?, enabled=?, sort=? WHERE id=?`,
		p.Name, p.Kind, p.BaseURL, p.APIKey, models, b2i(p.Enabled), p.Sort, p.ID)
	return err
}

func (s *Store) DeleteProvider(id int64) error {
	_, err := s.db.Exec(`DELETE FROM providers WHERE id=?`, id)
	return err
}

func (s *Store) CountProviders() (n int, err error) {
	err = s.db.QueryRow(`SELECT COUNT(*) FROM providers`).Scan(&n)
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
	rows, err := s.db.Query(`SELECT id, chat_id, role, content, reasoning, attachments, model, raw, tokens_in, tokens_out, error, created_at
		FROM messages WHERE chat_id=? ORDER BY id`, chatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*Message{}
	for rows.Next() {
		var m Message
		var att string
		if err := rows.Scan(&m.ID, &m.ChatID, &m.Role, &m.Content, &m.Reasoning, &att, &m.Model, &m.Raw, &m.TokensIn, &m.TokensOut, &m.Error, &m.CreatedAt); err != nil {
			return nil, err
		}
		m.Attachments = decodeAttachments(att)
		out = append(out, &m)
	}
	return out, rows.Err()
}

func (s *Store) AddMessage(m *Message) error {
	m.CreatedAt = now()
	res, err := s.db.Exec(`INSERT INTO messages(chat_id, role, content, reasoning, attachments, model, raw, tokens_in, tokens_out, error, created_at)
		VALUES(?,?,?,?,?,?,?,?,?,?,?)`,
		m.ChatID, m.Role, m.Content, m.Reasoning, encodeAttachments(m.Attachments), m.Model, m.Raw, m.TokensIn, m.TokensOut, m.Error, m.CreatedAt)
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

func (s *Store) AddUsage(userID, in, out int64) error {
	_, err := s.db.Exec(`INSERT INTO usage(user_id, month, tokens_in, tokens_out, requests) VALUES(?,?,?,?,1)
		ON CONFLICT(user_id, month) DO UPDATE SET tokens_in=tokens_in+excluded.tokens_in, tokens_out=tokens_out+excluded.tokens_out, requests=requests+1`,
		userID, month(), in, out)
	return err
}

func (s *Store) MonthTokens(userID int64) (n int64, err error) {
	err = s.db.QueryRow(`SELECT COALESCE(SUM(tokens_in+tokens_out),0) FROM usage WHERE user_id=? AND month=?`, userID, month()).Scan(&n)
	return
}

func (s *Store) Usage(m string) ([]UsageRow, error) {
	rows, err := s.db.Query(`SELECT u.id, u.username, COALESCE(g.tokens_in,0), COALESCE(g.tokens_out,0), COALESCE(g.requests,0), u.token_limit
		FROM users u LEFT JOIN usage g ON g.user_id=u.id AND g.month=? ORDER BY COALESCE(g.tokens_in+g.tokens_out,0) DESC, u.id`, m)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []UsageRow{}
	for rows.Next() {
		var r UsageRow
		if err := rows.Scan(&r.UserID, &r.Username, &r.TokensIn, &r.TokensOut, &r.Requests, &r.Limit); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
