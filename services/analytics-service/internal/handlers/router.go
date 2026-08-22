package handlers

import (
	"analytics-service/internal/server/middlewares"

	"github.com/gin-gonic/gin"
)

// SetupRouter builds the public engine: the redirect hot path and nothing else.
func SetupRouter(clickService ClickService) *gin.Engine {
	router := newEngine()

	analyticsHandler := NewAnalyticsHandler(clickService)

	apiRouter := router.Group("api/v1")

	// analytics router
	analyticsRouter := apiRouter.Group("analytics")
	analyticsRouter.GET(":code", analyticsHandler.GetCodeAnalyticsHandler)

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
