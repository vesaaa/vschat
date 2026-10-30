package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	maxAttachments    = 10
	maxAttachmentSize = 1 << 20 // characters of text per attachment
)

func encodeAttachments(a []Attachment) string {
	if len(a) == 0 {
		return ""
	}
	b, _ := json.Marshal(a)
	return string(b)
}

func decodeAttachments(s string) []Attachment {
	if s == "" {
		return nil
	}
	var a []Attachment
	json.Unmarshal([]byte(s), &a)
	return a
}

func pathID(r *http.Request) int64 {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	return id
}

// ModelOption is one entry in the model picker: "<providerID>/<model>".
type ModelOption struct {
	ID       string `json:"id"`
	Model    string `json:"model"`
	Provider string `json:"provider"`
	Kind     string `json:"kind"`
	Own      bool   `json:"own"` // on the user's own key
	Preset   string `json:"preset,omitempty"`
	Default  bool   `json:"default,omitempty"` // the provider's mid-tier pick
}

func (a *App) modelOptions(u *User) []ModelOption {
	ps, _ := a.store.UsableProviders(u.ID)
	out := []ModelOption{}
	for _, p := range ps {
		if !a.canUse(u, p) || p.APIKey == "" {
			continue
		}
		def := ""
		if pr := presetByID(p.Preset); pr != nil {
			def = pr.Default
		}
		for _, m := range p.Models {
			out = append(out, ModelOption{ID: fmt.Sprintf("%d/%s", p.ID, m), Model: m, Provider: p.Name, Kind: p.Kind,
				Own: p.OwnerID != 0, Preset: p.Preset, Default: m == def})
		}
	}
	return out
}

func (a *App) resolveModel(u *User, ref string) (*Provider, string, error) {
	pid, model, ok := strings.Cut(ref, "/")
	if !ok {
		return nil, "", fmt.Errorf("unknown model")
	}
	id, _ := strconv.ParseInt(pid, 10, 64)
	p, err := a.store.Provider(id)
	if err != nil || !a.canUse(u, p) {
		return nil, "", fmt.Errorf("model not available")
	}
	for _, m := range p.Models {
		if m == model {
			return p, model, nil
		}
	}
	return nil, "", fmt.Errorf("model not available")
}

func (a *App) handleModels(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, a.modelOptions(currentUser(r)))
}

func (a *App) handleListChats(w http.ResponseWriter, r *http.Request) {
	chats, err := a.store.ListChats(currentUser(r).ID)
	if err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, chats)
}

func (a *App) handleGetChat(w http.ResponseWriter, r *http.Request) {
	c, err := a.store.Chat(currentUser(r).ID, pathID(r))
	if err != nil {
		writeErr(w, 404, "not found")
		return
	}
	msgs, err := a.store.Messages(c.ID)
	if err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, map[string]any{"chat": c, "messages": msgs})
}

func (a *App) handleUpdateChat(w http.ResponseWriter, r *http.Request) {
	c, err := a.store.Chat(currentUser(r).ID, pathID(r))
	if err != nil {
		writeErr(w, 404, "not found")
		return
	}
	var req struct {
		Title *string `json:"title"`
		Model *string `json:"model"`
	}
	if !readJSON(w, r, &req) {
		return
	}
	if req.Title != nil {
		c.Title = truncate(strings.TrimSpace(*req.Title), 100)
	}
	if req.Model != nil {
		c.Model = *req.Model
	}
	a.store.UpdateChat(c)
	writeJSON(w, c)
}

func (a *App) handleDeleteChat(w http.ResponseWriter, r *http.Request) {
	a.store.DeleteChat(currentUser(r).ID, pathID(r))
	w.WriteHeader(http.StatusNoContent)
}

func truncate(s string, n int) string {
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	return string([]rune(s)[:n])
}

func titleFrom(m *Message) string {
	t := strings.Join(strings.Fields(m.Content), " ")
	if t == "" && len(m.Attachments) > 0 {
		t = m.Attachments[0].Name
	}
	return truncate(t, 40)
}

// handleSend appends a user message (or regenerates the last answer) and
// streams the assistant reply back as server-sent events.
func (a *App) handleSend(w http.ResponseWriter, r *http.Request) {
	u := currentUser(r)
	var req struct {
		ChatID      int64        `json:"chat_id"`
		Content     string       `json:"content"`
		Attachments []Attachment `json:"attachments"`
		Model       string       `json:"model"`
		Regenerate  bool         `json:"regenerate"`
	}
	r.Body = http.MaxBytesReader(w, r.Body, 16<<20)
	if !readJSON(w, r, &req) {
		return
	}
	p, model, err := a.resolveModel(u, req.Model)
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	own := p.OwnerID != 0
	if u.TokenLimit > 0 && !own {
		if used, _ := a.store.MonthTokens(u.ID); used >= u.TokenLimit {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(429)
			json.NewEncoder(w).Encode(map[string]any{"error": "monthly token limit reached", "code": "quota_exceeded", "reset_at": monthReset()})
			return
		}
	}
	if len(req.Attachments) > maxAttachments {
		writeErr(w, 400, "too many attachments")
		return
	}
	for _, at := range req.Attachments {
		if len(at.Text) > maxAttachmentSize {
			writeErr(w, 400, "attachment too large: "+at.Name)
			return
		}
	}

	var chat *Chat
	if req.ChatID == 0 {
		chat, err = a.store.CreateChat(u.ID, req.Model)
	} else {
		chat, err = a.store.Chat(u.ID, req.ChatID)
	}
	if err != nil {
		writeErr(w, 404, "chat not found")
		return
	}
	history, err := a.store.Messages(chat.ID)
	if err != nil {
		writeErr(w, 500, err.Error())
		return
	}

	var userMsg *Message
	if req.Regenerate {
		if n := len(history); n > 0 && history[n-1].Role == "assistant" {
			a.store.DeleteMessage(history[n-1].ID)
			history = history[:n-1]
		}
		if len(history) == 0 || history[len(history)-1].Role != "user" {
			writeErr(w, 400, "nothing to regenerate")
			return
		}
	} else {
		if strings.TrimSpace(req.Content) == "" && len(req.Attachments) == 0 {
			writeErr(w, 400, "empty message")
			return
		}
		userMsg = &Message{ChatID: chat.ID, Role: "user", Content: req.Content, Attachments: req.Attachments}
		if err := a.store.AddMessage(userMsg); err != nil {
			writeErr(w, 500, err.Error())
			return
		}
		history = append(history, userMsg)
	}
	if chat.Title == "" {
		chat.Title = titleFrom(history[0])
	}
	chat.Model = req.Model
	chat.UpdatedAt = now()
	a.store.UpdateChat(chat)

	flusher, _ := w.(http.Flusher)
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Accel-Buffering", "no")
	send := func(v map[string]any) {
		b, _ := json.Marshal(v)
		fmt.Fprintf(w, "data: %s\n\n", b)
		if flusher != nil {
			flusher.Flush()
		}
	}
	start := map[string]any{"type": "start", "chat": chat}
	if userMsg != nil {
		start["user_message"] = userMsg
	}
	send(start)

	// Keep the connection alive through long thinking pauses.
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	events := make(chan map[string]any, 64)
	done := make(chan struct{})
	var res Result
	var streamErr error
	started := time.Now()
	var firstText time.Duration // time to the first answer token: what the UI shows as thinking
	go func() {
		defer close(done)
		res, streamErr = p.Stream(ctx, model, history, Sink{
			Text: func(t string) {
				if firstText == 0 {
					firstText = time.Since(started)
				}
				events <- map[string]any{"type": "text", "t": t}
			},
			Reasoning: func(t string) { events <- map[string]any{"type": "reasoning", "t": t} },
		})
	}()
	tick := time.NewTicker(15 * time.Second)
	defer tick.Stop()
loop:
	for {
		select {
		case ev := <-events:
			send(ev)
		case <-tick.C:
			fmt.Fprint(w, ": ping\n\n")
			if flusher != nil {
				flusher.Flush()
			}
		case <-done:
			for len(events) > 0 {
				send(<-events)
			}
			break loop
		}
	}

	interrupted := r.Context().Err() != nil
	reply := &Message{
		ChatID: chat.ID, Role: "assistant", Content: res.Content, Reasoning: res.Reasoning,
		Model: model, Raw: res.Raw, TokensIn: res.In, TokensOut: res.Out,
		ThinkMs: firstText.Milliseconds(), DurationMs: time.Since(started).Milliseconds(),
	}
	switch {
	case interrupted:
		reply.Raw = ""
		reply.Error = "stopped"
	case streamErr != nil:
		reply.Raw = ""
		reply.Error = streamErr.Error()
		log.Printf("chat %d: %s: %v", chat.ID, model, streamErr)
	case res.Notice != "":
		reply.Error = res.Notice
	}
	if err := a.store.AddMessage(reply); err != nil {
		log.Printf("save reply: %v", err)
	}
	if res.In+res.Out > 0 {
		a.store.AddUsage(u.ID, res.In, res.Out, own)
	}
	if rate, ok := parseRate(res.Header); ok && !own {
		a.rates.set(p.ID, rate)
	}
	detail := model
	if res.In+res.Out > 0 {
		detail += fmt.Sprintf(" · %s→%s", fmtTokens(res.In), fmtTokens(res.Out))
	}
	if res.Cached > 0 {
		detail += " · cache " + fmtTokens(res.Cached)
	}
	if own {
		detail += " · own key"
	}
	if reply.Error != "" {
		detail += " · " + truncate(reply.Error, 120)
	}
	a.logAccess(r, u.ID, u.Username, "chat", detail, streamErr == nil && !interrupted)
	if !interrupted {
		send(map[string]any{"type": "done", "message": reply})
	}
}

// monthReset is when monthly usage starts over: the 1st of next month, server time.
func monthReset() int64 {
	t := time.Now()
	return time.Date(t.Year(), t.Month()+1, 1, 0, 0, 0, 0, t.Location()).Unix()
}

func fmtTokens(n int64) string {
	switch {
	case n >= 1_000_000:
		return fmt.Sprintf("%.1fM", float64(n)/1e6)
	case n >= 1000:
		return fmt.Sprintf("%.1fK", float64(n)/1e3)
	}
	return strconv.FormatInt(n, 10)
}
