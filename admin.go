package main

import (
	"net/http"
	"strings"
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
