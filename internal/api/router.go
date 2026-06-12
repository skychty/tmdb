package api

import (
	"github.com/gin-gonic/gin"

	"tmdb/internal/api/handler"
	"tmdb/internal/api/middleware"
	"tmdb/internal/geoip"
	"tmdb/internal/service"
)

func NewRouter(movieService *service.MovieService, geoIP *geoip.Resolver) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery(), middleware.CORS(), middleware.RequestLog())

	movieHandler := handler.NewMovieHandler(movieService, geoIP)

	router.GET("/", handler.Index)
	router.GET("/health", handler.Health)

	v1 := router.Group("/api/v1")
	{
		v1.GET("/movies/latest", movieHandler.GetLatest)
		v1.GET("/movies/popular", movieHandler.GetPopular)
	}

	return router
}
