package main

import (
	"slices"
	"strings"
)

// The built-in provider catalog. Providers created from a preset follow its
// model list unless an admin picks models by hand, so updating this file and
// shipping a new release is all it takes to roll out new models.
//
// Chinese providers default to their mainland endpoints, the others to their
// official global ones. All but Claude speak the OpenAI Chat Completions API.
//
// Checked against each provider's docs on 2026-10-30.

type Preset struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Kind    string   `json:"kind"`
	BaseURL string   `json:"base_url"`
	Region  string   `json:"region"` // "cn" or "global"
	Models  []string `json:"models"`
	Default string   `json:"default"` // mid-tier pick preselected for new users; "" = first
	KeyURL  string   `json:"key_url"`
	KeyEnv  string   `json:"-"` // seeds the key on first start
	Color   string   `json:"color"`
}

var catalog = []Preset{
	{ID: "openai", Name: "OpenAI", Kind: "openai", BaseURL: "https://api.openai.com/v1", Region: "global",
		Models: []string{"gpt-6.1-sol", "gpt-6-astra", "gpt-6.1-terra", "gpt-6-luna"}, Default: "gpt-6.1-sol",
		KeyURL: "https://platform.openai.com/api-keys", KeyEnv: "OPENAI_API_KEY", Color: "#10a37f"},
	{ID: "anthropic", Name: "Claude", Kind: "anthropic", BaseURL: "https://api.anthropic.com", Region: "global",
		Models: []string{"claude-opus-5-5", "claude-sonnet-5-5", "claude-haiku-5-5"}, Default: "claude-sonnet-5-5",
		KeyURL: "https://console.anthropic.com/settings/keys", KeyEnv: "ANTHROPIC_API_KEY", Color: "#d97757"},
	{ID: "gemini", Name: "Gemini", Kind: "openai", BaseURL: "https://generativelanguage.googleapis.com/v1beta/openai", Region: "global",
		Models: []string{"gemini-4-argon", "gemini-3.8-flash", "gemini-3.5-flash-lite"}, Default: "gemini-3.8-flash",
		KeyURL: "https://aistudio.google.com/apikey", KeyEnv: "GEMINI_API_KEY", Color: "#4285f4"},
	{ID: "xai", Name: "Grok", Kind: "openai", BaseURL: "https://api.x.ai/v1", Region: "global",
		Models: []string{"grok-4.7", "grok-4.6"},
		KeyURL: "https://console.x.ai", KeyEnv: "XAI_API_KEY", Color: "#1f1f1f"},
	{ID: "deepseek", Name: "DeepSeek", Kind: "openai", BaseURL: "https://api.deepseek.com", Region: "cn",
		Models: []string{"deepseek-v4-pro", "deepseek-flash"}, Default: "deepseek-flash",
		KeyURL: "https://platform.deepseek.com/api_keys", KeyEnv: "DEEPSEEK_API_KEY", Color: "#4d6bfe"},
	{ID: "qwen", Name: "通义千问 Qwen", Kind: "openai", BaseURL: "https://dashscope.aliyuncs.com/compatible-mode/v1", Region: "cn",
		Models: []string{"qwen3.8-max", "qwen3.7-plus", "qwen3.8-flash"}, Default: "qwen3.7-plus",
		KeyURL: "https://bailian.console.aliyun.com/?apiKey=1", KeyEnv: "DASHSCOPE_API_KEY", Color: "#615ced"},
	{ID: "kimi", Name: "Kimi", Kind: "openai", BaseURL: "https://api.moonshot.cn/v1", Region: "cn",
		Models: []string{"kimi-k3", "kimi-k2.6"},
		KeyURL: "https://platform.moonshot.cn/console/api-keys", KeyEnv: "MOONSHOT_API_KEY", Color: "#16191e"},
	{ID: "glm", Name: "智谱 GLM", Kind: "openai", BaseURL: "https://open.bigmodel.cn/api/paas/v4", Region: "cn",
		Models: []string{"glm-5.3", "glm-5.3-flash"}, Default: "glm-5.3-flash",
		KeyURL: "https://bigmodel.cn/usercenter/proj-mgmt/apikeys", KeyEnv: "ZHIPU_API_KEY", Color: "#3859ff"},
}

func presetByID(id string) *Preset {
	for i := range catalog {
		if catalog[i].ID == id {
			return &catalog[i]
		}
	}
	return nil
}

func normURL(u string) string { return strings.TrimRight(strings.TrimSpace(u), "/") }

// presetFor finds the preset whose official endpoint matches kind + base URL.
func presetFor(kind, baseURL string) *Preset {
	base := normURL(baseURL)
	for i := range catalog {
		pr := &catalog[i]
		if pr.Kind != kind {
			continue
		}
		if base == pr.BaseURL || base == "" && (pr.ID == "openai" || pr.ID == "anthropic") {
			return pr
		}
	}
	return nil
}

// isOfficial reports whether kind + base URL is one of the catalog endpoints.
func isOfficial(kind, baseURL string) bool { return presetFor(kind, baseURL) != nil }

// legacySeeds are the model lists v0.1/v0.2 seeded; providers still on them
// switch to following the catalog.
var legacySeeds = [][]string{
	{"gpt-5", "gpt-5-mini"},
	{"claude-opus-5-5", "claude-sonnet-5-5", "claude-haiku-5-5"},
	{"deepseek-chat", "deepseek-reasoner"},
}

func isLegacySeed(models []string) bool {
	for _, s := range legacySeeds {
		if slices.Equal(s, models) {
			return true
		}
	}
	return false
}
