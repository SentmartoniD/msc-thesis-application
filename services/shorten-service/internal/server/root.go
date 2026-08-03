package server

import (
	"net"
	"net/http"
	"os"
	"shorten-service/pkg/logger"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var (
	ip         = net.IPv4(127, 0, 0, 1)
	port       = 8000
	httpServer *http.Server
)

func LoadServerConfig() {
	serverPort := os.Getenv("SERVER_PORT")
	if serverPort != "" {
		port, _ = strconv.Atoi(serverPort)
	}

	serverIP := os.Getenv("SERVER_IP")
	if serverIP != "" {
		parsedIP, _, err := net.ParseCIDR(serverIP + "/32")
		if err == nil {
			ip = parsedIP
		}
	}
}

func Start() {
	log := logger.Log.WithOptions(zap.Fields(
		zap.String("ip", ip.String()),
	))

	log.Debug("starting engine")

	// Sets default size only if not set
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// delete server part from header
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Del("Server")
		c.Next()
	})

	// router.Use(middlewares.NoCache())
	// router.Use(middlewares.Session())
	// router.Use(middlewares.CORS())
	//router.Use(middlewares.Security())

	//handlers.SetupRouter(router)
	httpServer = Initialize(ip, port, router)
	go func() {
		errorsStartingUp := 0
		var err error

		for errorsStartingUp < 5 {
			log.Info("attempting to start HTTP server",
				zap.Int("port", port),
				zap.Int("errorsStartingUp", errorsStartingUp),
			)

			httpServer = Initialize(ip, port, router)

			err = httpServer.ListenAndServe()
			if err != nil {
				log.Info("retrying to start HTTP server, because of an error",
					zap.Int("port", port),
					zap.Int("errorsStartingUp", errorsStartingUp),
					zap.Error(err),
				)

				port++
				errorsStartingUp++
				continue
			}
			break
		}

		log.Panic("failed to start HTTP server",
			zap.Int("port", port),
			zap.Int("errorsStartingUp", errorsStartingUp),
			zap.Error(err),
		)
	}()

	Wait(httpServer, log)
}
