package service

import "testing"

func TestNormalizePageLimit(t *testing.T) {
	tests := []struct {
		in   int
		want int
	}{
		{0, MaxPageSize},
		{-1, MaxPageSize},
		{1, 1},
		{10, 10},
		{20, 20},
		{25, MaxPageSize},
	}

	for _, tt := range tests {
		if got := normalizePageLimit(tt.in); got != tt.want {
			t.Fatalf("normalizePageLimit(%d) = %d, want %d", tt.in, got, tt.want)
		}
	}
}
