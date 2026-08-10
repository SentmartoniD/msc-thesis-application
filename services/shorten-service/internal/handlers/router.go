package handlers

import (
	"shorten-service/internal/server/middlewares"

	"github.com/gin-gonic/gin"
)

// SetupRouter builds the public engine: the redirect hot path and nothing else.
func SetupRouter() *gin.Engine {
	router := newEngine()

	// Phase 1 — the measured endpoint:
	// rh := NewRedirectHandler(links.NewRepository(database.Pool))
	// router.GET("/:code", rh.Redirect)

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
