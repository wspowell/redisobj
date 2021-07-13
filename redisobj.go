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
	objTypes    map[string]objStruct
}

func NewStore(redisClient redis.UniversalClient) *Store {
	return &Store{
		redisClient: redisClient,
		objTypes:    map[string]objStruct{},
	}
}

func (self *Store) getObjectStruct(obj interface{}) (reflect.Value, objStruct, error) {
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
			return objValue, objStructRef, err
		}
		self.objTypes[objValue.Type().Name()] = objStructRef
	}

	return objValue, objStructRef, nil
}

func (self *Store) Write(obj interface{}, options ...Option) error {
	objValue, objStructRef, err := self.getObjectStruct(obj)
	if err != nil {
		return nil
	}

	pipe := self.redisClient.Pipeline()

	if err = self.writeStruct(pipe, rootKey, objValue, objStructRef, options...); err != nil {
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

func (self *Store) writeStruct(pipe redis.Pipeliner, parentKey string, objValue reflect.Value, objStructRef objStruct, options ...Option) error {
	for fieldName, embeddedObjStructRef := range objStructRef.structFields {
		objStructValue := objValue.FieldByName(fieldName)
		parentKey := parentKey + ":" + objStructRef.key(objValue)
		if err := self.writeStruct(pipe, parentKey, objStructValue, embeddedObjStructRef, options...); err != nil {
			return err
		}
	}

	for _, redisValueField := range objStructRef.redisValueFields {
		self.writeValue(pipe, parentKey, objValue, objStructRef, redisValueField, options...)
	}

	for _, redisHashField := range objStructRef.redisHashFields {
		self.writeHash(pipe, parentKey, objValue, objStructRef, redisHashField, options...)
	}

	return nil
}

func (self *Store) writeValue(pipe redis.Pipeliner, parentKey string, objValue reflect.Value, objStructRef objStruct, redisValueRef redisValue, options ...Option) {
	key := parentKey + ":" + objStructRef.key(objValue) + ":" + redisValueRef.key
	value := objValue.Field(redisValueRef.fieldNum).Interface()

	var optionNx bool
	var optionXx bool
	var ttl time.Duration

	for _, option := range options {
		switch option.name {
		case OptionIfNotExists.name:
			optionNx = true
		case OptionIfExists.name:
			optionXx = true
		case "ttl":
			ttl = option.value.(time.Duration)
		}
	}

	if optionNx {
		pipe.SetNX(key, value, ttl)
	} else if optionXx {
		pipe.SetXX(key, value, ttl)
	} else {
		pipe.Set(key, value, ttl)
	}
}

func (self *Store) writeHash(pipe redis.Pipeliner, parentKey string, objValue reflect.Value, objStructRef objStruct, redisHashRef redisHash, options ...Option) {
	key := parentKey + ":" + objStructRef.key(objValue)

	var optionNx bool
	var ttl time.Duration

	for _, option := range options {
		switch option.name {
		case OptionIfNotExists.name:
			optionNx = true
		case "ttl":
			ttl = option.value.(time.Duration)
		}
	}

	value := objValue.Field(redisHashRef.fieldNum)
	if value.Kind() == reflect.Map {
		// Maps get put into their own keys.
		key += ":" + redisHashRef.key

		if len(value.MapKeys()) == 0 {
			return
		}

		// FIXME: There is no easy way to convert map[string]string into anything other than map[string]<T>.
		//        The writeHash() needs to change to only support map[string]<T>.
		valueMap := make(map[string]interface{}, len(value.MapKeys()))

		iter := value.MapRange()
		for iter.Next() {
			key := iter.Key()
			value := iter.Value()

			valueMap[fmt.Sprintf("%v", key)] = value.Interface()
		}

		if optionNx {
			for field, value := range valueMap {
				pipe.HSetNX(key, field, value)
			}
		} else {
			// There is no HSetXX. Just call HSet.
			pipe.HSet(key, valueMap)
		}
	} else {
		if optionNx {
			pipe.HSetNX(key, redisHashRef.key, value.Interface())
		} else {
			// There is no HSetXX. Just call HSet.
			pipe.HSet(key, redisHashRef.key, value.Interface())
		}
	}

	if ttl != 0 {
		pipe.Expire(key, ttl)
	}
}

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
