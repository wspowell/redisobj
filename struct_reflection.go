package redisobj

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/mitchellh/hashstructure/v2"
)

const (
	// structTagRedisobj defines the struct tag key for all redisobj struct tag options.
	structTagKeyRedisobj = "redisobj"
	structTagValueKey    = "key"
)

type reflectionData struct {
	objType      reflect.Type
	objName      string
	structIndex  int
	redisWriteFn func(pipe redis.Pipeliner, keyPrefix string, objValue reflect.Value, options ...Option) error
	redisReadFn  func(pipe redis.Pipeliner, keyPrefix string, objValue reflect.Value) readResultsCallback
}

// objStruct defines the reflection parameters of the object type.
// This struct saves reflection information on an object to avoid repeated reflection operations during use.
type objStruct struct {
	structData    reflectionData
	keyFieldIndex int
	valueFields   []*reflectionData
	sliceFields   []*reflectionData
	mapFields     []*reflectionData
	structFields  []*objStruct
}

func newObjStruct(obj interface{}) (*objStruct, error) {
	objType := reflect.TypeOf(obj)
	objValue := reflect.ValueOf(obj)
	if objType.Kind() == reflect.Ptr {
		objType = objType.Elem()
		objValue = objValue.Elem()
	}

	objStructRef := &objStruct{
		structData: reflectionData{
			objType:     objType,
			objName:     objType.Name(),
			structIndex: -1,
		},
		keyFieldIndex: -1,
		valueFields:   []*reflectionData{},
		sliceFields:   []*reflectionData{},
		mapFields:     []*reflectionData{},
		structFields:  []*objStruct{},
	}

	// Iterate over all available fields and read the tag value
	for structFieldIndex := 0; structFieldIndex < objType.NumField(); structFieldIndex++ {
		fieldValue := objValue.Field(structFieldIndex)
		fieldType := objType.Field(structFieldIndex)

		switch fieldType.Type.Kind() {
		case reflect.Struct:
			// Recurse over embedded structs.
			structField, err := newObjStruct(fieldValue.Interface())
			if err != nil {
				return nil, err
			}
			structField.structData.structIndex = structFieldIndex
			objStructRef.structFields = append(objStructRef.structFields, structField)

		case reflect.Slice:
			// TODO: This could probably support struct values with a bit more effort.
			if !isStringParsable(fieldType.Type.Elem()) {
				return nil, fmt.Errorf("%w: %s", ErrInvalidFieldType, "slice values must be a primitive type that is string parsable with strconv")
			}
			data := &reflectionData{
				objType:     fieldType.Type,
				objName:     fieldType.Name,
				structIndex: structFieldIndex,
			}
			data.redisWriteFn = func(pipe redis.Pipeliner, keyPrefix string, objValue reflect.Value, options ...Option) error {
				key := keyPrefix + "." + data.objName
				sliceField := objValue.Field(data.structIndex)

				var ttl time.Duration

				for _, option := range options {
					switch option.name {
					case "ttl":
						ttl = option.value.(time.Duration)
					}
				}

				if sliceField.Len() == 0 {
					return nil
				}

				valueSlice := make([]*redis.Z, sliceField.Len())

				for i := 0; i < sliceField.Len(); i++ {
					value := sliceField.Index(i)

					valueString, err := valueToString(value)
					if err != nil {
						return err
					}
					valueSlice[i] = &redis.Z{
						Score:  float64(i),
						Member: valueString,
					}
				}

				pipe.Del(key)
				pipe.ZAdd(key, valueSlice...)

				if ttl != 0 {
					pipe.Expire(key, ttl)
				}

				return nil
			}
			data.redisReadFn = func(pipe redis.Pipeliner, keyPrefix string, objValue reflect.Value) readResultsCallback {
				key := keyPrefix + "." + data.objName
				sliceField := objValue.Field(data.structIndex)

				pipe.ZRange(key, 0, -1)

				return func(result redis.Cmder) error {
					redisValue, err := result.(*redis.StringSliceCmd).Result()
					if err != nil {
						if err == redis.Nil {
							redisValue = []string{}
						} else {
							return fmt.Errorf("%w Get: %s", ErrRedisCommandError, err)
						}
					}

					sliceField.Set(reflect.MakeSlice(data.objType, len(redisValue), len(redisValue)))
					for index, readValue := range redisValue {
						value := reflect.New(data.objType.Elem()).Elem()
						if err := setFieldFromString(value, readValue); err != nil {
							return err
						}

						sliceIndex := sliceField.Index(index)
						sliceIndex.Set(value)
					}

					return nil
				}
			}

			objStructRef.sliceFields = append(objStructRef.sliceFields, data)

		case reflect.Map:
			if !isStringParsable(fieldType.Type.Key()) {
				return nil, fmt.Errorf("%w: %s", ErrInvalidFieldType, "map keys must be a primitive type that is string parsable with strconv")
			}

			// TODO: This could probably support struct values with a bit more effort.
			if !isStringParsable(fieldType.Type.Elem()) {
				return nil, fmt.Errorf("%w: %s", ErrInvalidFieldType, "map values must be a primitive type that is string parsable with strconv")
			}
			data := &reflectionData{
				objType:     fieldType.Type,
				objName:     fieldType.Name,
				structIndex: structFieldIndex,
			}
			data.redisWriteFn = func(pipe redis.Pipeliner, keyPrefix string, objValue reflect.Value, options ...Option) error {
				key := keyPrefix + "." + data.objName
				mapField := objValue.Field(data.structIndex)

				var ttl time.Duration

				for _, option := range options {
					switch option.name {
					case "ttl":
						ttl = option.value.(time.Duration)
					}
				}

				if mapField.Len() == 0 {
					return nil
				}

				valueMap := make(map[string]interface{}, mapField.Len())

				iter := mapField.MapRange()
				for iter.Next() {
					key := iter.Key()
					value := iter.Value()

					keyString, err := valueToString(key)
					if err != nil {
						return err
					}
					valueString, err := valueToString(value)
					if err != nil {
						return err
					}
					valueMap[keyString] = valueString
				}

				pipe.Del(key)
				pipe.HSet(key, valueMap)

				if ttl != 0 {
					pipe.Expire(key, ttl)
				}

				return nil
			}
			data.redisReadFn = func(pipe redis.Pipeliner, keyPrefix string, objValue reflect.Value) readResultsCallback {
				key := keyPrefix + "." + data.objName
				mapField := objValue.Field(data.structIndex)
				mapField.Set(reflect.MakeMap(data.objType))

				pipe.HGetAll(key)

				return func(result redis.Cmder) error {
					redisValue, err := result.(*redis.StringStringMapCmd).Result()
					if err != nil {
						if err == redis.Nil {
							redisValue = map[string]string{}
						} else {
							return fmt.Errorf("%w Get: %s", ErrRedisCommandError, err)
						}
					}

					for readKey, readValue := range redisValue {
						keyValue := reflect.New(data.objType.Key()).Elem()
						if err := setFieldFromString(keyValue, readKey); err != nil {
							return err
						}

						valueValue := reflect.New(data.objType.Elem()).Elem()
						if err := setFieldFromString(valueValue, readValue); err != nil {
							return err
						}

						mapField.SetMapIndex(keyValue, valueValue)
					}

					return nil
				}
			}

			objStructRef.mapFields = append(objStructRef.mapFields, data)
		default:
			if tagValue, exists := fieldType.Tag.Lookup(structTagKeyRedisobj); exists {
				if strings.EqualFold(tagValue, structTagValueKey) {
					objStructRef.keyFieldIndex = structFieldIndex
				}
			}

			data := &reflectionData{
				objType:     fieldType.Type,
				objName:     fieldType.Name,
				structIndex: structFieldIndex,
			}
			data.redisWriteFn = func(pipe redis.Pipeliner, keyPrefix string, objValue reflect.Value, options ...Option) error {
				key := keyPrefix
				value := objValue.Field(data.structIndex).Interface()

				var ttl time.Duration

				for _, option := range options {
					switch option.name {
					case "ttl":
						ttl = option.value.(time.Duration)
					}
				}

				pipe.HSet(key, data.objName, value)

				if ttl != 0 {
					pipe.Expire(key, ttl)
				}

				return nil
			}
			data.redisReadFn = func(pipe redis.Pipeliner, keyPrefix string, objValue reflect.Value) readResultsCallback {
				key := keyPrefix

				pipe.HGet(key, data.objName)

				return func(result redis.Cmder) error {
					redisValue, err := result.(*redis.StringCmd).Result()
					if err != nil {
						if err == redis.Nil {
							// Return a "not found" error if this was a key.
							if objStructRef.keyFieldIndex == data.structIndex {
								return ErrObjectNotFound
							}

							redisValue = ""
						} else {
							return fmt.Errorf("%w: %s", ErrRedisCommandError, err)
						}
					}
					return setFieldFromString(objValue.Field(data.structIndex), redisValue)
				}
			}

			objStructRef.valueFields = append(objStructRef.valueFields, data)
		}
	}

	return objStructRef, nil
}

func (self objStruct) key(objValue reflect.Value) string {
	key := self.structData.objName

	if self.keyFieldIndex != -1 {
		keyValue, err := valueToString(objValue.Field(self.keyFieldIndex))
		if err != nil || keyValue == "" {
			keyValue = "none"
		}

		key += ":" + keyValue
	}

	return key
}

func (self objStruct) isCacheFresh(ctx context.Context, redisClient *redis.Client, key string, objValue reflect.Value, ttl time.Duration, write bool) (bool, error) {
	objHash, err := hashstructure.Hash(objValue.Interface(), hashstructure.FormatV2, nil)
	if err != nil {
		return false, fmt.Errorf("%w: %s", ErrCacheFailure, err)
	}
	hashString := strconv.FormatUint(objHash, 10)

	hashKey := key + ".__HASH__"

	if !write {
		ttl, err = redisClient.WithContext(ctx).TTL(hashKey).Result()
		if err != nil {
			return false, fmt.Errorf("%w: %s", ErrCacheFailure, err)
		}
	}

	pipe := redisClient.WithContext(ctx).TxPipeline()
	pipe.Get(hashKey)
	if write {
		pipe.Set(hashKey, hashString, ttl)
	}
	results, _ := pipe.Exec()

	previousHash, err := results[0].(*redis.StringCmd).Result()
	if err != nil && err != redis.Nil {
		return false, fmt.Errorf("%w: %s", ErrCacheFailure, err)
	}

	if write {
		status, err := results[1].(*redis.StatusCmd).Result()
		if err != nil && err != redis.Nil {
			return false, fmt.Errorf("%w: %s", ErrCacheFailure, err)
		}
		if status != "OK" {
			return false, nil
		}
	}

	return hashString == previousHash, nil
}

func (self objStruct) writeToRedis(ctx context.Context, redisClient *redis.Client, pipe redis.Pipeliner, keyPrefix string, objValue reflect.Value, options ...Option) error {
	key := keyPrefix + ":" + self.key(objValue)

	// This struct is the root or is keyed so utilize hash tags to co-locate the data on the same redis node.
	if self.structData.structIndex == -1 || self.keyFieldIndex != -1 {
		key = "{" + key + "}"

		var ttl time.Duration

		for _, option := range options {
			switch option.name {
			case "ttl":
				ttl = option.value.(time.Duration)
			}
		}

		if fresh, err := self.isCacheFresh(ctx, redisClient, key, objValue, ttl, true); err != nil {
			return err
		} else if fresh {
			// Do not write anything for this struct.
			//fmt.Println("write cache hit")
			return nil
		}
		//fmt.Println("write cache miss")
	}

	pipe.Del(key)

	for _, structField := range self.structFields {
		objStructValue := objValue.Field(structField.structData.structIndex)

		var childKeyPrefix string

		// If the nested struct has a key, then treat this struct as unique data.
		if structField.keyFieldIndex != -1 {
			childKeyPrefix = keyPrefix
		} else {
			childKeyPrefix = key
		}

		if err := structField.writeToRedis(ctx, redisClient, pipe, childKeyPrefix, objStructValue, options...); err != nil {
			return err
		}
	}

	for _, valueField := range self.valueFields {
		if err := valueField.redisWriteFn(pipe, key, objValue, options...); err != nil {
			return err
		}
	}

	for _, sliceField := range self.sliceFields {
		if err := sliceField.redisWriteFn(pipe, key, objValue, options...); err != nil {
			return err
		}
	}

	for _, mapField := range self.mapFields {
		if err := mapField.redisWriteFn(pipe, key, objValue, options...); err != nil {
			return err
		}
	}

	return nil
}

func (self objStruct) readFromRedis(ctx context.Context, redisClient *redis.Client, pipe redis.Pipeliner, callbacks *[]readResultsCallback, keyPrefix string, objValue reflect.Value) error {
	key := keyPrefix + ":" + self.key(objValue)

	// This struct is the root or is keyed so utilize hash tags to co-locate the data on the same redis node.
	if self.structData.structIndex == -1 || self.keyFieldIndex != -1 {
		key = "{" + key + "}"

		if fresh, err := self.isCacheFresh(ctx, redisClient, key, objValue, 0, false); err != nil {
			return err
		} else if fresh {
			// Do not read anything for this struct.
			//fmt.Println("read cache hit")
			return nil
		}
		//fmt.Println("read cache miss")
	}

	for _, structField := range self.structFields {
		objStructValue := objValue.Field(structField.structData.structIndex)

		var childKeyPrefix string

		// If the nested struct has a key, then treat this struct as unique data.
		if structField.keyFieldIndex != -1 {
			childKeyPrefix = keyPrefix
		} else {
			childKeyPrefix = key
		}

		if err := structField.readFromRedis(ctx, redisClient, pipe, callbacks, childKeyPrefix, objStructValue); err != nil {
			return err
		}
	}

	for _, valueField := range self.valueFields {
		callback := valueField.redisReadFn(pipe, key, objValue)
		*callbacks = append(*callbacks, callback)
	}

	for _, sliceField := range self.sliceFields {
		callback := sliceField.redisReadFn(pipe, key, objValue)
		*callbacks = append(*callbacks, callback)
	}

	for _, mapField := range self.mapFields {
		callback := mapField.redisReadFn(pipe, key, objValue)
		*callbacks = append(*callbacks, callback)
	}

	return nil
}
