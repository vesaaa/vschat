package main

import (
	"compress/gzip"
	"io"
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

	// gzip: compressed when accepted, identical once decompressed, never for the API stream
	plain := get(refs[5][1])
	zipped := get(refs[5][1], "Accept-Encoding", "gzip, deflate, br")
	if zipped.Header().Get("Content-Encoding") != "gzip" || zipped.Body.Len() >= plain.Body.Len()/2 {
		t.Fatalf("app.js gzip: %q %d vs %d", zipped.Header().Get("Content-Encoding"), zipped.Body.Len(), plain.Body.Len())
	}
	zr, _ := gzip.NewReader(zipped.Body)
	un, _ := io.ReadAll(zr)
	if string(un) != plain.Body.String() || !strings.Contains(zipped.Header().Get("Content-Type"), "javascript") {
		t.Error("gzip body or type mismatch")
	}
	if rec := get("/", "Accept-Encoding", "gzip"); rec.Header().Get("Content-Encoding") != "gzip" {
		t.Error("index not gzipped")
	}
	if rec := get(refs[5][1], "Accept-Encoding", "gzip;q=0"); rec.Header().Get("Content-Encoding") != "" {
		t.Error("gzip sent despite q=0")
	}
	if rec := get("/api/health", "Accept-Encoding", "gzip"); rec.Header().Get("Content-Encoding") != "" {
		t.Error("api gzipped")
	}
}
