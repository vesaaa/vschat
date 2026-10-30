package main

import (
	"context"
	"net/http"
	"slices"
	"strings"
	"time"
)

const (
	settingUserKeys       = "allow_user_keys"       // users may add their own providers
	settingPrivateNetwork = "allow_private_network" // their providers may reach local/private addresses
)

func (a *App) userKeysAllowed() bool { return a.store.Setting(settingUserKeys) == "1" }

// canUse decides whether u may chat through p. It also marks users' own
// providers for the private-network guard.
func (a *App) canUse(u *User, p *Provider) bool {
	if !p.Enabled {
		return false
	}
	if p.OwnerID != 0 {
		p.guard = a.store.Setting(settingPrivateNetwork) != "1"
		return p.OwnerID == u.ID && a.userKeysAllowed()
	}
	return p.Visibility != "users" || slices.Contains(p.UserIDs, u.ID)
}

// Keys never leave the server; the UI only sees whether one is set.
type providerView struct {
	*Provider
	HasKey  bool   `json:"has_key"`
	KeyHint string `json:"key_hint"`
	KeyErr  string `json:"key_error,omitempty"`
}

func viewProvider(p *Provider) providerView {
	v := providerView{HasKey: p.APIKey != "" || p.KeyErr != nil}
	if n := len(p.APIKey); n > 8 {
		v.KeyHint = p.APIKey[:3] + "…" + p.APIKey[n-4:]
	}
	if p.KeyErr != nil {
		v.KeyErr = p.KeyErr.Error()
	}
	cp := *p
	cp.APIKey = ""
	v.Provider = &cp
	return v
}

func writeProviders(w http.ResponseWriter, ps []*Provider, err error) {
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
	Name       string   `json:"name"`
	Kind       string   `json:"kind"`
	BaseURL    string   `json:"base_url"`
	APIKey     *string  `json:"api_key"` // nil keeps the stored key
	Models     []string `json:"models"`
	Enabled    bool     `json:"enabled"`
	Sort       int64    `json:"sort"`
	Visibility string   `json:"visibility"`
	UserIDs    []int64  `json:"user_ids"`
}

// saveProvider applies a create/update request. Shared providers (OwnerID 0)
// take visibility settings; a user's own provider is always just theirs.
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
	req.BaseURL = strings.TrimSpace(req.BaseURL)
	if !validBaseURL(req.BaseURL) {
		writeErr(w, 400, "base URL must be an http(s) URL")
		return
	}
	p.Name, p.Kind, p.BaseURL = strings.TrimSpace(req.Name), req.Kind, req.BaseURL
	p.Models = splitModels(strings.Join(req.Models, "\n"))
	p.Enabled, p.Sort = req.Enabled, req.Sort
	if req.APIKey != nil {
		p.APIKey = strings.TrimSpace(*req.APIKey)
	}
	if p.OwnerID == 0 {
		p.Visibility, p.UserIDs = "all", nil
		if req.Visibility == "users" {
			p.Visibility, p.UserIDs = "users", req.UserIDs
		}
	} else {
		p.Visibility, p.UserIDs = "all", nil
	}
	if err := a.store.SaveProvider(p); err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	saved, err := a.store.Provider(p.ID)
	if err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, viewProvider(saved))
}

// remoteModels asks the provider which models the key can use. The request may
// carry an unsaved key/base URL so the form can be tested before saving;
// saved is the stored provider to borrow the key from, if any.
func (a *App) remoteModels(w http.ResponseWriter, r *http.Request, saved *Provider, guard bool) {
	var req providerReq
	if !readJSON(w, r, &req) {
		return
	}
	p := &Provider{Kind: req.Kind, BaseURL: strings.TrimSpace(req.BaseURL), guard: guard}
	if !validBaseURL(p.BaseURL) {
		writeErr(w, 400, "base URL must be an http(s) URL")
		return
	}
	if req.APIKey != nil && *req.APIKey != "" {
		p.APIKey = strings.TrimSpace(*req.APIKey)
	} else if saved != nil {
		p.APIKey, p.KeyErr = saved.APIKey, saved.KeyErr
	}
	if p.KeyErr != nil {
		writeErr(w, 400, p.KeyErr.Error())
		return
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

// --- shared providers (admin) ---

func (a *App) sharedProvider(w http.ResponseWriter, r *http.Request) *Provider {
	p, err := a.store.Provider(pathID(r))
	if err != nil || p.OwnerID != 0 {
		writeErr(w, 404, "not found")
		return nil
	}
	return p
}

func (a *App) handleListProviders(w http.ResponseWriter, r *http.Request) {
	ps, err := a.store.SharedProviders()
	writeProviders(w, ps, err)
}

func (a *App) handleCreateProvider(w http.ResponseWriter, r *http.Request) {
	a.saveProvider(w, r, &Provider{})
}

func (a *App) handleUpdateProvider(w http.ResponseWriter, r *http.Request) {
	if p := a.sharedProvider(w, r); p != nil {
		a.saveProvider(w, r, p)
	}
}

func (a *App) handleDeleteProvider(w http.ResponseWriter, r *http.Request) {
	if p := a.sharedProvider(w, r); p != nil {
		a.store.DeleteProvider(p.ID)
		w.WriteHeader(http.StatusNoContent)
	}
}

func (a *App) handleRemoteModels(w http.ResponseWriter, r *http.Request) {
	var saved *Provider
	if pathID(r) > 0 {
		if saved = a.sharedProvider(w, r); saved == nil {
			return
		}
	}
	a.remoteModels(w, r, saved, false)
}

// --- the user's own providers ---

func (a *App) requireUserKeys(next http.HandlerFunc) http.HandlerFunc {
	return a.requireUser(func(w http.ResponseWriter, r *http.Request) {
		if !a.userKeysAllowed() {
			writeErr(w, http.StatusForbidden, "personal API keys are turned off by the admin")
			return
		}
		next(w, r)
	})
}

func (a *App) ownProvider(w http.ResponseWriter, r *http.Request) *Provider {
	p, err := a.store.Provider(pathID(r))
	if err != nil || p.OwnerID == 0 || p.OwnerID != currentUser(r).ID {
		writeErr(w, 404, "not found")
		return nil
	}
	return p
}

func (a *App) privateNetworkBlocked() bool { return a.store.Setting(settingPrivateNetwork) != "1" }

func (a *App) handleListOwn(w http.ResponseWriter, r *http.Request) {
	ps, err := a.store.OwnProviders(currentUser(r).ID)
	writeProviders(w, ps, err)
}

func (a *App) handleCreateOwn(w http.ResponseWriter, r *http.Request) {
	a.saveProvider(w, r, &Provider{OwnerID: currentUser(r).ID})
}

func (a *App) handleUpdateOwn(w http.ResponseWriter, r *http.Request) {
	if p := a.ownProvider(w, r); p != nil {
		a.saveProvider(w, r, p)
	}
}

func (a *App) handleDeleteOwn(w http.ResponseWriter, r *http.Request) {
	if p := a.ownProvider(w, r); p != nil {
		a.store.DeleteProvider(p.ID)
		w.WriteHeader(http.StatusNoContent)
	}
}

func (a *App) handleOwnRemoteModels(w http.ResponseWriter, r *http.Request) {
	var saved *Provider
	if pathID(r) > 0 {
		if saved = a.ownProvider(w, r); saved == nil {
			return
		}
	}
	a.remoteModels(w, r, saved, a.privateNetworkBlocked())
}

// --- settings (admin) ---

func (a *App) settingsView() map[string]bool {
	return map[string]bool{
		settingUserKeys:       a.userKeysAllowed(),
		settingPrivateNetwork: !a.privateNetworkBlocked(),
		"keys_encrypted":      a.store.box != nil,
	}
}

func (a *App) handleGetSettings(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, a.settingsView())
}

func (a *App) handlePutSettings(w http.ResponseWriter, r *http.Request) {
	var req map[string]bool
	if !readJSON(w, r, &req) {
		return
	}
	for _, k := range []string{settingUserKeys, settingPrivateNetwork} {
		if v, ok := req[k]; ok {
			val := "0"
			if v {
				val = "1"
			}
			a.store.SetSetting(k, val)
		}
	}
	writeJSON(w, a.settingsView())
}
