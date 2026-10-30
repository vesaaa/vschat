package main

import (
	"context"
	"net/http"
	"strings"
	"time"
)

func (a *App) handleListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := a.store.ListUsers()
	if err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, users)
}

type userReq struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	IsAdmin    *bool  `json:"is_admin"`
	Disabled   *bool  `json:"disabled"`
	TokenLimit *int64 `json:"token_limit"`
}

func (a *App) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var req userReq
	if !readJSON(w, r, &req) {
		return
	}
	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" || len(req.Password) < 6 {
		writeErr(w, 400, "username required, password at least 6 characters")
		return
	}
	var limit int64
	if req.TokenLimit != nil {
		limit = *req.TokenLimit
	}
	u, err := a.store.CreateUser(req.Username, hashPassword(req.Password), req.IsAdmin != nil && *req.IsAdmin, limit)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			writeErr(w, 409, "username already exists")
			return
		}
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, u)
}

func (a *App) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	u, err := a.store.UserByID(pathID(r))
	if err != nil {
		writeErr(w, 404, "not found")
		return
	}
	var req userReq
	if !readJSON(w, r, &req) {
		return
	}
	self := u.ID == currentUser(r).ID
	if req.IsAdmin != nil {
		if self && !*req.IsAdmin {
			writeErr(w, 400, "cannot remove your own admin role")
			return
		}
		u.IsAdmin = *req.IsAdmin
	}
	if req.Disabled != nil {
		if self && *req.Disabled {
			writeErr(w, 400, "cannot disable yourself")
			return
		}
		u.Disabled = *req.Disabled
	}
	if req.TokenLimit != nil {
		u.TokenLimit = *req.TokenLimit
	}
	if err := a.store.UpdateUser(u); err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	if req.Password != "" {
		if len(req.Password) < 6 {
			writeErr(w, 400, "password must be at least 6 characters")
			return
		}
		a.store.SetPassword(u.ID, hashPassword(req.Password))
		a.store.DeleteUserSessions(u.ID, sessionHash(r))
	}
	if u.Disabled {
		a.store.DeleteUserSessions(u.ID, "")
	}
	writeJSON(w, u)
}

func (a *App) handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	if id == currentUser(r).ID {
		writeErr(w, 400, "cannot delete yourself")
		return
	}
	a.store.DeleteUser(id)
	w.WriteHeader(http.StatusNoContent)
}

// Keys never leave the server; the admin UI only sees whether one is set.
type providerView struct {
	*Provider
	HasKey  bool   `json:"has_key"`
	KeyHint string `json:"key_hint"`
}

func viewProvider(p *Provider) providerView {
	v := providerView{Provider: p, HasKey: p.APIKey != ""}
	if n := len(p.APIKey); n > 8 {
		v.KeyHint = p.APIKey[:3] + "…" + p.APIKey[n-4:]
	}
	cp := *p
	cp.APIKey = ""
	v.Provider = &cp
	return v
}

func (a *App) handleListProviders(w http.ResponseWriter, r *http.Request) {
	ps, err := a.store.ListProviders()
	if err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	out := []providerView{}
	for _, p := range ps {
		out = append(out, viewProvider(p))
	}
	writeJSON(w, out)
}

type providerReq struct {
	Name    string   `json:"name"`
	Kind    string   `json:"kind"`
	BaseURL string   `json:"base_url"`
	APIKey  *string  `json:"api_key"` // nil keeps the stored key
	Models  []string `json:"models"`
	Enabled bool     `json:"enabled"`
	Sort    int64    `json:"sort"`
}

func (a *App) saveProvider(w http.ResponseWriter, r *http.Request, p *Provider) {
	var req providerReq
	if !readJSON(w, r, &req) {
		return
	}
	if req.Kind != "openai" && req.Kind != "anthropic" {
		writeErr(w, 400, "kind must be openai or anthropic")
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		writeErr(w, 400, "name required")
		return
	}
	p.Name, p.Kind, p.BaseURL = strings.TrimSpace(req.Name), req.Kind, strings.TrimSpace(req.BaseURL)
	p.Models = splitModels(strings.Join(req.Models, "\n"))
	p.Enabled, p.Sort = req.Enabled, req.Sort
	if req.APIKey != nil {
		p.APIKey = strings.TrimSpace(*req.APIKey)
	}
	if err := a.store.SaveProvider(p); err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, viewProvider(p))
}

func (a *App) handleCreateProvider(w http.ResponseWriter, r *http.Request) {
	a.saveProvider(w, r, &Provider{})
}

func (a *App) handleUpdateProvider(w http.ResponseWriter, r *http.Request) {
	p, err := a.store.Provider(pathID(r))
	if err != nil {
		writeErr(w, 404, "not found")
		return
	}
	a.saveProvider(w, r, p)
}

func (a *App) handleDeleteProvider(w http.ResponseWriter, r *http.Request) {
	a.store.DeleteProvider(pathID(r))
	w.WriteHeader(http.StatusNoContent)
}

// handleRemoteModels asks the provider which models the key can use. The
// request may carry an unsaved key/base URL so admins can test before saving.
func (a *App) handleRemoteModels(w http.ResponseWriter, r *http.Request) {
	var req providerReq
	if !readJSON(w, r, &req) {
		return
	}
	p := &Provider{Kind: req.Kind, BaseURL: strings.TrimSpace(req.BaseURL)}
	if req.APIKey != nil && *req.APIKey != "" {
		p.APIKey = strings.TrimSpace(*req.APIKey)
	} else if id := pathID(r); id > 0 {
		if saved, err := a.store.Provider(id); err == nil {
			p.APIKey = saved.APIKey
		}
	}
	if p.APIKey == "" {
		writeErr(w, 400, "API key required")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
	defer cancel()
	models, err := p.ListRemoteModels(ctx)
	if err != nil {
		writeErr(w, 502, err.Error())
		return
	}
	writeJSON(w, models)
}

func (a *App) handleUsage(w http.ResponseWriter, r *http.Request) {
	m := r.URL.Query().Get("month")
	if m == "" {
		m = month()
	}
	rows, err := a.store.Usage(m)
	if err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, map[string]any{"month": m, "rows": rows})
}
