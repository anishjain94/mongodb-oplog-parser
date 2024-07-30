package constants

import (
	"sync"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type LastReadCheckpointConfig struct {
	FileLastReadPosition  int64
	MongoLastReadPosition map[string]primitive.Timestamp
	Mutex                 sync.RWMutex
}

// TODO: use config manager to unify lock and unlock logic.
var LastReadCheckpoint = LastReadCheckpointConfig{
	MongoLastReadPosition: make(map[string]primitive.Timestamp),
	Mutex:                 sync.RWMutex{},
	FileLastReadPosition:  0,
}

// TODO: check locks and unlocks.
func (c *LastReadCheckpointConfig) GetFileCheckpoint() int64 {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	return c.FileLastReadPosition
}

func (c *LastReadCheckpointConfig) SetFileCheckpoint(val int64) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	c.FileLastReadPosition = val
}

func (c *LastReadCheckpointConfig) GetMongoCheckpoint(key string) primitive.Timestamp {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	return c.MongoLastReadPosition[key]
}

func (c *LastReadCheckpointConfig) SetMongoCheckpoint(key string, val primitive.Timestamp) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	c.MongoLastReadPosition[key] = val
}
