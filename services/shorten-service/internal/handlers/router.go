package handlers

import (
	"shorten-service/internal/config"
	"shorten-service/internal/repositories"
	"shorten-service/internal/server/middlewares"
	"shorten-service/internal/services"
	"shorten-service/platform/database"

	"github.com/gin-gonic/gin"
)

// SetupRouter builds the public gin api engine.
func SetupRouter(cfg *config.Config) *gin.Engine {
	router := newEngine()

	linkRepository := repositories.NewLinkRepository(database.Pool)
	linkService := services.NewLinkService(linkRepository, cfg)
	linkHandler := NewLinkHandler(linkService)

	apiRouter := router.Group("api/v1")

	// link router
	linkRouter := apiRouter.Group("link")
	linkRouter.POST("", linkHandler.CreateLinkHandler)
	linkRouter.GET(":code", linkHandler.GetLinkHandler)

	return router
}

func SetupAdminRouter(hh *HealthHandler) *gin.Engine {
	router := newEngine()

	router.GET("/healthz", hh.Live)
	router.GET("/readyz", hh.Ready)

	return router
}

// newEngine returns a gin engine with the middlewares every server shares
func newEngine() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(middlewares.Recovery())

	return router
}
