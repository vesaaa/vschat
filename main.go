// VsChat is a small self-hosted multi-user chat for OpenAI, Claude and DeepSeek.
package main

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

var version = "dev"

//go:embed web
var webFS embed.FS

type App struct {
	store   *Store
	limiter *loginLimiter
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func readJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid request body")
		return false
	}
	return true
}

// bootstrap creates the first admin and the default providers on an empty database.
func (a *App) bootstrap() error {
	if n, err := a.store.CountUsers(); err != nil {
		return err
	} else if n == 0 {
		name := env("VSCHAT_ADMIN_USER", "admin")
		pw := os.Getenv("VSCHAT_ADMIN_PASSWORD")
		generated := pw == ""
		if generated {
			pw = randomToken(12)
		}
		if _, err := a.store.CreateUser(name, hashPassword(pw), true, 0); err != nil {
			return err
		}
		if generated {
			log.Printf("created admin user %q with password: %s  (change it after logging in)", name, pw)
		} else {
			log.Printf("created admin user %q", name)
		}
	}
	if n, err := a.store.CountProviders(); err != nil {
		return err
	} else if n == 0 {
		defaults := []*Provider{
			{Name: "OpenAI", Kind: "openai", BaseURL: defaultOpenAIBase, APIKey: os.Getenv("OPENAI_API_KEY"),
				Models: []string{"gpt-5", "gpt-5-mini"}, Enabled: true, Sort: 1},
			{Name: "Claude", Kind: "anthropic", BaseURL: defaultAnthropicBase, APIKey: os.Getenv("ANTHROPIC_API_KEY"),
				Models: []string{"claude-opus-5-5", "claude-sonnet-5-5", "claude-haiku-5-5"}, Enabled: true, Sort: 2},
			{Name: "DeepSeek", Kind: "openai", BaseURL: "https://api.deepseek.com", APIKey: os.Getenv("DEEPSEEK_API_KEY"),
				Models: []string{"deepseek-chat", "deepseek-reasoner"}, Enabled: true, Sort: 3},
		}
		for _, p := range defaults {
			if err := a.store.SaveProvider(p); err != nil {
				return err
			}
		}
	}
	return nil
}

func (a *App) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/login", a.handleLogin)
	mux.HandleFunc("POST /api/logout", a.handleLogout)
	mux.HandleFunc("GET /api/me", a.requireUser(a.handleMe))
	mux.HandleFunc("POST /api/me/password", a.requireUser(a.handleChangePassword))
	mux.HandleFunc("GET /api/models", a.requireUser(a.handleModels))
	mux.HandleFunc("GET /api/chats", a.requireUser(a.handleListChats))
	mux.HandleFunc("GET /api/chats/{id}", a.requireUser(a.handleGetChat))
	mux.HandleFunc("PATCH /api/chats/{id}", a.requireUser(a.handleUpdateChat))
	mux.HandleFunc("DELETE /api/chats/{id}", a.requireUser(a.handleDeleteChat))
	mux.HandleFunc("POST /api/send", a.requireUser(a.handleSend))

	mux.HandleFunc("GET /api/admin/users", a.requireAdmin(a.handleListUsers))
	mux.HandleFunc("POST /api/admin/users", a.requireAdmin(a.handleCreateUser))
	mux.HandleFunc("PATCH /api/admin/users/{id}", a.requireAdmin(a.handleUpdateUser))
	mux.HandleFunc("DELETE /api/admin/users/{id}", a.requireAdmin(a.handleDeleteUser))
	mux.HandleFunc("GET /api/admin/providers", a.requireAdmin(a.handleListProviders))
	mux.HandleFunc("POST /api/admin/providers", a.requireAdmin(a.handleCreateProvider))
	mux.HandleFunc("PUT /api/admin/providers/{id}", a.requireAdmin(a.handleUpdateProvider))
	mux.HandleFunc("DELETE /api/admin/providers/{id}", a.requireAdmin(a.handleDeleteProvider))
	mux.HandleFunc("POST /api/admin/providers/{id}/models", a.requireAdmin(a.handleRemoteModels))
	mux.HandleFunc("GET /api/admin/usage", a.requireAdmin(a.handleUsage))
	mux.HandleFunc("GET /api/admin/settings", a.requireAdmin(a.handleGetSettings))
	mux.HandleFunc("PUT /api/admin/settings", a.requireAdmin(a.handlePutSettings))

	mux.HandleFunc("GET /api/my/providers", a.requireUserKeys(a.handleListOwn))
	mux.HandleFunc("POST /api/my/providers", a.requireUserKeys(a.handleCreateOwn))
	mux.HandleFunc("PUT /api/my/providers/{id}", a.requireUserKeys(a.handleUpdateOwn))
	mux.HandleFunc("DELETE /api/my/providers/{id}", a.requireUserKeys(a.handleDeleteOwn))
	mux.HandleFunc("POST /api/my/providers/{id}/models", a.requireUserKeys(a.handleOwnRemoteModels))

	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]string{"status": "ok", "version": version})
	})

	static, _ := fs.Sub(webFS, "web")
	files := http.FileServer(http.FS(static))
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" || r.URL.Path == "/index.html" {
			w.Header().Set("Cache-Control", "no-cache")
		}
		files.ServeHTTP(w, r)
	})
	return securityHeaders(mux)
}

func securityHeaders(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "same-origin")
		w.Header().Set("X-Frame-Options", "DENY")
		h.ServeHTTP(w, r)
	})
}

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "-v" || os.Args[1] == "--version") {
		println("vschat", version)
		return
	}
	trustProxy = os.Getenv("VSCHAT_TRUST_PROXY") == "1"
	dataDir := env("VSCHAT_DATA", "data")
	if err := os.MkdirAll(dataDir, 0o700); err != nil {
		log.Fatal(err)
	}
	store, err := openStore(filepath.Join(dataDir, "vschat.db"))
	if err != nil {
		log.Fatal(err)
	}
	store.box = newSealer(os.Getenv("VSCHAT_SECRET"))
	if err := store.SealKeys(); err != nil {
		log.Fatal(err)
	}
	app := &App{store: store, limiter: &loginLimiter{fail: map[string][]time.Time{}}}
	if err := app.bootstrap(); err != nil {
		log.Fatal(err)
	}
	go func() {
		for range time.Tick(time.Hour) {
			store.PurgeSessions()
		}
	}()

	addr := env("VSCHAT_ADDR", ":8080")
	srv := &http.Server{Addr: addr, Handler: app.routes(), ReadHeaderTimeout: 10 * time.Second}
	go func() {
		log.Printf("VsChat %s listening on %s", version, addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
	store.db.Close()
}
