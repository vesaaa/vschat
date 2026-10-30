package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/anthropics/anthropic-sdk-go"
)

// A stored Claude turn must replay byte-for-byte, thinking signature included.
func TestAnthropicTurnRoundTrip(t *testing.T) {
	var msg anthropic.Message
	if err := json.Unmarshal([]byte(`{"id":"msg_1","type":"message","role":"assistant","model":"claude-opus-5-5",
		"content":[{"type":"thinking","thinking":"hmm","signature":"sig123"},{"type":"text","text":"Hello"}],
		"stop_reason":"end_turn","usage":{"input_tokens":3,"output_tokens":5}}`), &msg); err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(msg.ToParam())
	if err != nil {
		t.Fatal(err)
	}
	var back anthropic.MessageParam
	if err := json.Unmarshal(raw, &back); err != nil {
		t.Fatal(err)
	}
	again, _ := json.Marshal(back)
	if string(again) != string(raw) {
		t.Fatalf("round trip changed:\n%s\n%s", raw, again)
	}
	for _, want := range []string{`"signature":"sig123"`, `"thinking":"hmm"`, `"text":"Hello"`, `"role":"assistant"`} {
		if !strings.Contains(string(raw), want) {
			t.Errorf("missing %s in %s", want, raw)
		}
	}
}

func TestPassword(t *testing.T) {
	h := hashPassword("secret1")
	if !checkPassword(h, "secret1") || checkPassword(h, "secret2") {
		t.Fatal("password check wrong")
	}
}

func TestAdaptiveThinking(t *testing.T) {
	for m, want := range map[string]bool{
		"claude-opus-5-5": true, "claude-haiku-5-5": true, "claude-sonnet-4-6": true,
		"claude-haiku-4-5": false, "claude-3-5-sonnet": false, "gpt-5": false,
	} {
		if adaptiveThinking(m) != want {
			t.Errorf("%s: want %v", m, want)
		}
	}
}
