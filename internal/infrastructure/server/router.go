package server

import (
	apierrors "catalog-service/internal/errors"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

type Routes interface {
	RegisterRoutes(*gin.RouterGroup)
}

func NewRouter(routes ...Routes) *gin.Engine {
	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.CustomRecoveryWithWriter(nil, func(c *gin.Context, recovered any) {
		slog.Error(
			"panic recovered",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"panic", recovered,
			"stack", string(debug.Stack()),
		)
		c.AbortWithStatusJSON(http.StatusInternalServerError, apierrors.Internal())
	}))

	router.HandleMethodNotAllowed = true
	router.NoRoute(func(c *gin.Context) {
		apierrors.Write(c, apierrors.NotFound("ROUTE_NOT_FOUND", "route not found"), nil)
	})
	router.NoMethod(func(c *gin.Context) {
		apierrors.Write(c, apierrors.MethodNotAllowed(), nil)
	})

	api := router.Group("/api/v1")

	for _, route := range routes {
		route.RegisterRoutes(api)
	}

	return router
}
