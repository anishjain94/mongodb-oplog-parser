package constants

import (
	"sync"
)

type ConfigManager[T any] struct {
	mutex sync.RWMutex
	Data  map[string]T
}

func NewConfigManager[T any]() *ConfigManager[T] {
	return &ConfigManager[T]{
		Data: make(map[string]T),
	}
}

func (c *ConfigManager[T]) Get(key string) (T, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	val, ok := c.Data[key]
	return val, ok
}

func (c *ConfigManager[T]) Set(key string, value T) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.Data[key] = value
}

type GlobalConfigs struct {
	CreateSchemaQuery *ConfigManager[bool]
	CreateTableQuery  *ConfigManager[bool]
	TableColumnName   *ConfigManager[[]string]
}

var (
	globalDataStore *GlobalConfigs
	once            sync.Once
)

// GetGlobalVariables returns the singleton instance of GlobalConfigs
func GetGlobalVariables() *GlobalConfigs {
	once.Do(func() {
		globalDataStore = &GlobalConfigs{
			CreateSchemaQuery: NewConfigManager[bool](),
			CreateTableQuery:  NewConfigManager[bool](),
			TableColumnName:   NewConfigManager[[]string](),
		}
	})
	return globalDataStore
}
