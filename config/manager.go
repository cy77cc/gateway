package config

import (
	"sync"
	"sync/atomic"
)

// Watcher 配置观察者接口
type Watcher interface {
	OnConfigChange(config *MergedConfig)
}

// Manager 配置管理器
type Manager struct {
	localConfig  *LocalConfig
	remoteConfig *RemoteConfig
	mergedConfig atomic.Value // *MergedConfig
	watchers     []Watcher
	mu           sync.RWMutex
}

// NewConfigManager 创建配置管理器实例
func NewConfigManager() *Manager {
	cm := &Manager{
		localConfig:  &LocalConfig{},
		remoteConfig: &RemoteConfig{},
		watchers:     make([]Watcher, 0),
	}

	// 初始化空配置
	cm.mergedConfig.Store(&MergedConfig{})
	return cm
}

// SetLocalConfig 设置本地配置
func (cm *Manager) SetLocalConfig(config *LocalConfig) {
	cm.mu.Lock()

	cm.localConfig = config
	cm.mu.Unlock()
	cm.mergeConfig()
}

// SetRemoteConfig 设置远程配置
func (cm *Manager) SetRemoteConfig(config *RemoteConfig) {
	cm.mu.Lock()

	cm.remoteConfig = config
	cm.mu.Unlock()
	cm.mergeConfig()
}

// GetConfig 获取当前合并后的配置
func (cm *Manager) GetConfig() *MergedConfig {
	return cm.mergedConfig.Load().(*MergedConfig)
}

// RegisterWatcher 注册配置观察者
func (cm *Manager) RegisterWatcher(watcher Watcher) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.watchers = append(cm.watchers, watcher)
}

// mergeConfig 合并本地和远程配置
func (cm *Manager) mergeConfig() {
	merged := &MergedConfig{
		Server:     cm.localConfig.Server,
		Proxy:      cm.localConfig.Proxy,
		Logging:    cm.localConfig.Logging,
		Routes:     cm.remoteConfig.Routes,
		Middleware: cm.remoteConfig.Middleware,
	}

	cm.mergedConfig.Store(merged)
	cm.notifyWatchers(merged)
}

// notifyWatchers 通知所有观察者配置已变更
func (cm *Manager) notifyWatchers(config *MergedConfig) {
	// 创建副本避免锁竞争
	cm.mu.RLock()
	watchers := make([]Watcher, len(cm.watchers))
	copy(watchers, cm.watchers)
	cm.mu.RUnlock()

	for _, watcher := range watchers {
		watcher.OnConfigChange(config)
	}
}
