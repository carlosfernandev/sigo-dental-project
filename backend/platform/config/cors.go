package config

import (
	"net/url"
	"strings"
)

const defaultLocalCORSOrigins = "http://localhost:5173,http://127.0.0.1:5173"

// CORS contains the browser origins allowed to call the API.
type CORS struct {
	AllowedOrigins []string
}

func loadCORS(l *loader, environment string) CORS {
	rawOrigins := l.optional("SIGO_CORS_ALLOWED_ORIGINS")
	if rawOrigins == "" {
		if environment == EnvProduction || environment == EnvStaging {
			l.add("SIGO_CORS_ALLOWED_ORIGINS is required when SIGO_ENV=%s", environment)
			return CORS{}
		}
		rawOrigins = defaultLocalCORSOrigins
	}

	origins := make([]string, 0)
	seen := make(map[string]struct{})
	for _, rawOrigin := range strings.Split(rawOrigins, ",") {
		origin := strings.TrimSpace(rawOrigin)
		parsed, err := url.Parse(origin)
		if origin == "" || err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") ||
			parsed.Host == "" || strings.Contains(parsed.Host, "*") || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" ||
			(parsed.Path != "" && parsed.Path != "/") {
			l.add("SIGO_CORS_ALLOWED_ORIGINS contains an invalid origin")
			continue
		}
		origin = strings.TrimSuffix(origin, "/")
		if _, exists := seen[origin]; exists {
			continue
		}
		seen[origin] = struct{}{}
		origins = append(origins, origin)
	}
	if len(origins) == 0 {
		l.add("SIGO_CORS_ALLOWED_ORIGINS must contain at least one valid origin")
	}
	return CORS{AllowedOrigins: origins}
}
