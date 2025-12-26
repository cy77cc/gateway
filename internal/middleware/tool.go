package middleware

import (
	"strings"

	"github.com/cy77cc/gateway/config"
)

var currentRoutes []config.Route

func SetRoutes(rs []config.Route) {
	currentRoutes = rs
}

func matchRoute(path string) *config.Route {
	cfg := currentRoutes

	var matched *config.Route
	longest := 0

	for i := range cfg {
		r := &cfg[i]

		if strings.HasPrefix(path, r.PathPrefix) {
			if len(r.PathPrefix) > longest {
				longest = len(r.PathPrefix)
				matched = r
			}
		}
	}

	return matched
}
