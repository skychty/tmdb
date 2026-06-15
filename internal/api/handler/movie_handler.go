package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"tmdb/internal/geoip"
	"tmdb/internal/model"
	"tmdb/internal/service"
)

type MovieHandler struct {
	service  *service.MovieService
	geoIP    *geoip.Resolver
}

func NewMovieHandler(service *service.MovieService, geoIP *geoip.Resolver) *MovieHandler {
	return &MovieHandler{
		service: service,
		geoIP:   geoIP,
	}
}

func (h *MovieHandler) GetLatest(c *gin.Context) {
	h.handleList(c, h.service.GetLatestMovies)
}

func (h *MovieHandler) GetPopular(c *gin.Context) {
	h.handleList(c, h.service.GetPopularMovies)
}

func (h *MovieHandler) GetRegionalPopular(c *gin.Context) {
	h.handleList(c, h.service.GetRegionalPopularMovies)
}

type listFunc func(ctx context.Context, region, language string, page, limit int) (model.MovieListResponse, error)

func (h *MovieHandler) handleList(c *gin.Context, fn listFunc) {
	regionResult, err := resolveRegion(c, h.geoIP)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	language, page, limit, err := parseListQuery(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := fn(c.Request.Context(), regionResult.region, language, page, limit)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	applyRegionHeaders(c, regionResult)
	c.JSON(http.StatusOK, resp)
}

var errInvalidRegion = &regionError{msg: "region must be a 2-letter ISO 3166-1 code"}

type regionError struct {
	msg string
}

func (e *regionError) Error() string {
	return e.msg
}

func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
