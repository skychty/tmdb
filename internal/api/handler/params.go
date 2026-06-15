package handler

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"

	"tmdb/internal/service"
)

var errInvalidLimit = &regionError{msg: fmt.Sprintf("limit must be between 1 and %d", service.MaxPageSize)}

func parseListQuery(c *gin.Context) (language string, page int, limit int, err error) {
	language = c.DefaultQuery("language", "en-US")

	page, err = strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		return "", 0, 0, errInvalidPage
	}

	limitStr := c.DefaultQuery("limit", strconv.Itoa(service.MaxPageSize))
	limit, err = strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > service.MaxPageSize {
		return "", 0, 0, errInvalidLimit
	}

	return language, page, limit, nil
}

var errInvalidPage = &regionError{msg: "page must be a positive integer"}
