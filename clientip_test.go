package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClientIP(t *testing.T) {
	req := func(peer string, h map[string]string) *http.Request {
		r := httptest.NewRequest("GET", "/", nil)
		r.RemoteAddr = peer
		for k, v := range h {
			r.Header.Set(k, v)
		}
		return r
	}
	cf := map[string]string{"CF-Connecting-IP": "203.0.113.10", "X-Forwarded-For": "203.0.113.10, 104.16.0.1", "X-Real-IP": "104.16.0.1", "CF-IPCountry": "cn"}
	cases := []struct {
		mode, peer string
		h          map[string]string
		want, cc   string
	}{
		{"auto", "104.16.0.1:443", cf, "203.0.113.10", "CN"},                                                            // straight from Cloudflare
		{"auto", "127.0.0.1:5000", cf, "203.0.113.10", "CN"},                                                            // via local nginx
		{"auto", "172.18.0.1:5000", map[string]string{"X-Forwarded-For": "198.51.100.7, 10.0.0.2"}, "198.51.100.7", ""}, // docker network proxy
		{"auto", "198.51.100.99:5000", cf, "198.51.100.99", ""},                                                         // direct internet client: headers ignored
		{"0", "127.0.0.1:5000", cf, "127.0.0.1", ""},
		{"1", "198.51.100.99:5000", cf, "203.0.113.10", "CN"},
		{"auto", "127.0.0.1:5000", map[string]string{"CF-Connecting-IP": "not-an-ip", "X-Real-IP": "192.0.2.4"}, "192.0.2.4", ""},
	}
	defer func() { proxyMode = "auto" }()
	for _, c := range cases {
		proxyMode = c.mode
		r := req(c.peer, c.h)
		if got := clientIP(r); got != c.want {
			t.Errorf("%s %s: ip %s want %s", c.mode, c.peer, got, c.want)
		}
		if got := clientCountry(r); got != c.cc {
			t.Errorf("%s %s: country %q want %q", c.mode, c.peer, got, c.cc)
		}
	}
}

func TestParseRate(t *testing.T) {
	h := http.Header{}
	h.Set("x-ratelimit-limit-requests", "5000")
	h.Set("x-ratelimit-remaining-requests", "4999")
	h.Set("x-ratelimit-reset-requests", "12ms")
	h.Set("x-ratelimit-limit-tokens", "2000000")
	h.Set("x-ratelimit-remaining-tokens", "1998000")
	h.Set("x-ratelimit-reset-tokens", "6m0s")
	r, ok := parseRate(h)
	if !ok || r.ReqRemaining != 4999 || r.TokLimit != 2000000 || r.TokReset < time.Now().Add(5*time.Minute).Unix() {
		t.Errorf("openai headers: %+v", r)
	}
	a := http.Header{}
	reset := time.Now().Add(time.Minute).UTC().Format(time.RFC3339)
	a.Set("anthropic-ratelimit-requests-limit", "50")
	a.Set("anthropic-ratelimit-requests-remaining", "49")
	a.Set("anthropic-ratelimit-requests-reset", reset)
	a.Set("anthropic-ratelimit-input-tokens-limit", "40000")
	a.Set("anthropic-ratelimit-input-tokens-remaining", "39000")
	r, ok = parseRate(a)
	if !ok || r.ReqLimit != 50 || r.TokRemaining != 39000 || r.ReqReset == 0 {
		t.Errorf("anthropic headers: %+v", r)
	}
	if _, ok := parseRate(http.Header{}); ok {
		t.Error("empty headers parsed as a rate")
	}
}

func TestQuotaAndAccessLog(t *testing.T) {
	a := testApp(t, "")
	u, _ := a.store.CreateUser("u", hashPassword("password1"), false, 100)
	a.store.SaveProvider(&Provider{Name: "P", Kind: "openai", APIKey: "k", Models: []string{"m"}, Enabled: true})
	a.store.AddUsage(u.ID, 80, 30, false)
	h := a.routes()
	do := func(method, path string, body any, cookie string) *httptest.ResponseRecorder {
		var b bytes.Buffer
		json.NewEncoder(&b).Encode(body)
		r := httptest.NewRequest(method, path, &b)
		r.RemoteAddr = "127.0.0.1:1234"
		r.Header.Set("CF-Connecting-IP", "203.0.113.9")
		if cookie != "" {
			r.AddCookie(&http.Cookie{Name: sessionCookie, Value: cookie})
		}
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, r)
		return rec
	}
	do("POST", "/api/login", map[string]string{"username": "u", "password": "wrong"}, "")
	rec := do("POST", "/api/login", map[string]string{"username": "u", "password": "password1"}, "")
	var tok string
	for _, c := range rec.Result().Cookies() {
		if c.Name == sessionCookie {
			tok = c.Value
		}
	}
	rec = do("POST", "/api/send", map[string]any{"content": "hi", "model": "1/m"}, tok)
	var body struct {
		Code    string `json:"code"`
		ResetAt int64  `json:"reset_at"`
	}
	json.Unmarshal(rec.Body.Bytes(), &body)
	if rec.Code != 429 || body.Code != "quota_exceeded" || body.ResetAt <= time.Now().Unix() {
		t.Errorf("quota: %d %s", rec.Code, rec.Body)
	}
	log, _ := a.store.AccessLog(0, 0, 10)
	if len(log) != 2 || log[0].Event != "login" || log[1].Event != "login_failed" || log[0].IP != "203.0.113.9" || log[1].OK {
		t.Errorf("access log: %+v", log)
	}
	if got, _ := a.store.UserByID(u.ID); got.LastIP != "203.0.113.9" || got.LastSeenAt == 0 {
		t.Errorf("last seen not recorded: %+v", got)
	}
	// default limit for new users
	a.store.SetSetting(settingDefaultLimit, "5000")
	if a.settings().DefaultLimit != 5000 {
		t.Error("default limit setting")
	}
}
