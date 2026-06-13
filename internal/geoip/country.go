package geoip

import "strings"

var invalidCountryCodes = map[string]struct{}{
	"XX": {}, // Cloudflare: unknown
	"T1": {}, // Cloudflare: Tor
}

func IsValidCountryCode(code string) bool {
	code = strings.ToUpper(strings.TrimSpace(code))
	if len(code) != 2 {
		return false
	}
	for _, ch := range code {
		if ch < 'A' || ch > 'Z' {
			return false
		}
	}
	if _, invalid := invalidCountryCodes[code]; invalid {
		return false
	}
	return true
}

func NormalizeCountryCode(code string) string {
	code = strings.ToUpper(strings.TrimSpace(code))
	if !IsValidCountryCode(code) {
		return ""
	}
	return code
}
