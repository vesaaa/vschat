package main

import (
	"errors"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"syscall"
	"time"
)

// Users' own providers make the server fetch a URL the user chose. Unless an
// admin allows it, those requests may only reach public addresses, so nobody can
// point VsChat at the host itself or the internal network. The check runs on the
// resolved address at connect time, which also covers DNS rebinding.

var errBlockedAddr = errors.New("address not allowed: private and local networks are blocked for personal providers")

var cgnat = netip.MustParsePrefix("100.64.0.0/10")

func publicAddr(ip netip.Addr) bool {
	ip = ip.Unmap()
	return ip.IsGlobalUnicast() && !ip.IsPrivate() && !cgnat.Contains(ip)
}

var guardedClient = &http.Client{
	Transport: &http.Transport{
		Proxy: nil, // a proxy would hide the real destination from the check
		DialContext: (&net.Dialer{
			Timeout:   15 * time.Second,
			KeepAlive: 30 * time.Second,
			Control: func(network, address string, _ syscall.RawConn) error {
				ap, err := netip.ParseAddrPort(address)
				if err != nil || !publicAddr(ap.Addr()) {
					return errBlockedAddr
				}
				return nil
			},
		}).DialContext,
		ForceAttemptHTTP2:     true,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 2 * time.Minute,
		IdleConnTimeout:       90 * time.Second,
	},
}

func (p *Provider) httpClient() *http.Client {
	if p.guard {
		return guardedClient
	}
	return http.DefaultClient
}

func validBaseURL(s string) bool {
	if s == "" {
		return true
	}
	u, err := url.Parse(s)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Host != "" && u.User == nil
}
