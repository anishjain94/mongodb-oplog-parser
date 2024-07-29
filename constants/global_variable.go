package constants

import (
	"sync"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type LastReadCheckpointConfig struct {
	FileLastReadPosition  int64
	MongoLastReadPosition primitive.Timestamp
	mutex                 sync.RWMutex
}

var LastReadCheckpoint LastReadCheckpointConfig

func (c *LastReadCheckpointConfig) GetFileCheckpoint() int64 {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.FileLastReadPosition
}

func (c *LastReadCheckpointConfig) SetFileCheckpoint(val int64) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.FileLastReadPosition = val
}

func (c *LastReadCheckpointConfig) GetMongoCheckpoint() primitive.Timestamp {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.MongoLastReadPosition
}

func (c *LastReadCheckpointConfig) SetMongoCheckpoint(val primitive.Timestamp) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.MongoLastReadPosition = val
}
