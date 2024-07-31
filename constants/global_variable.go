package constants

import (
	"encoding/gob"
	"os"
	"sync"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type LastReadCheckpointConfigWrapper struct {
	checkpoint LastReadEncodeCheckpoint
	mutex      sync.RWMutex
}

type LastReadEncodeCheckpoint struct {
	FileLastReadPosition  int64
	MongoLastReadPosition map[string]primitive.Timestamp
}

var LastReadCheckpoint = LastReadCheckpointConfigWrapper{
	checkpoint: LastReadEncodeCheckpoint{
		FileLastReadPosition:  0,
		MongoLastReadPosition: make(map[string]primitive.Timestamp),
	},
}

func (c *LastReadCheckpointConfigWrapper) GetFileCheckpoint() int64 {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.checkpoint.FileLastReadPosition
}

func (c *LastReadCheckpointConfigWrapper) SetFileCheckpoint(val int64) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.checkpoint.FileLastReadPosition = val
}

func (c *LastReadCheckpointConfigWrapper) GetMongoCheckpoint(key string) primitive.Timestamp {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.checkpoint.MongoLastReadPosition[key]
}

func (c *LastReadCheckpointConfigWrapper) SetMongoCheckpoint(key string, val primitive.Timestamp) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.checkpoint.MongoLastReadPosition[key] = val
}

func (c *LastReadCheckpointConfigWrapper) EncodeToGob(file *os.File) error {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return gob.NewEncoder(file).Encode(c.checkpoint)
}

func (c *LastReadCheckpointConfigWrapper) DecodeFromGob(file *os.File) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return gob.NewDecoder(file).Decode(&c.checkpoint)
}
