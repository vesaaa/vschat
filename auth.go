package main

import (
	"context"
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	sessionCookie = "vschat_session"
	sessionTTL    = 30 * 24 * time.Hour
	pbkdf2Iter    = 210000
)

func randomToken(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func hashToken(t string) string {
	h := sha256.Sum256([]byte(t))
	return hex.EncodeToString(h[:])
}

// hashPassword returns "pbkdf2-sha256$iter$salt$hash".
func hashPassword(pw string) string {
	salt := make([]byte, 16)
	rand.Read(salt)
	key, _ := pbkdf2.Key(sha256.New, pw, salt, pbkdf2Iter, 32)
	return fmt.Sprintf("pbkdf2-sha256$%d$%s$%s", pbkdf2Iter,
		base64.RawStdEncoding.EncodeToString(salt), base64.RawStdEncoding.EncodeToString(key))
}

func checkPassword(stored, pw string) bool {
	parts := strings.Split(stored, "$")
	if len(parts) != 4 || parts[0] != "pbkdf2-sha256" {
		return false
	}
	iter, err := strconv.Atoi(parts[1])
	if err != nil {
		return false
	}
	salt, err1 := base64.RawStdEncoding.DecodeString(parts[2])
	want, err2 := base64.RawStdEncoding.DecodeString(parts[3])
	if err1 != nil || err2 != nil {
		return false
	}
	got, err := pbkdf2.Key(sha256.New, pw, salt, iter, len(want))
	return err == nil && subtle.ConstantTimeCompare(got, want) == 1
}

type ctxKey int

const userKey ctxKey = 0

func currentUser(r *http.Request) *User {
	u, _ := r.Context().Value(userKey).(*User)
	return u
}

func sessionHash(r *http.Request) string {
	c, err := r.Cookie(sessionCookie)
	if err != nil || c.Value == "" {
		return ""
	}
	return hashToken(c.Value)
}

func (a *App) requireUser(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h := sessionHash(r)
		if h == "" {
			writeErr(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		u, err := a.store.SessionUser(h)
		if err != nil || u.Disabled {
			writeErr(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		next(w, r.WithContext(context.WithValue(r.Context(), userKey, u)))
	}
}

func (a *App) requireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return a.requireUser(func(w http.ResponseWriter, r *http.Request) {
		if !currentUser(r).IsAdmin {
			writeErr(w, http.StatusForbidden, "forbidden")
			return
		}
		next(w, r)
	})
}

// loginLimiter slows down password guessing: 10 failures per IP per 15 minutes.
type loginLimiter struct {
	mu   sync.Mutex
	fail map[string][]time.Time
}

func (l *loginLimiter) blocked(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	cut := time.Now().Add(-15 * time.Minute)
	ts := l.fail[ip][:0]
	for _, t := range l.fail[ip] {
		if t.After(cut) {
			ts = append(ts, t)
		}
	}
	l.fail[ip] = ts
	return len(ts) >= 10
}

func (l *loginLimiter) record(ip string) {
	l.mu.Lock()
	l.fail[ip] = append(l.fail[ip], time.Now())
	l.mu.Unlock()
}

// trustProxy is set when VSCHAT_TRUST_PROXY=1, i.e. VsChat sits behind a reverse
// proxy that sets X-Forwarded-For. Otherwise the header is ignored (spoofable).
var trustProxy bool

func clientIP(r *http.Request) string {
	if f := r.Header.Get("X-Forwarded-For"); trustProxy && f != "" {
		return strings.TrimSpace(strings.Split(f, ",")[0])
	}
	host := r.RemoteAddr
	if i := strings.LastIndex(host, ":"); i > 0 {
		host = host[:i]
	}
	return host
}

func (a *App) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req struct{ Username, Password string }
	if !readJSON(w, r, &req) {
		return
	}
	ip := clientIP(r)
	if a.limiter.blocked(ip) {
		writeErr(w, http.StatusTooManyRequests, "too many attempts, try later")
		return
	}
	u, err := a.store.UserByName(strings.TrimSpace(req.Username))
	if err != nil || u.Disabled || !checkPassword(u.PassHash, req.Password) {
		a.limiter.record(ip)
		writeErr(w, http.StatusUnauthorized, "invalid username or password")
		return
	}
	tok := randomToken(32)
	if err := a.store.CreateSession(hashToken(tok), u.ID, sessionTTL); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name: sessionCookie, Value: tok, Path: "/", HttpOnly: true,
		SameSite: http.SameSiteLaxMode, Secure: r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https",
		MaxAge: int(sessionTTL.Seconds()),
	})
	writeJSON(w, u)
}

func (a *App) handleLogout(w http.ResponseWriter, r *http.Request) {
	if h := sessionHash(r); h != "" {
		a.store.DeleteSession(h)
	}
	http.SetCookie(w, &http.Cookie{Name: sessionCookie, Value: "", Path: "/", MaxAge: -1})
	w.WriteHeader(http.StatusNoContent)
}

func (a *App) handleMe(w http.ResponseWriter, r *http.Request) {
	u := currentUser(r)
	used, _ := a.store.MonthTokens(u.ID)
	writeJSON(w, map[string]any{"user": u, "month_tokens": used, "user_keys": a.userKeysAllowed()})
}

func (a *App) handleChangePassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Old string `json:"old"`
		New string `json:"new"`
	}
	if !readJSON(w, r, &req) {
		return
	}
	u := currentUser(r)
	if !checkPassword(u.PassHash, req.Old) {
		writeErr(w, http.StatusBadRequest, "wrong current password")
		return
	}
	if len(req.New) < 6 {
		writeErr(w, http.StatusBadRequest, "password must be at least 6 characters")
		return
	}
	a.store.SetPassword(u.ID, hashPassword(req.New))
	a.store.DeleteUserSessions(u.ID, sessionHash(r))
	w.WriteHeader(http.StatusNoContent)
}
