package middleware

import (
	"strings"

	"github.com/cy77cc/gateway/config"
)

func matchRoute(path string) *config.Route {
	cfg := config.NewConfigManager().GetConfig().Routes

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
