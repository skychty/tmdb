package model

import (
	"strings"
	"testing"
)

func strPtr(s string) *string { return &s }

func TestPickLogoURLPrefersLanguage(t *testing.T) {
	url := PickLogoURL([]TMDBLogo{
		{FilePath: "/en.png", Iso639_1: strPtr("en"), VoteAverage: 1},
		{FilePath: "/zh.png", Iso639_1: strPtr("zh"), VoteAverage: 0.5},
	}, "zh-CN", "https://image.tmdb.org/t/p")

	if !strings.Contains(url, "/zh.png") {
		t.Fatalf("PickLogoURL() = %q, want zh logo", url)
	}
}

func TestPickLogoURLUsesNeutralFallback(t *testing.T) {
	url := PickLogoURL([]TMDBLogo{
		{FilePath: "/neutral.png", Iso639_1: nil, VoteAverage: 2},
		{FilePath: "/ja.png", Iso639_1: strPtr("ja"), VoteAverage: 5},
	}, "zh-CN", "https://image.tmdb.org/t/p")

	if !strings.Contains(url, "/neutral.png") {
		t.Fatalf("PickLogoURL() = %q, want neutral logo", url)
	}
}
