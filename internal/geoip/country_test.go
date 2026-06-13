package geoip

import "testing"

func TestIsValidCountryCode(t *testing.T) {
	tests := []struct {
		code  string
		valid bool
	}{
		{"CN", true},
		{"US", true},
		{"GB", true},
		{"XX", false},
		{"T1", false},
		{"", false},
		{"CHN", false},
		{"1A", false},
	}

	for _, tt := range tests {
		if got := IsValidCountryCode(tt.code); got != tt.valid {
			t.Errorf("IsValidCountryCode(%q) = %v, want %v", tt.code, got, tt.valid)
		}
	}
}

func TestNormalizeCountryCode(t *testing.T) {
	if got := NormalizeCountryCode(" cn "); got != "CN" {
		t.Fatalf("NormalizeCountryCode = %q, want CN", got)
	}
	if got := NormalizeCountryCode("XX"); got != "" {
		t.Fatalf("NormalizeCountryCode(XX) = %q, want empty", got)
	}
}
