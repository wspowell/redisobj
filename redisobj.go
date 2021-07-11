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

	return self.writeStruct(rootKey, objValue, objStructRef, options...)
}

func (self *Store) writeStruct(parentKey string, objValue reflect.Value, objStructRef objStruct, options ...Option) error {
	for fieldName, embeddedObjStructRef := range objStructRef.structFields {
		objStructValue := objValue.FieldByName(fieldName)
		parentKey := parentKey + ":" + objStructRef.key(objValue)
		if err := self.writeStruct(parentKey, objStructValue, embeddedObjStructRef, options...); err != nil {
			return err
		}
	}

	for _, redisValueField := range objStructRef.redisValueFields {
		if err := self.writeValue(parentKey, objValue, objStructRef, redisValueField, options...); err != nil {
			return err
		}
	}

	return nil
}

func (self *Store) writeValue(parentKey string, objValue reflect.Value, objStructRef objStruct, redisValueRef redisValue, options ...Option) error {
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

	fmt.Println("write key", key)

	if optionNx {
		if err := self.redisClient.SetNX(key, value, ttl).Err(); err != nil {
			return fmt.Errorf("%w SetNX: %s", ErrRedisCommandError, err)
		}
	} else if optionXx {
		// FIXME: XX is an impossible case with the current workflow.
		//        The key must be able to be set somehow or else it will never be set.
		if err := self.redisClient.SetXX(key, value, ttl).Err(); err != nil {
			return fmt.Errorf("%w SetXX: %s", ErrRedisCommandError, err)
		}
	} else {
		if err := self.redisClient.Set(key, value, ttl).Err(); err != nil {
			return fmt.Errorf("%w Set: %s", ErrRedisCommandError, err)
		}
	}

	return nil
}

func (self *Store) Read(obj interface{}) error {
	objValue, objStructRef, err := self.getObjectStruct(obj)
	if err != nil {
		return nil
	}

	return self.readStruct(rootKey, objValue, objStructRef)
}

func (self *Store) readStruct(parentKey string, objValue reflect.Value, objStructRef objStruct) error {
	for fieldName, embeddedObjStructRef := range objStructRef.structFields {
		objStructValue := objValue.FieldByName(fieldName)
		parentKey := parentKey + ":" + objStructRef.key(objValue)
		if err := self.readStruct(parentKey, objStructValue, embeddedObjStructRef); err != nil {
			return err
		}
	}

	for _, redisValueField := range objStructRef.redisValueFields {
		if err := self.readValue(parentKey, objValue, objStructRef, redisValueField); err != nil {
			return err
		}
	}

	return nil
}

func (self *Store) readValue(parentKey string, objValue reflect.Value, objStructRef objStruct, redisValueRef redisValue) error {
	key := parentKey + ":" + objStructRef.key(objValue) + ":" + redisValueRef.key

	fmt.Println("read key", key)

	redisValue, err := self.redisClient.Get(key).Result()
	if err != nil {
		if err == redis.Nil {
			redisValue = ""
		} else {
			return fmt.Errorf("%w Get: %s", ErrRedisCommandError, err)
		}
	}

	if err := setFieldFromString(objValue.Field(redisValueRef.fieldNum), redisValue); err != nil {
		return err
	}

	return nil
}
