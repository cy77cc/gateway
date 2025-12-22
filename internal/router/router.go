package router

import (
	"github.com/cy77cc/gateway/config"
	"github.com/cy77cc/gateway/internal/proxy"
	"github.com/gin-gonic/gin"
)

type Router struct {
}

func NewRouter() *Router {
	return &Router{}
}

func (*Router) RegisterRoutes(r *gin.Engine, routes []config.Route, proxyHandler *proxy.Handler) {
	for _, route := range routes {
		r.Any(route.PathPrefix+"/*path", proxyHandler.HandleRoute(route.Service, route.StripPrefix))
	}

	// Default/Fallback Route
	r.Any("/api/:service/*path", proxyHandler.HandleGeneric)
}
