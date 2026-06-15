package model

import (
	"fmt"
	"sort"
	"strings"
)

type TMDBVideosResponse struct {
	ID      int         `json:"id"`
	Results []TMDBVideo `json:"results"`
}

type TMDBVideo struct {
	Key      string `json:"key"`
	Name     string `json:"name"`
	Site     string `json:"site"`
	Type     string `json:"type"`
	Official bool   `json:"official"`
	Size     int    `json:"size"`
}

func PickTrailerURL(videos []TMDBVideo) string {
	candidates := filterYouTubeVideos(videos, true)
	if len(candidates) == 0 {
		candidates = filterYouTubeVideos(videos, false)
	}
	if len(candidates) == 0 {
		return ""
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		left, right := candidates[i], candidates[j]
		if left.Official != right.Official {
			return left.Official
		}
		leftTypeRank := videoTypeRank(left.Type)
		rightTypeRank := videoTypeRank(right.Type)
		if leftTypeRank != rightTypeRank {
			return leftTypeRank < rightTypeRank
		}
		return left.Size > right.Size
	})

	return fmt.Sprintf("https://www.youtube.com/watch?v=%s", candidates[0].Key)
}

func filterYouTubeVideos(videos []TMDBVideo, trailerOnly bool) []TMDBVideo {
	result := make([]TMDBVideo, 0, len(videos))
	for _, video := range videos {
		if !strings.EqualFold(video.Site, "YouTube") || video.Key == "" {
			continue
		}
		if trailerOnly && videoTypeRank(video.Type) > 1 {
			continue
		}
		result = append(result, video)
	}
	return result
}

func videoTypeRank(videoType string) int {
	switch strings.ToLower(strings.TrimSpace(videoType)) {
	case "trailer":
		return 0
	case "teaser":
		return 1
	default:
		return 2
	}
}
