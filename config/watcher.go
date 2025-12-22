package config

// RouteConfigWatcher 路由配置观察者
type RouteConfigWatcher struct {
	manager *Manager
}

func (w *RouteConfigWatcher) SetConfigManager(manager *Manager) {
	w.manager = manager
}

func (w *RouteConfigWatcher) ApplyConfig(content string) error {
	routes, err := ParseRemoteRoutes(content)
	if err != nil {
		return err
	}

	remoteConfig := &RemoteConfig{
		Routes: routes,
		// 保持现有的Middleware配置
		Middleware: w.manager.GetConfig().Middleware,
	}
	w.manager.SetRemoteConfig(remoteConfig)
	return nil
}

// MiddlewareConfigWatcher 中间件配置观察者
type MiddlewareConfigWatcher struct {
	manager *Manager
}

func (w *MiddlewareConfigWatcher) ApplyConfig(content string) error {
	middleware, err := ParseRemoteMiddleware(content)
	if err != nil {
		return err
	}

	remoteConfig := &RemoteConfig{
		// 保持现有的Routes配置
		Routes:     w.manager.GetConfig().Routes,
		Middleware: *middleware,
	}
	w.manager.SetRemoteConfig(remoteConfig)
	return nil
}

func (w *MiddlewareConfigWatcher) SetConfigManager(manager *Manager) {
	w.manager = manager
}
