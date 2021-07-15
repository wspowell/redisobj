package redisobj

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/go-redis/redis/v7"
)

const (
	rootKeyPrefix = "redisobj"
)

type Option struct {
	name  string
	value interface{}
}

func newOption(name string, value interface{}) Option {
	return Option{
		name:  name,
		value: value,
	}
}

var (
	OptionIfExists    = newOption("xx", nil)
	OptionIfNotExists = newOption("nx", nil)
	OptionTtl         = func(ttl time.Duration) Option { return newOption("ttl", ttl) }

	// TODO: Objects could have their values hashed and stored in a special key in redis.
	//       This could provide a quick way to check if an object is different than the one in memory without having to pull the entire object.
	OptionCache = newOption("cache", nil)
)

type Writer interface {
	Write(obj interface{}) error
}

type Reader interface {
	Read(obj interface{}) error
}

type Store struct {
	redisClient *redis.Client // FIXME: This is forced to be either Client or ClusterClient which is really annoying.
	mutex       *sync.RWMutex
	objTypes    map[string]*objStruct // FIXME: Need to sync this map
}

func NewStore(redisClient *redis.Client) *Store {
	return &Store{
		redisClient: redisClient,
		mutex:       &sync.RWMutex{},
		objTypes:    map[string]*objStruct{},
	}
}

func (self *Store) getObjectStruct(obj interface{}) (*objStruct, reflect.Value, error) {
	var err error
	objValue := reflect.ValueOf(obj)
	if objValue.Kind() == reflect.Ptr {
		objValue = objValue.Elem()
	}

	self.mutex.RLock()
	objStructRef, exists := self.objTypes[objValue.Type().Name()]
	self.mutex.RUnlock()

	if !exists {
		// Lazy initialize struct definitions.
		objStructRef, err = newObjStruct(obj)
		if err != nil {
			return objStructRef, objValue, err
		}

		self.mutex.Lock()
		self.objTypes[objValue.Type().Name()] = objStructRef
		self.mutex.Unlock()
	}

	return objStructRef, objValue, nil
}

func (self *Store) Write(ctx context.Context, obj interface{}, options ...Option) error {
	objStructRef, objValue, err := self.getObjectStruct(obj)
	if err != nil {
		return nil
	}

	pipe := self.redisClient.WithContext(ctx).Pipeline()

	if err = objStructRef.writeToRedis(ctx, self.redisClient, pipe, rootKeyPrefix, objValue, options...); err != nil {
		return err
	}

	results, _ := pipe.Exec()
	for _, result := range results {
		if err := result.Err(); err != nil && err != redis.Nil {
			return fmt.Errorf("%w: %s", ErrRedisCommandError, err)
		}
	}

	return nil
}

type readResultsCallback func(result redis.Cmder) error

func (self *Store) Read(ctx context.Context, obj interface{}) error {
	objStructRef, objValue, err := self.getObjectStruct(obj)
	if err != nil {
		return nil
	}

	pipe := self.redisClient.WithContext(ctx).Pipeline()

	callbacks := []readResultsCallback{}
	if err := objStructRef.readFromRedis(ctx, self.redisClient, pipe, &callbacks, rootKeyPrefix, objValue); err != nil {
		return err
	}

	results, _ := pipe.Exec()
	for index, result := range results {
		if err := result.Err(); err != nil && err != redis.Nil {
			return fmt.Errorf("%w: %s", ErrRedisCommandError, err)
		}

		// index-1 due to TX pipeline.
		if err := callbacks[index](result); err != nil {
			return err
		}
	}

	return nil
}
