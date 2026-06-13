package handler

import "testing"

func TestPickClientIP(t *testing.T) {
	tests := []struct {
		name     string
		xRealIP  string
		xff      string
		expected string
	}{
		{
			name:     "google proxy in front of nginx",
			xRealIP:  "34.160.111.145",
			xff:      "154.81.14.124, 34.160.111.145",
			expected: "154.81.14.124",
		},
		{
			name:     "direct client via nginx",
			xRealIP:  "154.81.14.124",
			xff:      "154.81.14.124",
			expected: "154.81.14.124",
		},
		{
			name:     "only x-real-ip",
			xRealIP:  "154.81.14.124",
			xff:      "",
			expected: "154.81.14.124",
		},
		{
			name:     "private x-real-ip ignored",
			xRealIP:  "127.0.0.1",
			xff:      "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pickClientIP(tt.xRealIP, tt.xff); got != tt.expected {
				t.Fatalf("pickClientIP() = %q, want %q", got, tt.expected)
			}
		})
	}
}
