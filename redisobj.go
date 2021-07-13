package redisobj

import (
	"fmt"
	"reflect"
	"time"

	"github.com/go-redis/redis/v7"
)

const (
	rootKey = "redisobj"
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

	if err = self.writeStruct(pipe, objStructRef, objValue, options...); err != nil {
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

func (self *Store) writeStruct(pipe redis.Pipeliner, objStructRef *objStruct, objValue reflect.Value, options ...Option) error {
	for _, embeddedObjStructRef := range objStructRef.structFields {
		objStructValue := objValue.Field(embeddedObjStructRef.structData.structIndex)
		if err := self.writeStruct(pipe, embeddedObjStructRef, objStructValue, options...); err != nil {
			return err
		}
	}

	for _, valueField := range objStructRef.valueFields {
		if err := valueField.redisWriteFn(pipe, objValue, options...); err != nil {
			return err
		}
	}

	for _, sliceField := range objStructRef.sliceFields {
		if err := sliceField.redisWriteFn(pipe, objValue, options...); err != nil {
			return err
		}
	}

	for _, mapField := range objStructRef.mapFields {
		if err := mapField.redisWriteFn(pipe, objValue, options...); err != nil {
			return err
		}
	}

	return nil
}

/*
type readResultsCallback func(result redis.Cmder) error

func (self *Store) Read(obj interface{}) error {
	objValue, objStructRef, err := self.getObjectStruct(obj)
	if err != nil {
		return nil
	}

	pipe := self.redisClient.Pipeline()

	callbacks := []readResultsCallback{}

	if err := self.readStruct(pipe, &callbacks, rootKey, objValue, objStructRef); err != nil {
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

func (self *Store) readStruct(pipe redis.Pipeliner, callbacks *[]readResultsCallback, parentKey string, objValue reflect.Value, objStructRef objStruct) error {
	for fieldName, embeddedObjStructRef := range objStructRef.structFields {
		objStructValue := objValue.FieldByName(fieldName)
		parentKey := parentKey + ":" + objStructRef.key(objValue)
		if err := self.readStruct(pipe, callbacks, parentKey, objStructValue, embeddedObjStructRef); err != nil {
			return err
		}
	}

	for _, redisValueField := range objStructRef.redisValueFields {
		if err := self.readValue(pipe, callbacks, parentKey, objValue, objStructRef, redisValueField); err != nil {
			return err
		}
	}

	for _, redisHashField := range objStructRef.redisHashFields {
		if err := self.readHash(pipe, callbacks, parentKey, objValue, objStructRef, redisHashField); err != nil {
			return err
		}
	}

	return nil
}

func (self *Store) readValue(pipe redis.Pipeliner, callbacks *[]readResultsCallback, parentKey string, objValue reflect.Value, objStructRef objStruct, redisValueRef redisValue) error {
	key := parentKey + ":" + objStructRef.key(objValue) + ":" + redisValueRef.key
	pipe.Get(key)

	callback := func(result redis.Cmder) error {
		redisValue, err := result.(*redis.StringCmd).Result()
		if err != nil {
			if err == redis.Nil {
				redisValue = ""
			} else {
				return fmt.Errorf("%w Get: %s", ErrRedisCommandError, err)
			}
		}
		return setFieldFromString(objValue.Field(redisValueRef.fieldNum), redisValue)
	}

	*callbacks = append(*callbacks, callback)

	return nil
}

func (self *Store) readHash(pipe redis.Pipeliner, callbacks *[]readResultsCallback, parentKey string, objValue reflect.Value, objStructRef objStruct, redisHashRef redisHash) error {
	key := parentKey + ":" + objStructRef.key(objValue)

	value := objValue.Field(redisHashRef.fieldNum)
	if value.Kind() == reflect.Map {
		// Maps get put into their own keys.
		key += ":" + redisHashRef.key

		pipe.HGetAll(key)

		// FIXME: There is no easy way to convert map[string]string into anything other than map[string]<T>.
		//        The writeHash() needs to change to only support map[string]<T>.

		// TODO: Finish this callback for hash.
		callback := func(result redis.Cmder) error {
			redisValue, err := result.(*redis.StringCmd).Result()
			if err != nil {
				if err == redis.Nil {
					redisValue = ""
				} else {
					return fmt.Errorf("%w Get: %s", ErrRedisCommandError, err)
				}
			}
			return setFieldFromString(objValue.Field(redisHashRef.fieldNum), redisValue)
		}

		*callbacks = append(*callbacks, callback)
	} else {
		pipe.HMGet(key, redisHashRef.key)

		// TODO: Finish this callback for hash.
		callback := func(result redis.Cmder) error {
			redisValue, err := result.(*redis.StringCmd).Result()
			if err != nil {
				if err == redis.Nil {
					redisValue = ""
				} else {
					return fmt.Errorf("%w Get: %s", ErrRedisCommandError, err)
				}
			}
			return setFieldFromString(objValue.Field(redisHashRef.fieldNum), redisValue)
		}

		*callbacks = append(*callbacks, callback)
	}

	return nil
}
*/
