package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"tmdb/internal/geoip"
	"tmdb/internal/model"
	"tmdb/internal/service"
)

type TVHandler struct {
	service *service.TVService
	geoIP   *geoip.Resolver
}

func NewTVHandler(service *service.TVService, geoIP *geoip.Resolver) *TVHandler {
	return &TVHandler{
		service: service,
		geoIP:   geoIP,
	}
}

func (h *TVHandler) GetOnTheAir(c *gin.Context) {
	h.handleList(c, h.service.GetOnTheAirTV)
}

func (h *TVHandler) GetPopular(c *gin.Context) {
	h.handleList(c, h.service.GetPopularTV)
}

func (h *TVHandler) GetRegionalPopular(c *gin.Context) {
	h.handleList(c, h.service.GetRegionalPopularTV)
}

type tvListFunc func(ctx context.Context, region, language string, page, limit int) (model.TVListResponse, error)

func (h *TVHandler) handleList(c *gin.Context, fn tvListFunc) {
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
