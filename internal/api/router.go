package api

import (
	"github.com/gin-gonic/gin"

	"tmdb/internal/api/handler"
	"tmdb/internal/api/middleware"
	"tmdb/internal/geoip"
	"tmdb/internal/service"
)

func NewRouter(movieService *service.MovieService, tvService *service.TVService, geoIP *geoip.Resolver) *gin.Engine {
	router := gin.New()
	_ = router.SetTrustedProxies([]string{"127.0.0.1", "::1"})
	router.Use(gin.Recovery(), middleware.CORS(), middleware.RequestLog())

	movieHandler := handler.NewMovieHandler(movieService, geoIP)
	tvHandler := handler.NewTVHandler(tvService, geoIP)

	router.GET("/", handler.Index)
	router.GET("/health", handler.Health)

	v1 := router.Group("/api/v1")
	{
		v1.GET("/movies/latest", movieHandler.GetLatest)
		v1.GET("/movies/popular", movieHandler.GetPopular)
		v1.GET("/movies/regional-popular", movieHandler.GetRegionalPopular)

		v1.GET("/tv/on-the-air", tvHandler.GetOnTheAir)
		v1.GET("/tv/popular", tvHandler.GetPopular)
		v1.GET("/tv/regional-popular", tvHandler.GetRegionalPopular)
	}

	return router
}
