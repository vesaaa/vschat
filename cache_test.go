package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// A fake Messages endpoint: checks the request asks for automatic caching and
// that earlier Claude turns are replayed verbatim, then streams a reply whose
// usage reports a cache hit.
func TestAnthropicPromptCache(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		json.Unmarshal(b, &body)
		w.Header().Set("Content-Type", "text/event-stream")
		ev := func(name, data string) { fmt.Fprintf(w, "event: %s\ndata: %s\n\n", name, data) }
		ev("message_start", `{"type":"message_start","message":{"id":"m1","type":"message","role":"assistant","model":"claude-sonnet-5-5","content":[],"stop_reason":null,"usage":{"input_tokens":12,"cache_read_input_tokens":2048,"cache_creation_input_tokens":40,"output_tokens":1}}}`)
		ev("content_block_start", `{"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}`)
		ev("content_block_delta", `{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"hi"}}`)
		ev("content_block_stop", `{"type":"content_block_stop","index":0}`)
		ev("message_delta", `{"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"output_tokens":5}}`)
		ev("message_stop", `{"type":"message_stop"}`)
	}))
	defer srv.Close()

	prev := `{"role":"assistant","content":[{"type":"thinking","thinking":"t","signature":"sig"},{"type":"text","text":"earlier"}]}`
	history := []*Message{
		{Role: "user", Content: "q1"},
		{Role: "assistant", Content: "earlier", Raw: prev},
		{Role: "user", Content: "q2"},
	}
	p := &Provider{Kind: "anthropic", BaseURL: srv.URL, APIKey: "k"}
	res, err := p.Stream(context.Background(), "claude-sonnet-5-5", history, Sink{Text: func(string) {}, Reasoning: func(string) {}})
	if err != nil {
		t.Fatal(err)
	}
	cc, _ := body["cache_control"].(map[string]any)
	if cc["type"] != "ephemeral" {
		t.Errorf("request has no top-level cache_control: %v", body["cache_control"])
	}
	msgs, _ := json.Marshal(body["messages"])
	if !strings.Contains(string(msgs), `"signature":"sig"`) {
		t.Errorf("earlier Claude turn not replayed verbatim: %s", msgs)
	}
	if res.Cached != 2048 || res.In != 12+2048+40 || res.Out != 5 || res.Content != "hi" {
		t.Errorf("result %+v", res)
	}
}
