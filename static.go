package main

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"mime"
	"net/http"
	"path"
	"sort"
	"strings"
)

// Static files are served under a content fingerprint, /s/<hash>/app.js, so
// browsers, nginx and Cloudflare can cache them forever while every release
// gets fresh URLs. index.html itself is always revalidated (ETag = hash).

type staticSite struct {
	files fs.FS
	hash  string
	index []byte
	gz    map[string][]byte // pre-compressed text files, "index.html" holds the rewritten page
}

func gzipBytes(b []byte) []byte {
	var buf bytes.Buffer
	zw, _ := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	zw.Write(b)
	zw.Close()
	return buf.Bytes()
}

func compressible(name string) bool {
	switch path.Ext(name) {
	case ".html", ".js", ".css", ".svg", ".json", ".txt":
		return true
	}
	return false
}

func acceptsGzip(r *http.Request) bool {
	for _, part := range strings.Split(r.Header.Get("Accept-Encoding"), ",") {
		enc, q, _ := strings.Cut(strings.TrimSpace(part), ";")
		if strings.EqualFold(strings.TrimSpace(enc), "gzip") && strings.TrimSpace(q) != "q=0" {
			return true
		}
	}
	return false
}

// writeGzip sends the pre-compressed copy when the client takes gzip.
func (s *staticSite) writeGzip(w http.ResponseWriter, r *http.Request, name string) bool {
	w.Header().Add("Vary", "Accept-Encoding")
	gz, ok := s.gz[name]
	if !ok || !acceptsGzip(r) {
		return false
	}
	w.Header().Set("Content-Type", mime.TypeByExtension(path.Ext(name)))
	w.Header().Set("Content-Encoding", "gzip")
	if r.Method != http.MethodHead {
		w.Write(gz)
	}
	return true
}

func newStaticSite(files fs.FS) (*staticSite, error) {
	h := sha256.New()
	var names []string
	fs.WalkDir(files, ".", func(p string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			names = append(names, p)
		}
		return nil
	})
	sort.Strings(names)
	for _, n := range names {
		b, err := fs.ReadFile(files, n)
		if err != nil {
			return nil, err
		}
		h.Write([]byte(n))
		h.Write(b)
	}
	site := &staticSite{files: files, hash: hex.EncodeToString(h.Sum(nil))[:12], gz: map[string][]byte{}}

	index, err := fs.ReadFile(files, "index.html")
	if err != nil {
		return nil, err
	}
	prefix := "/s/" + site.hash + "/"
	html := string(index)
	for _, n := range names {
		if n == "index.html" {
			continue
		}
		for _, attr := range []string{`href="`, `src="`} {
			html = strings.ReplaceAll(html, attr+n+`"`, attr+prefix+n+`"`)
		}
	}
	site.index = []byte(html)
	for _, n := range names {
		if !compressible(n) {
			continue
		}
		b := site.index
		if n != "index.html" {
			b, _ = fs.ReadFile(files, n)
		}
		if len(b) > 1024 {
			site.gz[n] = gzipBytes(b)
		}
	}
	return site, nil
}

func (s *staticSite) serveIndex(w http.ResponseWriter, r *http.Request) {
	etag := `"` + s.hash + `"`
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("ETag", etag)
	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	if s.writeGzip(w, r, "index.html") {
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(s.index)
}

// serveAsset handles /s/{hash}/{file...}. A stale hash (a page from before an
// upgrade) still gets the current file, but without long-term caching.
func (s *staticSite) serveAsset(w http.ResponseWriter, r *http.Request) {
	name := path.Clean(r.PathValue("file"))
	if name == "index.html" || strings.HasPrefix(name, "..") {
		http.NotFound(w, r)
		return
	}
	if r.PathValue("hash") == s.hash {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	} else {
		w.Header().Set("Cache-Control", "no-cache")
	}
	if s.writeGzip(w, r, name) {
		return
	}
	http.ServeFileFS(w, r, s.files, name)
}

// noStore keeps proxies and browsers from caching API responses.
func noStore(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			w.Header().Set("Cache-Control", "no-store")
		}
		h.ServeHTTP(w, r)
	})
}
