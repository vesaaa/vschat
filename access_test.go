package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func testApp(t *testing.T, secret string) *App {
	t.Helper()
	st, err := openStore(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	st.box = newSealer(secret)
	t.Cleanup(func() { st.db.Close() })
	return &App{store: st, limiter: &loginLimiter{fail: map[string][]time.Time{}}}
}

func mustUser(t *testing.T, a *App, name string, admin bool) *User {
	u, err := a.store.CreateUser(name, hashPassword("password1"), admin, 0)
	if err != nil {
		t.Fatal(err)
	}
	return u
}

func ids(opts []ModelOption) string {
	var s []string
	for _, o := range opts {
		s = append(s, o.Model)
	}
	return strings.Join(s, ",")
}

func TestModelVisibility(t *testing.T) {
	a := testApp(t, "")
	alice, bob := mustUser(t, a, "alice", false), mustUser(t, a, "bob", false)
	pub := &Provider{Name: "Pub", Kind: "openai", APIKey: "k", Models: []string{"pub-1"}, Enabled: true}
	vip := &Provider{Name: "VIP", Kind: "openai", APIKey: "k", Models: []string{"vip-1"}, Enabled: true, Visibility: "users", UserIDs: []int64{alice.ID}}
	off := &Provider{Name: "Off", Kind: "openai", APIKey: "k", Models: []string{"off-1"}, Enabled: false}
	own := &Provider{Name: "Mine", Kind: "openai", APIKey: "k", Models: []string{"own-1"}, Enabled: true, OwnerID: bob.ID}
	for _, p := range []*Provider{pub, vip, off, own} {
		if err := a.store.SaveProvider(p); err != nil {
			t.Fatal(err)
		}
	}

	if got := ids(a.modelOptions(alice)); got != "pub-1,vip-1" {
		t.Errorf("alice sees %q", got)
	}
	// Bob's own key is hidden until the admin allows personal keys.
	if got := ids(a.modelOptions(bob)); got != "pub-1" {
		t.Errorf("bob sees %q", got)
	}
	a.store.SetSetting(settingKeyMode, keyModeCustom)
	if got := ids(a.modelOptions(bob)); got != "pub-1,own-1" {
		t.Errorf("bob with own keys sees %q", got)
	}
	if got := ids(a.modelOptions(alice)); got != "pub-1,vip-1" {
		t.Errorf("alice must not see bob's provider, got %q", got)
	}

	for _, c := range []struct {
		u   *User
		ref string
		ok  bool
	}{
		{alice, "2/vip-1", true}, {bob, "2/vip-1", false}, {alice, "3/off-1", false},
		{alice, "4/own-1", false}, {bob, "4/own-1", true}, {bob, "1/not-listed", false},
	} {
		p, _, err := a.resolveModel(c.u, c.ref)
		if (err == nil) != c.ok {
			t.Errorf("%s %s: err=%v want ok=%v", c.u.Username, c.ref, err, c.ok)
		}
		if err == nil && p.OwnerID != 0 && !p.guard {
			t.Errorf("own provider must be network-guarded by default")
		}
	}
	a.store.SetSetting(settingPrivateNetwork, "1")
	if p, _, _ := a.resolveModel(bob, "4/own-1"); p.guard {
		t.Errorf("guard should be off when private network is allowed")
	}
}

func TestOwnProviderAPI(t *testing.T) {
	a := testApp(t, "")
	alice, bob := mustUser(t, a, "alice", false), mustUser(t, a, "bob", false)
	h := a.routes()
	call := func(u *User, method, path string, body any) *httptest.ResponseRecorder {
		var b bytes.Buffer
		json.NewEncoder(&b).Encode(body)
		tok := randomToken(16)
		a.store.CreateSession(hashToken(tok), u.ID, time.Hour)
		req := httptest.NewRequest(method, path, &b)
		req.AddCookie(&http.Cookie{Name: sessionCookie, Value: tok})
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec
	}
	newProv := map[string]any{"name": "Mine", "kind": "openai", "api_key": "sk-alice-secret-123", "models": []string{"m1"}, "enabled": true,
		"visibility": "users", "user_ids": []int64{bob.ID}}

	if rec := call(alice, "POST", "/api/my/providers", newProv); rec.Code != 403 {
		t.Fatalf("personal keys off: got %d", rec.Code)
	}
	if rec := call(alice, "POST", "/api/my/providers", newProv); rec.Code != 403 {
		t.Fatalf("still off: got %d", rec.Code)
	}
	a.store.SetSetting(settingKeyMode, keyModeCustom)
	rec := call(alice, "POST", "/api/my/providers", newProv)
	if rec.Code != 200 {
		t.Fatalf("create: %d %s", rec.Code, rec.Body)
	}
	if strings.Contains(rec.Body.String(), "sk-alice-secret-123") {
		t.Fatal("key leaked in response")
	}
	var created providerView
	json.Unmarshal(rec.Body.Bytes(), &created)
	if created.OwnerID != alice.ID || created.Visibility != "all" || len(created.UserIDs) != 0 {
		t.Fatalf("own provider must not take sharing settings: %+v", created.Provider)
	}
	path := "/api/my/providers/" + jsonID(created.ID)
	if rec := call(bob, "PUT", path, newProv); rec.Code != 404 {
		t.Errorf("bob editing alice's provider: %d", rec.Code)
	}
	if rec := call(bob, "DELETE", path, nil); rec.Code != 404 {
		t.Errorf("bob deleting alice's provider: %d", rec.Code)
	}
	if rec := call(alice, "PUT", "/api/admin/providers/"+jsonID(created.ID), newProv); rec.Code != 403 {
		t.Errorf("non-admin on admin API: %d", rec.Code)
	}
	if rec := call(bob, "GET", "/api/my/providers", nil); strings.Contains(rec.Body.String(), "Mine") {
		t.Errorf("bob lists alice's provider: %s", rec.Body)
	}
	bad := map[string]any{"name": "x", "kind": "openai", "base_url": "file:///etc/passwd"}
	if rec := call(alice, "POST", "/api/my/providers", bad); rec.Code != 400 {
		t.Errorf("non-http base URL accepted: %d", rec.Code)
	}
	// Admin's shared list never includes personal providers.
	admin := mustUser(t, a, "root", true)
	if rec := call(admin, "GET", "/api/admin/providers", nil); strings.Contains(rec.Body.String(), "Mine") {
		t.Errorf("admin list includes personal provider")
	}
	// Deleting a user removes their providers.
	a.store.DeleteUser(alice.ID)
	if _, err := a.store.Provider(created.ID); !errors.Is(err, errNotFound) {
		t.Errorf("provider survived owner deletion: %v", err)
	}
}

func jsonID(id int64) string { b, _ := json.Marshal(id); return string(b) }

func TestUsageLimitIgnoresOwnKeys(t *testing.T) {
	a := testApp(t, "")
	u := mustUser(t, a, "u", false)
	a.store.AddUsage(u.ID, 100, 50, false)
	a.store.AddUsage(u.ID, 9000, 9000, true)
	if n, _ := a.store.MonthTokens(u.ID); n != 150 {
		t.Errorf("limited tokens = %d, want 150", n)
	}
	rows, _ := a.store.Usage(month())
	if rows[0].OwnTokens != 18000 || rows[0].Requests != 2 {
		t.Errorf("usage row %+v", rows[0])
	}
}

func TestKeyEncryption(t *testing.T) {
	a := testApp(t, "s3cret")
	p := &Provider{Name: "P", Kind: "openai", APIKey: "sk-plain-key", Enabled: true}
	a.store.SaveProvider(p)
	var raw string
	a.store.db.QueryRow(`SELECT api_key FROM providers WHERE id=?`, p.ID).Scan(&raw)
	if !strings.HasPrefix(raw, encPrefix) || strings.Contains(raw, "sk-plain") {
		t.Fatalf("stored key not encrypted: %q", raw)
	}
	got, _ := a.store.Provider(p.ID)
	if got.APIKey != "sk-plain-key" || got.KeyErr != nil {
		t.Fatalf("decrypt: %q %v", got.APIKey, got.KeyErr)
	}
	// Wrong secret: key unusable, and saving without a new key keeps the ciphertext.
	a.store.box = newSealer("other")
	got, _ = a.store.Provider(p.ID)
	if got.KeyErr == nil {
		t.Fatal("expected decrypt error with wrong secret")
	}
	got.Name = "renamed"
	a.store.SaveProvider(got)
	var after string
	a.store.db.QueryRow(`SELECT api_key FROM providers WHERE id=?`, p.ID).Scan(&after)
	if after != raw {
		t.Fatal("ciphertext lost on save")
	}
	// Existing plain-text keys are sealed when a secret is configured.
	a.store.box = nil
	plain := &Provider{Name: "Old", Kind: "openai", APIKey: "sk-old", Enabled: true}
	a.store.SaveProvider(plain)
	a.store.box = newSealer("s3cret")
	a.store.SealKeys()
	a.store.db.QueryRow(`SELECT api_key FROM providers WHERE id=?`, plain.ID).Scan(&raw)
	if !strings.HasPrefix(raw, encPrefix) {
		t.Fatal("SealKeys left plain text")
	}
}

func TestMigrateV01(t *testing.T) {
	path := filepath.Join(t.TempDir(), "old.db")
	db, _ := sql.Open("sqlite", path)
	_, err := db.Exec(`CREATE TABLE providers (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, kind TEXT NOT NULL,
		base_url TEXT NOT NULL DEFAULT '', api_key TEXT NOT NULL DEFAULT '', models TEXT NOT NULL DEFAULT '',
		enabled INTEGER NOT NULL DEFAULT 1, sort INTEGER NOT NULL DEFAULT 0);
		CREATE TABLE usage (user_id INTEGER NOT NULL, month TEXT NOT NULL, tokens_in INTEGER NOT NULL DEFAULT 0,
		tokens_out INTEGER NOT NULL DEFAULT 0, requests INTEGER NOT NULL DEFAULT 0, PRIMARY KEY (user_id, month));
		INSERT INTO providers(name, kind, api_key, models) VALUES('Old', 'openai', 'sk-x', 'gpt-x');`)
	db.Close()
	if err != nil {
		t.Fatal(err)
	}
	st, err := openStore(path)
	if err != nil {
		t.Fatal(err)
	}
	defer st.db.Close()
	ps, err := st.SharedProviders()
	if err != nil || len(ps) != 1 || ps[0].Visibility != "all" || ps[0].OwnerID != 0 || ps[0].APIKey != "sk-x" {
		t.Fatalf("migrated providers: %v %+v", err, ps)
	}
}

func TestNetworkGuard(t *testing.T) {
	for addr, want := range map[string]bool{
		"8.8.8.8": true, "2606:4700::1111": true,
		"127.0.0.1": false, "10.1.2.3": false, "192.168.1.1": false, "172.16.0.1": false, "169.254.169.254": false,
		"100.64.0.1": false, "0.0.0.0": false, "::1": false, "fd00::1": false, "fe80::1": false, "::ffff:127.0.0.1": false,
	} {
		if got := publicAddr(netip.MustParseAddr(addr)); got != want {
			t.Errorf("%s: public=%v want %v", addr, got, want)
		}
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer srv.Close()
	p := &Provider{Kind: "openai", BaseURL: srv.URL, APIKey: "k", guard: true}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := p.ListRemoteModels(ctx); err == nil || !strings.Contains(err.Error(), "not allowed") {
		t.Errorf("guarded request to loopback: %v", err)
	}
	p.guard = false
	if _, err := p.ListRemoteModels(ctx); err != nil && strings.Contains(err.Error(), "not allowed") {
		t.Errorf("unguarded request blocked: %v", err)
	}
}

func TestOfficialKeyMode(t *testing.T) {
	a := testApp(t, "")
	u := mustUser(t, a, "u", false)
	official := &Provider{Name: "Mine", Kind: "anthropic", BaseURL: "https://api.anthropic.com/", APIKey: "k", Models: []string{"claude-x"}, Enabled: true, OwnerID: u.ID}
	custom := &Provider{Name: "Relay", Kind: "openai", BaseURL: "https://relay.example.com/v1", APIKey: "k", Models: []string{"relay-x"}, Enabled: true, OwnerID: u.ID}
	// Wrong kind for an official URL does not count as official.
	sneaky := &Provider{Name: "Sneaky", Kind: "openai", BaseURL: "https://api.anthropic.com", APIKey: "k", Models: []string{"s-x"}, Enabled: true, OwnerID: u.ID}
	for _, p := range []*Provider{official, custom, sneaky} {
		a.store.SaveProvider(p)
	}
	for mode, want := range map[string]string{
		keyModeOff: "", keyModeOfficial: "claude-x", keyModeCustom: "claude-x,relay-x,s-x",
	} {
		a.store.SetSetting(settingKeyMode, mode)
		if got := ids(a.modelOptions(u)); got != want {
			t.Errorf("%s: got %q want %q", mode, got, want)
		}
	}
	// v0.2.0 switch is read as custom.
	a.store.SetSetting(settingKeyMode, "")
	a.store.SetSetting(settingUserKeysV020, "1")
	if a.keyMode() != keyModeCustom {
		t.Errorf("legacy setting read as %q", a.keyMode())
	}
	for _, c := range []struct {
		kind, url string
		want      bool
	}{
		{"openai", "", true}, {"openai", "https://api.openai.com/v1/", true}, {"openai", "https://api.deepseek.com", true},
		{"anthropic", "", true}, {"openai", "https://api.openai.com.evil.com/v1", false}, {"openai", "http://api.openai.com/v1", false},
	} {
		if isOfficial(c.kind, c.url) != c.want {
			t.Errorf("isOfficial(%s, %q) want %v", c.kind, c.url, c.want)
		}
	}
}

func TestCatalogPresets(t *testing.T) {
	a := testApp(t, "")
	// v0.2-style rows: one on the old seed list, one hand-picked, one custom relay.
	for _, p := range []*Provider{
		{Name: "DeepSeek", Kind: "openai", BaseURL: "https://api.deepseek.com", APIKey: "k", Models: []string{"deepseek-chat", "deepseek-reasoner"}, Enabled: true},
		{Name: "OpenAI", Kind: "openai", BaseURL: "https://api.openai.com/v1/", APIKey: "k", Models: []string{"gpt-6-luna"}, Enabled: true},
		{Name: "Relay", Kind: "openai", BaseURL: "https://relay.example.com/v1", APIKey: "k", Models: []string{"x"}, Enabled: true},
	} {
		a.store.SaveProvider(p)
	}
	if err := a.store.LinkPresets(); err != nil {
		t.Fatal(err)
	}
	ps, _ := a.store.SharedProviders()
	ds, oa, relay := ps[0], ps[1], ps[2]
	if ds.Preset != "deepseek" || !ds.AutoModels || strings.Join(ds.Models, ",") != strings.Join(presetByID("deepseek").Models, ",") {
		t.Errorf("legacy seed should follow catalog: %+v", ds)
	}
	if oa.Preset != "openai" || oa.AutoModels || strings.Join(oa.Models, ",") != "gpt-6-luna" {
		t.Errorf("hand-picked list must be kept: %+v", oa)
	}
	if relay.Preset != "" || relay.AutoModels {
		t.Errorf("custom provider got a preset: %+v", relay)
	}

	// Saving through the API with a preset: kind and address come from the catalog.
	admin := mustUser(t, a, "root", true)
	tok := randomToken(16)
	a.store.CreateSession(hashToken(tok), admin.ID, time.Hour)
	post := func(body any) *httptest.ResponseRecorder {
		var b bytes.Buffer
		json.NewEncoder(&b).Encode(body)
		req := httptest.NewRequest("POST", "/api/admin/providers", &b)
		req.AddCookie(&http.Cookie{Name: sessionCookie, Value: tok})
		rec := httptest.NewRecorder()
		a.routes().ServeHTTP(rec, req)
		return rec
	}
	rec := post(map[string]any{"preset": "anthropic", "kind": "openai", "name": "Claude", "api_key": "k", "auto_models": true, "enabled": true})
	var v providerView
	json.Unmarshal(rec.Body.Bytes(), &v)
	if rec.Code != 200 || v.Kind != "anthropic" || v.BaseURL != "https://api.anthropic.com" || !v.AutoModels || len(v.Models) == 0 {
		t.Fatalf("preset save: %d %s", rec.Code, rec.Body)
	}
	var raw string
	a.store.db.QueryRow(`SELECT models FROM providers WHERE id=?`, v.ID).Scan(&raw)
	if raw != "" {
		t.Errorf("auto models stored a fixed list: %q", raw)
	}
	if rec := post(map[string]any{"kind": "openai", "name": "Empty", "api_key": "k", "enabled": true}); rec.Code != 400 {
		t.Errorf("custom provider without models accepted: %d", rec.Code)
	}
	for _, pr := range catalog {
		if !isOfficial(pr.Kind, pr.BaseURL) || len(pr.Models) == 0 || pr.KeyEnv == "" {
			t.Errorf("catalog entry incomplete: %s", pr.ID)
		}
	}
}
