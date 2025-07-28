package config

import (
	"IsaacCoyote/common/config/model"
	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"os"
	"sync"
)

type Manager struct {
	configFile     string
	config         *model.ConfigRoot
	reloadHandlers []func(*Manager) error

	watcher       *fsnotify.Watcher
	configLock    sync.RWMutex
	isInitialized bool
}

func (m *Manager) Init() error {
	if m.isInitialized {
		return nil
	}
	go m.watchConfig()

	err := m.reloadConfig()
	if err != nil {
		return err
	}

	m.isInitialized = true

	return nil
}

func (m *Manager) RegReloadHandler(handler func(*Manager) error) {
	m.reloadHandlers = append(m.reloadHandlers, handler)
}

func (m *Manager) GetConfig() *model.ConfigRoot {
	m.configLock.RLock()
	defer m.configLock.RUnlock()

	return m.config
}

func (m *Manager) watchConfig() {
	err := m.watcher.Add(m.configFile)
	if err != nil {
		return
	}
	for event := range m.watcher.Events {
		if event.Op&fsnotify.Write == fsnotify.Write {
			err = m.reloadConfig()
			if err != nil {
				zap.L().Error("重载配置失败", zap.Error(err))
				continue
			}
			zap.L().Info("配置已更新")
		}
	}
}

func (m *Manager) reloadConfig() error {
	data, err := os.ReadFile(m.configFile)
	if err != nil {
		return err
	}

	m.configLock.Lock()
	err = yaml.Unmarshal(data, &m.config)
	if err != nil {
		return err
	}
	m.configLock.Unlock()

	for _, handler := range m.reloadHandlers {
		err = handler(m)
		if err != nil {
			return err
		}
	}

	return nil
}

func NewConfigManager(configFile string) (*Manager, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &Manager{
		watcher:        watcher,
		configFile:     configFile,
		reloadHandlers: make([]func(*Manager) error, 0),
	}, nil
}
