package constants

import (
	"encoding/gob"
	"os"
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

func (config *GlobalConfigs) OverrideConfig() {
	globalConfigs = config
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

func StoreCheckpoint() error {
	globalConfig := GetGlobalVariables()

	gobFile, err := os.Create("checkpoint.gob")
	if err != nil {
		return err
	}

	err = gob.NewEncoder(gobFile).Encode(globalConfig)
	if err != nil {
		return err
	}

	err = gobFile.Close()
	if err != nil {
		return err
	}

	return nil
}

func RestoreCheckpoint() error {
	var globalConfig GlobalConfigs

	gobFile, err := os.OpenFile("checkpoint.gob", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer gobFile.Close()

	fileInfo, err := gobFile.Stat()
	if err != nil {
		return err
	}

	if fileInfo.Size() != 0 {
		err = gob.NewDecoder(gobFile).Decode(&globalConfig)
		if err != nil {
			return err
		}
		globalConfig.OverrideConfig()
	}

	return nil
}
