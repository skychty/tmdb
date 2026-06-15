package service

const MaxPageSize = 20

func normalizePageLimit(limit int) int {
	if limit < 1 {
		return MaxPageSize
	}
	if limit > MaxPageSize {
		return MaxPageSize
	}
	return limit
}
