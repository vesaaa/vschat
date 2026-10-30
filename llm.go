package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

const (
	defaultOpenAIBase    = "https://api.openai.com/v1"
	defaultAnthropicBase = "https://api.anthropic.com"
	anthropicMaxTokens   = 64000
)

// Result is one finished (or interrupted) assistant turn.
type Result struct {
	Content   string
	Reasoning string
	Raw       string // provider-native turn to replay on the next request
	In, Out   int64
	Notice    string      // e.g. refusal or truncation, shown under the answer
	Header    http.Header // upstream response headers (rate limits)
	Cached    int64       // input tokens served from the provider's prompt cache (part of In)
}

type Sink struct {
	Text      func(string)
	Reasoning func(string)
}

// userText inlines attachments ahead of the typed message.
func userText(m *Message) string {
	if len(m.Attachments) == 0 {
		return m.Content
	}
	var b strings.Builder
	for _, a := range m.Attachments {
		fmt.Fprintf(&b, "<file name=%q>\n%s\n</file>\n\n", a.Name, a.Text)
	}
	b.WriteString(m.Content)
	return b.String()
}

func (p *Provider) Stream(ctx context.Context, model string, history []*Message, sink Sink) (Result, error) {
	if p.KeyErr != nil {
		return Result{}, p.KeyErr
	}
	if p.APIKey == "" {
		return Result{}, errors.New("API key not configured for " + p.Name)
	}
	if p.Kind == "anthropic" {
		return p.streamAnthropic(ctx, model, history, sink)
	}
	return p.streamOpenAI(ctx, model, history, sink)
}

func (p *Provider) ListRemoteModels(ctx context.Context) ([]string, error) {
	if p.Kind == "anthropic" {
		client := p.anthropicClient()
		var out []string
		iter := client.Models.ListAutoPaging(ctx, anthropic.ModelListParams{})
		for iter.Next() {
			out = append(out, iter.Current().ID)
		}
		return out, iter.Err()
	}
	req, _ := http.NewRequestWithContext(ctx, "GET", p.openAIBase()+"/models", nil)
	req.Header.Set("Authorization", "Bearer "+p.APIKey)
	resp, err := p.httpClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, upstreamError(resp)
	}
	var body struct {
		Data []struct{ ID string } `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}
	var out []string
	for _, m := range body.Data {
		out = append(out, m.ID)
	}
	sort.Strings(out)
	return out, nil
}

// --- OpenAI-compatible (OpenAI, DeepSeek, most others) ---

func (p *Provider) openAIBase() string {
	if p.BaseURL == "" {
		return defaultOpenAIBase
	}
	return strings.TrimRight(p.BaseURL, "/")
}

func upstreamError(resp *http.Response) error {
	data, _ := io.ReadAll(io.LimitReader(resp.Body, 8192))
	var e struct {
		Error struct{ Message string } `json:"error"`
	}
	if json.Unmarshal(data, &e) == nil && e.Error.Message != "" {
		return fmt.Errorf("%d: %s", resp.StatusCode, e.Error.Message)
	}
	return fmt.Errorf("%d: %s", resp.StatusCode, strings.TrimSpace(string(data)))
}

func (p *Provider) streamOpenAI(ctx context.Context, model string, history []*Message, sink Sink) (Result, error) {
	type msg struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	msgs := make([]msg, 0, len(history))
	for _, m := range history {
		if m.Role == "user" {
			msgs = append(msgs, msg{"user", userText(m)})
		} else if m.Content != "" {
			msgs = append(msgs, msg{"assistant", m.Content})
		}
	}
	body, _ := json.Marshal(map[string]any{
		"model":          model,
		"messages":       msgs,
		"stream":         true,
		"stream_options": map[string]any{"include_usage": true},
	})
	req, _ := http.NewRequestWithContext(ctx, "POST", p.openAIBase()+"/chat/completions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.APIKey)
	resp, err := p.httpClient().Do(req)
	if err != nil {
		return Result{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return Result{Header: resp.Header}, upstreamError(resp)
	}

	res := Result{Header: resp.Header}
	var text, reasoning strings.Builder
	sc := bufio.NewScanner(resp.Body)
	sc.Buffer(make([]byte, 64*1024), 4*1024*1024)
	for sc.Scan() {
		line := sc.Text()
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimSpace(line[5:])
		if data == "[DONE]" {
			break
		}
		var chunk struct {
			Choices []struct {
				Delta struct {
					Content          string `json:"content"`
					ReasoningContent string `json:"reasoning_content"`
					Reasoning        string `json:"reasoning"`
				} `json:"delta"`
				FinishReason string `json:"finish_reason"`
			} `json:"choices"`
			Usage *struct {
				PromptTokens int64 `json:"prompt_tokens"`
				CacheHit     int64 `json:"prompt_cache_hit_tokens"` // DeepSeek
				Details      struct {
					Cached int64 `json:"cached_tokens"` // OpenAI and most compatibles
				} `json:"prompt_tokens_details"`
				CompletionTokens int64 `json:"completion_tokens"`
			} `json:"usage"`
			Error *struct{ Message string } `json:"error"`
		}
		if json.Unmarshal([]byte(data), &chunk) != nil {
			continue
		}
		if chunk.Error != nil {
			err = errors.New(chunk.Error.Message)
			break
		}
		for _, c := range chunk.Choices {
			if r := c.Delta.ReasoningContent + c.Delta.Reasoning; r != "" {
				reasoning.WriteString(r)
				sink.Reasoning(r)
			}
			if c.Delta.Content != "" {
				text.WriteString(c.Delta.Content)
				sink.Text(c.Delta.Content)
			}
			if c.FinishReason == "length" {
				res.Notice = "max_tokens"
			} else if c.FinishReason == "content_filter" {
				res.Notice = "refusal"
			}
		}
		if chunk.Usage != nil {
			res.In, res.Out = chunk.Usage.PromptTokens, chunk.Usage.CompletionTokens
			res.Cached = max(chunk.Usage.CacheHit, chunk.Usage.Details.Cached)
		}
	}
	if err == nil {
		err = sc.Err()
	}
	res.Content, res.Reasoning = text.String(), reasoning.String()
	return res, err
}

// --- Anthropic (Claude) via the official SDK ---

func (p *Provider) anthropicBase() string {
	if p.BaseURL == "" {
		return defaultAnthropicBase
	}
	return strings.TrimRight(p.BaseURL, "/")
}

func (p *Provider) anthropicClient() anthropic.Client {
	return anthropic.NewClient(
		option.WithBaseURL(p.anthropicBase()),
		option.WithAPIKey(p.APIKey),
		option.WithHTTPClient(p.httpClient()),
		option.WithRequestTimeout(30*time.Minute),
	)
}

// adaptiveThinking reports whether the model takes thinking {type: adaptive}
// (Claude 4.6 and newer). Older models are sent without a thinking config.
func adaptiveThinking(model string) bool {
	for _, old := range []string{"claude-3", "-4-5", "-4-1", "-4-0", "opus-4-2", "sonnet-4-2", "haiku-4-2"} {
		if strings.Contains(model, old) {
			return false
		}
	}
	return strings.HasPrefix(model, "claude-")
}

// serverFallback reports models that accept fallbacks:"default", which lets the
// API re-serve a safety-classifier refusal on another model in the same call.
func serverFallback(model string) bool {
	switch model {
	case "claude-fable-5-1", "claude-opus-5-5", "claude-opus-5", "claude-sonnet-5-5":
		return true
	}
	return false
}

func (p *Provider) streamAnthropic(ctx context.Context, model string, history []*Message, sink Sink) (Result, error) {
	var msgs []anthropic.MessageParam
	for _, m := range history {
		if m.Role == "user" {
			msgs = append(msgs, anthropic.NewUserMessage(anthropic.NewTextBlock(userText(m))))
			continue
		}
		// Replay Claude's own turns unchanged (thinking blocks included) so the
		// conversation stays append-only; other providers' turns go in as text.
		if m.Raw != "" {
			var mp anthropic.MessageParam
			if json.Unmarshal([]byte(m.Raw), &mp) == nil {
				msgs = append(msgs, mp)
				continue
			}
		}
		if m.Content != "" {
			msgs = append(msgs, anthropic.NewAssistantMessage(anthropic.NewTextBlock(m.Content)))
		}
	}

	params := anthropic.MessageNewParams{
		Model:     anthropic.Model(model),
		MaxTokens: anthropicMaxTokens,
		Messages:  msgs,
		// Automatic prompt caching: the breakpoint moves to the newest message
		// each turn, and because the history is replayed unchanged the next turn
		// reads everything before it from cache. Default 5-minute TTL.
		CacheControl: anthropic.NewCacheControlEphemeralParam(),
	}
	if adaptiveThinking(model) {
		adaptive := anthropic.ThinkingConfigAdaptiveParam{Display: anthropic.ThinkingConfigAdaptiveDisplaySummarized}
		params.Thinking = anthropic.ThinkingConfigParamUnion{OfAdaptive: &adaptive}
	} else {
		params.MaxTokens = 8192
	}
	var opts []option.RequestOption
	if serverFallback(model) && p.anthropicBase() == defaultAnthropicBase {
		opts = append(opts,
			option.WithHeaderAdd("anthropic-beta", "server-side-fallback-2026-07-01"),
			option.WithJSONSet("fallbacks", "default"))
	}

	var httpResp *http.Response
	opts = append(opts, option.WithResponseInto(&httpResp))
	client := p.anthropicClient()
	stream := client.Messages.NewStreaming(ctx, params, opts...)
	defer stream.Close()

	var res Result
	var text, reasoning strings.Builder
	acc := anthropic.Message{}
	for stream.Next() {
		ev := stream.Current()
		acc.Accumulate(ev)
		switch e := ev.AsAny().(type) {
		case anthropic.ContentBlockDeltaEvent:
			switch d := e.Delta.AsAny().(type) {
			case anthropic.TextDelta:
				text.WriteString(d.Text)
				sink.Text(d.Text)
			case anthropic.ThinkingDelta:
				reasoning.WriteString(d.Thinking)
				sink.Reasoning(d.Thinking)
			}
		}
	}
	err := stream.Err()
	if httpResp != nil {
		res.Header = httpResp.Header
	}
	res.Content, res.Reasoning = text.String(), reasoning.String()
	res.In = acc.Usage.InputTokens + acc.Usage.CacheReadInputTokens + acc.Usage.CacheCreationInputTokens
	res.Cached = acc.Usage.CacheReadInputTokens
	res.Out = acc.Usage.OutputTokens

	switch acc.StopReason {
	case anthropic.StopReasonRefusal:
		// A refused turn is not replayed; the next request starts clean.
		res.Notice = "refusal"
		return res, err
	case anthropic.StopReasonMaxTokens:
		res.Notice = "max_tokens"
	}
	if err == nil && acc.StopReason != "" {
		if raw, mErr := json.Marshal(acc.ToParam()); mErr == nil {
			res.Raw = string(raw)
		}
	}
	return res, err
}
