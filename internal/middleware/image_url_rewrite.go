package middleware

import (
	"Backend/configs"
	"bytes"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

// legacyImageHosts mirrors pkg/utils/rewrite_image_url.go. Duplicated here
// to avoid an import cycle (middleware should not pull in pkg/utils for a
// trivial constant slice).
var legacyImageHosts = []string{
	"https://pufacompsci.my.id",
	"https://id.pufacomputing.live",
	"https://pub-a1fe3a899fbf461ca62a44252df9125f.r2.dev",
}

// bareKeyImageFields matches JSON fields that hold image URLs whose values
// are bare R2 object keys (e.g. "events/foo.jpg") rather than full URLs.
// We rewrite these to absolute URLs using R2_PUBLIC_URL so the frontend can
// load them directly. The regex deliberately:
//   - matches only known image field names to avoid touching unrelated strings
//   - requires the value to start with a known R2 directory prefix
//     (events/, event/, projects/, news/, profile/, profiles/, twofa/, logos/)
//     so we never double-prefix legitimate non-image strings
//   - skips values that already start with http:// or https:// or data:
var bareKeyImageFieldRegex = regexp.MustCompile(
	`"(thumbnail|image_url|profile_picture|twofa_image|logo|banner)"\s*:\s*"((?:events?|projects|news|profile|profiles|twofa|logos)/[^"]+)"`,
)

type rewriteResponseWriter struct {
	gin.ResponseWriter
	body   *bytes.Buffer
	status int
}

func (w *rewriteResponseWriter) Write(b []byte) (int, error) {
	return w.body.Write(b)
}

func (w *rewriteResponseWriter) WriteString(s string) (int, error) {
	return w.body.WriteString(s)
}

// WriteHeader captures the status code instead of flushing it immediately,
// so we can rewrite the body before headers are committed.
func (w *rewriteResponseWriter) WriteHeader(code int) {
	w.status = code
}

// ImageURLRewrite swaps any legacy image hosts that appear anywhere in the
// JSON response body for the currently configured public base URL. It is a
// safety net for rows in the database that still reference the dead custom
// domains (pufacompsci.my.id, id.pufacomputing.live).
//
// MUST be registered BEFORE the gzip middleware so we operate on the raw,
// uncompressed body.
func ImageURLRewrite() gin.HandlerFunc {
	// Resolve target host once at construction time — env doesn't change
	// per request, and LoadConfig is expensive (reads ~30 env vars).
	cfg := configs.LoadConfig()
	target := strings.TrimRight(cfg.R2PublicURL, "/")
	if target == "" {
		target = strings.TrimRight(cfg.S3PublicURL, "/")
	}
	if target == "" && cfg.CloudflareAccountId != "" && cfg.S3Bucket != "" {
		target = "https://" + cfg.CloudflareAccountId + ".r2.cloudflarestorage.com/" + cfg.S3Bucket
	}

	// No target configured — return a passthrough that does nothing.
	if target == "" {
		return func(c *gin.Context) { c.Next() }
	}

	targetBytes := []byte(target)
	hostBytes := make([][]byte, len(legacyImageHosts))
	for i, h := range legacyImageHosts {
		hostBytes[i] = []byte(h)
	}

	return func(c *gin.Context) {
		original := c.Writer
		buf := &bytes.Buffer{}
		writer := &rewriteResponseWriter{
			ResponseWriter: original,
			body:           buf,
			status:         http.StatusOK,
		}
		c.Writer = writer

		c.Next()

		// Restore the real writer before flushing so downstream middleware
		// (e.g. gzip) sees a normal ResponseWriter again.
		c.Writer = original

		// Only rewrite JSON bodies — non-JSON responses (files, HTML,
		// streamed downloads) are passed through untouched to avoid
		// buffering large payloads in memory.
		contentType := writer.Header().Get("Content-Type")
		body := buf.Bytes()
		if strings.HasPrefix(contentType, "application/json") && len(body) > 0 {
			// 1) Rewrite legacy hosts to the current R2 public URL.
			for _, host := range hostBytes {
				body = bytes.ReplaceAll(body, host, targetBytes)
			}
			// 2) Rewrite bare R2 keys (e.g. "events/foo.jpg") on known
			// image fields to absolute URLs. This handles legacy rows
			// where only the object key (not a full URL) was stored.
			body = bareKeyImageFieldRegex.ReplaceAll(body, []byte(
				`"$1":"`+target+`/$2"`,
			))
		}

		// Drop the stale Content-Length set by c.JSON(); downstream
		// (gzip / chunked transfer) will compute the correct value.
		original.Header().Del("Content-Length")

		original.WriteHeader(writer.status)
		if len(body) > 0 {
			_, _ = original.Write(body)
		}
	}
}
