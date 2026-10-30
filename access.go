package main

import (
	"net/http"
	"strconv"
	"sync"
	"time"
)

// seenCache throttles last-seen updates to one write per user per minute
// (or immediately when the IP changes).
type seenCache struct {
	mu   sync.Mutex
	last map[int64]seenMark
}

type seenMark struct {
	ip string
	at time.Time
}

func (a *App) touch(u *User, r *http.Request) {
	ip := clientIP(r)
	a.seen.mu.Lock()
	m := a.seen.last[u.ID]
	due := m.ip != ip || time.Since(m.at) > time.Minute
	if due {
		a.seen.last[u.ID] = seenMark{ip, time.Now()}
	}
	a.seen.mu.Unlock()
	if due {
		a.store.TouchUser(u.ID, ip, clientCountry(r))
	}
}

func (a *App) logAccess(r *http.Request, userID int64, username, event, detail string, ok bool) {
	ua := r.UserAgent()
	if len(ua) > 300 {
		ua = ua[:300]
	}
	a.store.LogAccess(&AccessEntry{UserID: userID, Username: username, Event: event, IP: clientIP(r),
		Country: clientCountry(r), UserAgent: ua, Detail: detail, OK: ok})
}

func (a *App) handleAccessLog(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	uid, _ := strconv.ParseInt(q.Get("user"), 10, 64)
	before, _ := strconv.ParseInt(q.Get("before"), 10, 64)
	rows, err := a.store.AccessLog(uid, before, 100)
	if err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, rows)
}
