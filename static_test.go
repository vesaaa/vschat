package main

import (
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

func TestStaticFingerprints(t *testing.T) {
	a := testApp(t, "")
	h := a.routes()
	get := func(path string, hdr ...string) *httptest.ResponseRecorder {
		r := httptest.NewRequest("GET", path, nil)
		for i := 0; i+1 < len(hdr); i += 2 {
			r.Header.Set(hdr[i], hdr[i+1])
		}
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, r)
		return rec
	}
	idx := get("/")
	if idx.Code != 200 || idx.Header().Get("Cache-Control") != "no-cache" || idx.Header().Get("ETag") == "" {
		t.Fatalf("index: %d %v", idx.Code, idx.Header())
	}
	body := idx.Body.String()
	refs := regexp.MustCompile(`(?:src|href)="(/s/[0-9a-f]{12}/[^"]+)"`).FindAllStringSubmatch(body, -1)
	if len(refs) != 6 { // favicon, css, 3 vendor scripts, app.js
		t.Fatalf("expected 6 fingerprinted refs, got %d", len(refs))
	}
	for _, m := range refs {
		rec := get(m[1])
		if rec.Code != 200 || !strings.Contains(rec.Header().Get("Cache-Control"), "immutable") {
			t.Errorf("%s: %d %q", m[1], rec.Code, rec.Header().Get("Cache-Control"))
		}
	}
	// favicon referenced from CSS resolves next to the stylesheet
	css := strings.Replace(refs[1][1], "style.css", "favicon.svg", 1)
	if rec := get(css); rec.Code != 200 {
		t.Errorf("relative favicon: %d", rec.Code)
	}
	if rec := get("/", "If-None-Match", idx.Header().Get("ETag")); rec.Code != 304 {
		t.Errorf("revalidation: %d", rec.Code)
	}
	if rec := get("/s/oldhash00000/app.js"); rec.Code != 200 || rec.Header().Get("Cache-Control") != "no-cache" {
		t.Errorf("stale hash: %d %q", rec.Code, rec.Header().Get("Cache-Control"))
	}
	if rec := get("/s/x/../main.go"); rec.Code == 200 {
		t.Error("path escape served")
	}
	if rec := get("/api/health"); rec.Header().Get("Cache-Control") != "no-store" {
		t.Errorf("api cache header %q", rec.Header().Get("Cache-Control"))
	}
	if rec := get("/app.js"); rec.Code != 404 {
		t.Errorf("unversioned asset still served: %d", rec.Code)
	}
}
