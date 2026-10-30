package main

import (
	"net"
	"net/http"
	"net/netip"
	"strings"
)

// Client IP behind proxies. Forwarding headers are only believed when the
// direct peer is a proxy we trust, otherwise anyone could forge them:
//
//	VSCHAT_TRUST_PROXY=auto (default)  peer is a Cloudflare edge, loopback or private address
//	VSCHAT_TRUST_PROXY=1               always (origin only reachable through your proxy)
//	VSCHAT_TRUST_PROXY=0               never; use the TCP peer address
//
// Header order follows vsterm-cloud: CF-Connecting-IP, True-Client-IP, the
// leftmost X-Forwarded-For hop, X-Real-IP.
var proxyMode = "auto"

// Cloudflare's published edge ranges (https://www.cloudflare.com/ips/).
var cloudflareNets = func() []netip.Prefix {
	var out []netip.Prefix
	for _, s := range []string{
		"173.245.48.0/20", "103.21.244.0/22", "103.22.200.0/22", "103.31.4.0/22", "141.101.64.0/18",
		"108.162.192.0/18", "190.93.240.0/20", "188.114.96.0/20", "197.234.240.0/22", "198.41.128.0/17",
		"162.158.0.0/15", "104.16.0.0/13", "104.24.0.0/14", "172.64.0.0/13", "131.0.72.0/22",
		"2400:cb00::/32", "2606:4700::/32", "2803:f800::/32", "2405:b500::/32", "2405:8100::/32",
		"2a06:98c0::/29", "2c0f:f248::/32",
	} {
		out = append(out, netip.MustParsePrefix(s))
	}
	return out
}()

func peerAddr(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return strings.TrimSpace(r.RemoteAddr)
	}
	return host
}

func trustedPeer(peer string) bool {
	switch proxyMode {
	case "1", "true", "always":
		return true
	case "0", "false", "never":
		return false
	}
	ip, err := netip.ParseAddr(peer)
	if err != nil {
		return false
	}
	ip = ip.Unmap()
	if ip.IsLoopback() || ip.IsPrivate() {
		return true
	}
	for _, n := range cloudflareNets {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}

func validIP(s string) string {
	s = strings.TrimSpace(s)
	if ip, err := netip.ParseAddr(s); err == nil {
		return ip.Unmap().String()
	}
	return ""
}

func clientIP(r *http.Request) string {
	peer := peerAddr(r)
	if !trustedPeer(peer) {
		return peer
	}
	for _, h := range []string{"CF-Connecting-IP", "True-Client-IP"} {
		if ip := validIP(r.Header.Get(h)); ip != "" {
			return ip
		}
	}
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		first, _, _ := strings.Cut(xff, ",")
		if ip := validIP(first); ip != "" {
			return ip
		}
	}
	if ip := validIP(r.Header.Get("X-Real-IP")); ip != "" {
		return ip
	}
	return peer
}

// clientCountry is Cloudflare's two-letter country code, when behind it.
func clientCountry(r *http.Request) string {
	if !trustedPeer(peerAddr(r)) {
		return ""
	}
	c := strings.ToUpper(strings.TrimSpace(r.Header.Get("CF-IPCountry")))
	if len(c) != 2 || c == "XX" {
		return ""
	}
	return c
}
