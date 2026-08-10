package handlers

import (
	"context"
	"errors"
	"net/http"
	"shorten-service/pkg/logger"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const readinessPingTimeout = 2 * time.Second

// PingFunc reports whether the database is reachable.
type PingFunc func(ctx context.Context) error

type HealthHandler struct {
	ping     PingFunc
	draining atomic.Bool
}

func NewHealthHandler(ping PingFunc) *HealthHandler {
	return &HealthHandler{ping: ping}
}

func (h *HealthHandler) Drain() {
	h.draining.Store(true)
}

func (h *HealthHandler) Live(c *gin.Context) {
	c.String(http.StatusOK, "ok")
}

func (h *HealthHandler) Ready(c *gin.Context) {
	if h.draining.Load() {
		c.String(http.StatusServiceUnavailable, "draining")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), readinessPingTimeout)
	defer cancel()

	switch err := h.ping(ctx); {
	case err == nil:
		c.String(http.StatusOK, "ready")
	case errors.Is(err, context.DeadlineExceeded):
		logger.Log.Warn("readiness ping timed out, reporting ready anyway")
		c.String(http.StatusOK, "degraded")
	default:
		logger.Log.Error("readiness ping failed", zap.Error(err))
		c.String(http.StatusServiceUnavailable, "database unreachable")
	}
}
