package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
)

// Only these headers may reach a provider; anything else could leak details
// about the server or its users.
var allowedUpstream = map[string]bool{
	"Accept": true, "Accept-Encoding": true, "Authorization": true, "Content-Length": true,
	"Content-Type": true, "User-Agent": true, "X-Api-Key": true,
	"Anthropic-Version": true, "Anthropic-Beta": true, "Idempotency-Key": true,
}

func checkHeaders(t *testing.T, h http.Header) {
	t.Helper()
	var bad []string
	for k := range h {
		if !allowedUpstream[k] {
			bad = append(bad, k)
		}
	}
	sort.Strings(bad)
	if len(bad) > 0 {
		t.Errorf("unexpected upstream headers: %v", bad)
	}
	if ua := h.Get("User-Agent"); ua != upstreamUA {
		t.Errorf("User-Agent = %q", ua)
	}
}

func TestUpstreamHeaders(t *testing.T) {
	var got []http.Header
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = append(got, r.Header.Clone())
		w.Header().Set("Content-Type", "text/event-stream")
		if strings.Contains(r.URL.Path, "/messages") {
			w.Write([]byte("event: message_start\ndata: {\"type\":\"message_start\",\"message\":{\"id\":\"m\",\"type\":\"message\",\"role\":\"assistant\",\"model\":\"x\",\"content\":[],\"usage\":{\"input_tokens\":1,\"output_tokens\":0}}}\n\nevent: message_stop\ndata: {\"type\":\"message_stop\"}\n\n"))
			return
		}
		w.Write([]byte("data: [DONE]\n\n"))
	}))
	defer srv.Close()

	hist := []*Message{{Role: "user", Content: "hi"}}
	nop := Sink{Text: func(string) {}, Reasoning: func(string) {}}
	for _, kind := range []string{"openai", "anthropic"} {
		got = nil
		p := &Provider{Kind: kind, BaseURL: srv.URL, APIKey: "k"}
		p.Stream(context.Background(), "claude-sonnet-5-5", hist, nop)
		if len(got) == 0 {
			t.Fatalf("%s: no request", kind)
		}
		t.Logf("%s headers: %v", kind, got[0])
		checkHeaders(t, got[0])
	}
}
