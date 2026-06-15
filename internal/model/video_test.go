package model

import "testing"

func TestPickTrailerURL(t *testing.T) {
	url := PickTrailerURL([]TMDBVideo{
		{Key: "teaser1", Site: "YouTube", Type: "Teaser", Official: true, Size: 1080},
		{Key: "trailer1", Site: "YouTube", Type: "Trailer", Official: false, Size: 720},
		{Key: "official", Site: "YouTube", Type: "Trailer", Official: true, Size: 1080},
	})
	if url != "https://www.youtube.com/watch?v=official" {
		t.Fatalf("PickTrailerURL() = %q, want official trailer", url)
	}
}

func TestPickTrailerURLEmpty(t *testing.T) {
	if got := PickTrailerURL(nil); got != "" {
		t.Fatalf("PickTrailerURL(nil) = %q, want empty", got)
	}
}
