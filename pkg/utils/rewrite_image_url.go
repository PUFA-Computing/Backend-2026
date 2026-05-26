package utils

import (
	"Backend/configs"
	"strings"
)

// legacyImageHosts are the custom domains that previously fronted the R2/S3
// buckets but are now unreachable. Any URL pointing to one of these hosts
// must be rewritten to the currently-configured public base URL so that
// images persisted in the database keep resolving.
var legacyImageHosts = []string{
	"https://pufacompsci.my.id",
	"https://id.pufacomputing.live",
}

// RewriteImageURL takes a URL stored in the database (possibly using one of
// the legacy custom domains) and returns a URL that points at the active
// public host configured via R2_PUBLIC_URL / S3_PUBLIC_URL. Empty input is
// returned unchanged. Inputs that do not match a legacy host are returned
// unchanged so this helper can be applied unconditionally.
func RewriteImageURL(rawURL string) string {
	if rawURL == "" {
		return rawURL
	}

	cfg := configs.LoadConfig()
	target := strings.TrimRight(cfg.R2PublicURL, "/")
	if target == "" {
		target = strings.TrimRight(cfg.S3PublicURL, "/")
	}
	if target == "" && cfg.CloudflareAccountId != "" && cfg.S3Bucket != "" {
		target = "https://" + cfg.CloudflareAccountId + ".r2.cloudflarestorage.com/" + cfg.S3Bucket
	}
	if target == "" {
		return rawURL
	}

	for _, host := range legacyImageHosts {
		if strings.HasPrefix(rawURL, host) {
			return target + strings.TrimPrefix(rawURL, host)
		}
	}
	return rawURL
}
