package redisobj

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/go-redis/redis/v7"
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
	redisWriteFn func(pipe redis.Pipeliner, options ...Option)
}

// objStruct defines the reflection parameters of the object type.
// This struct saves reflection information on an object to avoid repeated reflection operations during use.
type objStruct struct {
	parentStruct  *objStruct
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
			objName:     reflect.TypeOf(obj).Name(),
			structIndex: -1,
		},
		valueFields:  []*reflectionData{},
		sliceFields:  []*reflectionData{},
		mapFields:    []*reflectionData{},
		structFields: []*objStruct{},
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
			structField.parentStruct = objStructRef
			structField.structData.structIndex = structFieldIndex
			objStructRef.structFields = append(objStructRef.structFields, structField)

		case reflect.Slice:
			objStructRef.sliceFields = append(objStructRef.sliceFields, &reflectionData{
				objType:     fieldType.Type,
				objName:     fieldType.Name,
				structIndex: structFieldIndex,
			})

		case reflect.Map:
			if !isStringParsable(fieldType.Type.Key()) {
				return nil, fmt.Errorf("%w: %s", ErrInvalidFieldType, "map keys must be a primitive type that is string parsable with strconv")
			}

			// TODO: This could probably support struct values with a bit more effort.
			if !isStringParsable(fieldType.Type.Elem()) {
				return nil, fmt.Errorf("%w: %s", ErrInvalidFieldType, "map values must be a primitive type that is string parsable with strconv")
			}

			objStructRef.mapFields = append(objStructRef.mapFields, &reflectionData{
				objType:     fieldType.Type,
				objValue:    fieldValue,
				objName:     fieldType.Name,
				structIndex: structFieldIndex,
			})
		default:
			if tagValue, exists := fieldType.Tag.Lookup(structTagKeyRedisobj); exists {
				if strings.EqualFold(tagValue, structTagValueKey) {
					objStructRef.keyFieldIndex = structFieldIndex
				}

				data := reflectionData{
					objType:     fieldType.Type,
					objValue:    fieldValue,
					objName:     fieldType.Name,
					structIndex: structFieldIndex,
				}
				data.redisWriteFn = func(pipe redis.Pipeliner, options ...Option) {
					dataPtr := &data
					key := structPtr.key() + ":" + objStructRef.key(objValue) + ":" + redisValueRef.key
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

				objStructRef.valueFields = append(objStructRef.valueFields, data)

			}
		}
	}

	return objStructRef, nil
}

func (self objStruct) key() string {
	key := self.structData.objName

	if self.keyFieldIndex != -1 {
		keyValue := fmt.Sprintf("%v", self.structData.objValue.Field(self.keyFieldIndex).Interface())
		if keyValue == "" {
			keyValue = "none"
		}
		// TODO: Sprintf() is the quick/easy, but slow solution.
		key = fmt.Sprintf("%s:%v", key, keyValue)
	}

	return key
}
