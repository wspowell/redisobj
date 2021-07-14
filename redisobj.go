package redisobj

import (
	"fmt"
	"reflect"
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
	OptionETag = newOption("ETAG", nil)
)

type Writer interface {
	Write(obj interface{}) error
}

type Reader interface {
	Read(obj interface{}) error
}

type Store struct {
	redisClient redis.UniversalClient
	objTypes    map[string]*objStruct
}

func NewStore(redisClient redis.UniversalClient) *Store {
	return &Store{
		redisClient: redisClient,
		objTypes:    map[string]*objStruct{},
	}
}

func (self *Store) getObjectStruct(obj interface{}) (*objStruct, reflect.Value, error) {
	var err error
	objValue := reflect.ValueOf(obj)
	if objValue.Kind() == reflect.Ptr {
		objValue = objValue.Elem()
	}
	objStructRef, exists := self.objTypes[objValue.Type().Name()]
	if !exists {
		// Lazy initialize struct definitions.
		objStructRef, err = newObjStruct(obj)
		if err != nil {
			return objStructRef, objValue, err
		}
		self.objTypes[objValue.Type().Name()] = objStructRef
	}

	return objStructRef, objValue, nil
}

func (self *Store) Write(obj interface{}, options ...Option) error {
	objStructRef, objValue, err := self.getObjectStruct(obj)
	if err != nil {
		return nil
	}

	pipe := self.redisClient.Pipeline()

	if err = objStructRef.writeToRedis(pipe, rootKeyPrefix, objValue, options...); err != nil {
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

func (self *Store) Read(obj interface{}) error {
	objStructRef, objValue, err := self.getObjectStruct(obj)
	if err != nil {
		return nil
	}

	pipe := self.redisClient.Pipeline()

	callbacks := []readResultsCallback{}
	if err := objStructRef.readFromRedis(pipe, &callbacks, rootKeyPrefix, objValue); err != nil {
		return err
	}

	results, _ := pipe.Exec()
	for index, result := range results {
		if err := result.Err(); err != nil && err != redis.Nil {
			return fmt.Errorf("%w: %s", ErrRedisCommandError, err)
		}

		if err := callbacks[index](result); err != nil {
			return err
		}
	}

	return nil
}
