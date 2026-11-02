package main

import (
	"net/http"
	"strings"
)

// Upstream requests carry only what the API needs: the key, content type,
// API version/beta headers and the request body. Nothing about the user (IP,
// browser, account) is ever added, and the SDK's own telemetry headers
// (X-Stainless-*: OS, CPU arch, Go version, SDK version) are stripped here.

const upstreamUA = "VsChat"

type scrubTransport struct{ base http.RoundTripper }

func (t scrubTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r = r.Clone(r.Context())
	for k := range r.Header {
		if strings.HasPrefix(http.CanonicalHeaderKey(k), "X-Stainless-") {
			r.Header.Del(k)
		}
	}
	r.Header.Set("User-Agent", upstreamUA)
	return t.base.RoundTrip(r)
}

var upstreamClient = &http.Client{Transport: scrubTransport{http.DefaultTransport}}
