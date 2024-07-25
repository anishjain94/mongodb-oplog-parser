package constants

import (
	"sync"
)

type ConfigManager[T any] struct {
	mutex sync.RWMutex
	data  map[string]T
}

func NewConfigManager[T any]() *ConfigManager[T] {
	return &ConfigManager[T]{
		data: make(map[string]T),
	}
}

func (c *ConfigManager[T]) Get(key string) (T, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	val, ok := c.data[key]
	return val, ok
}

func (c *ConfigManager[T]) Set(key string, value T) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.data[key] = value
}

// TODO: keep data persistant across program restart.
// TODO: persistance.
// GlobalConfigs holds all your different configurations
type GlobalConfigs struct {
	CreateSchemaQuery *ConfigManager[bool]
	CreateTableQuery  *ConfigManager[bool]
	TableColumnName   *ConfigManager[[]string]
}

var (
	globalConfigs *GlobalConfigs
	once          sync.Once
)

// GetGlobalVariables returns the singleton instance of GlobalConfigs
func GetGlobalVariables() *GlobalConfigs {
	once.Do(func() {
		globalConfigs = &GlobalConfigs{
			CreateSchemaQuery: NewConfigManager[bool](),
			CreateTableQuery:  NewConfigManager[bool](),
			TableColumnName:   NewConfigManager[[]string](),
		}
	})
	return globalConfigs
}