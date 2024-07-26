package constants

import (
	"sync"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type LastReadPositionConfig struct {
	FileLastReadPosition  int64
	MongoLastReadPosition primitive.Timestamp
	mutex                 sync.RWMutex
}

var LastReadPosition LastReadPositionConfig

func (c *LastReadPositionConfig) Get() int64 {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.FileLastReadPosition
}

func (c *LastReadPositionConfig) Set(val int64) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.FileLastReadPosition = val
}

func (c *LastReadPositionConfig) GetMongo() primitive.Timestamp {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.MongoLastReadPosition
}

func (c *LastReadPositionConfig) SetMongo(val primitive.Timestamp) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.MongoLastReadPosition = val
}
