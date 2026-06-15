package model

import (
	"sort"
	"strings"
)

type TMDBImagesResponse struct {
	Logos []TMDBLogo `json:"logos"`
}

type TMDBLogo struct {
	FilePath    string  `json:"file_path"`
	Iso639_1    *string `json:"iso_639_1"`
	VoteAverage float64 `json:"vote_average"`
	FileType    string  `json:"file_type"`
}

type MediaURLs struct {
	TrailerURL string `json:"trailer_url"`
	LogoURL    string `json:"logo_url"`
}

func PickLogoURL(logos []TMDBLogo, language, imageBase string) string {
	candidates := make([]TMDBLogo, 0, len(logos))
	for _, logo := range logos {
		if strings.TrimSpace(logo.FilePath) != "" {
			candidates = append(candidates, logo)
		}
	}
	if len(candidates) == 0 {
		return ""
	}

	langPrefix := languagePrefix(language)
	sort.SliceStable(candidates, func(i, j int) bool {
		leftRank := logoLanguageRank(candidates[i].Iso639_1, langPrefix)
		rightRank := logoLanguageRank(candidates[j].Iso639_1, langPrefix)
		if leftRank != rightRank {
			return leftRank < rightRank
		}
		return candidates[i].VoteAverage > candidates[j].VoteAverage
	})

	chosen := candidates[0]
	size := "w185"
	if isSVGLogo(chosen) {
		size = "original"
	}
	return buildImageURL(imageBase, size, chosen.FilePath)
}

func languagePrefix(language string) string {
	language = strings.TrimSpace(language)
	if language == "" {
		return "en"
	}
	if idx := strings.Index(language, "-"); idx > 0 {
		return strings.ToLower(language[:idx])
	}
	return strings.ToLower(language)
}

func isSVGLogo(logo TMDBLogo) bool {
	if strings.EqualFold(strings.TrimSpace(logo.FileType), "svg") {
		return true
	}
	return strings.HasSuffix(strings.ToLower(strings.TrimSpace(logo.FilePath)), ".svg")
}

func logoLanguageRank(iso639 *string, langPrefix string) int {
	if iso639 == nil || strings.TrimSpace(*iso639) == "" {
		return 1
	}
	if strings.EqualFold(strings.TrimSpace(*iso639), langPrefix) {
		return 0
	}
	return 2
}
