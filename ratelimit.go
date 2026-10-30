package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Provider-side headroom for admins: rate-limit headers from the most recent
// response (kept in memory) and, where the API offers it, the account balance.

type RateInfo struct {
	ReqLimit     int64 `json:"req_limit,omitempty"`
	ReqRemaining int64 `json:"req_remaining,omitempty"`
	ReqReset     int64 `json:"req_reset,omitempty"` // unix seconds
	TokLimit     int64 `json:"tok_limit,omitempty"`
	TokRemaining int64 `json:"tok_remaining,omitempty"`
	TokReset     int64 `json:"tok_reset,omitempty"`
	UpdatedAt    int64 `json:"updated_at"`
}

type rateCache struct {
	mu sync.Mutex
	m  map[int64]RateInfo
}

func (c *rateCache) set(id int64, r RateInfo) {
	c.mu.Lock()
	c.m[id] = r
	c.mu.Unlock()
}

func (c *rateCache) get(id int64) (RateInfo, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	r, ok := c.m[id]
	return r, ok
}

func headerInt(h http.Header, keys ...string) int64 {
	for _, k := range keys {
		if v := strings.TrimSpace(h.Get(k)); v != "" {
			if n, err := strconv.ParseInt(v, 10, 64); err == nil {
				return n
			}
		}
	}
	return 0
}

// headerReset reads a reset that is either a duration ("6m0s", "20ms", "12")
// or an RFC 3339 time, as unix seconds.
func headerReset(h http.Header, at time.Time, keys ...string) int64 {
	for _, k := range keys {
		v := strings.TrimSpace(h.Get(k))
		if v == "" {
			continue
		}
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			return t.Unix()
		}
		if d, err := time.ParseDuration(v); err == nil {
			return at.Add(d).Unix()
		}
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return at.Add(time.Duration(f * float64(time.Second))).Unix()
		}
	}
	return 0
}

// parseRate understands OpenAI-style x-ratelimit-* and Anthropic's anthropic-ratelimit-*.
func parseRate(h http.Header) (RateInfo, bool) {
	if h == nil {
		return RateInfo{}, false
	}
	at := time.Now()
	r := RateInfo{
		ReqLimit:     headerInt(h, "x-ratelimit-limit-requests", "anthropic-ratelimit-requests-limit"),
		ReqRemaining: headerInt(h, "x-ratelimit-remaining-requests", "anthropic-ratelimit-requests-remaining"),
		ReqReset:     headerReset(h, at, "x-ratelimit-reset-requests", "anthropic-ratelimit-requests-reset"),
		TokLimit:     headerInt(h, "x-ratelimit-limit-tokens", "anthropic-ratelimit-tokens-limit", "anthropic-ratelimit-input-tokens-limit"),
		TokRemaining: headerInt(h, "x-ratelimit-remaining-tokens", "anthropic-ratelimit-tokens-remaining", "anthropic-ratelimit-input-tokens-remaining"),
		TokReset:     headerReset(h, at, "x-ratelimit-reset-tokens", "anthropic-ratelimit-tokens-reset", "anthropic-ratelimit-input-tokens-reset"),
		UpdatedAt:    at.Unix(),
	}
	return r, r.ReqLimit > 0 || r.TokLimit > 0
}

type Balance struct {
	Currency string `json:"currency"`
	Amount   string `json:"amount"`
}

func balanceSupported(p *Provider) bool { return p.Preset == "deepseek" || p.Preset == "kimi" }

// fetchBalance asks DeepSeek or Kimi for the account balance.
func (p *Provider) fetchBalance(ctx context.Context) ([]Balance, error) {
	path := map[string]string{"deepseek": "/user/balance", "kimi": "/users/me/balance"}[p.Preset]
	if path == "" {
		return nil, fmt.Errorf("balance not available for this provider")
	}
	req, _ := http.NewRequestWithContext(ctx, "GET", p.openAIBase()+path, nil)
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
		BalanceInfos []struct {
			Currency     string `json:"currency"`
			TotalBalance string `json:"total_balance"`
		} `json:"balance_infos"`
		Data struct {
			Available *float64 `json:"available_balance"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}
	var out []Balance
	for _, b := range body.BalanceInfos {
		out = append(out, Balance{b.Currency, b.TotalBalance})
	}
	if v := body.Data.Available; v != nil {
		out = append(out, Balance{"CNY", strconv.FormatFloat(*v, 'f', 2, 64)})
	}
	return out, nil
}

func (a *App) handleProviderStatus(w http.ResponseWriter, r *http.Request) {
	p := a.sharedProvider(w, r)
	if p == nil {
		return
	}
	out := map[string]any{}
	if rate, ok := a.rates.get(p.ID); ok {
		out["rate"] = rate
	}
	if balanceSupported(p) && p.APIKey != "" && p.KeyErr == nil {
		ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
		defer cancel()
		if bal, err := p.fetchBalance(ctx); err != nil {
			out["balance_error"] = err.Error()
		} else {
			out["balance"] = bal
		}
	}
	writeJSON(w, out)
}
