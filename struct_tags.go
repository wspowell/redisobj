package redisobj

import (
	"fmt"
	"reflect"
	"strings"
)

// objStruct defines the reflection parameters of the object type.
// This struct saves reflection information on an object to avoid repeated reflection operations during use.
type objStruct struct {
	objType reflect.Type
	// objName of this struct. This is the struct name.
	// This is prepended to all keys in this struct.
	objName string

	keyField int

	// structFields stores definitions of all embedded structs by name.
	// Embedded structs will form nested keys.
	structFields map[string]objStruct

	// redisValueFields stores definitions of all redis values.
	redisValueFields []redisValue
}

func (self objStruct) key(objValue reflect.Value) string {

	key := self.objName

	if self.keyField != -1 {
		keyValue := fmt.Sprintf("%v", objValue.Field(self.keyField).Interface())
		if keyValue == "" {
			keyValue = "none"
		}
		key = fmt.Sprintf("%s:%v", key, keyValue)
	}

	return key
}

func newObjStruct(obj interface{}) (objStruct, error) {
	objStructRef := objStruct{}

	// TypeOf returns the reflection Type that represents the dynamic type of variable.
	// If variable is a nil interface value, TypeOf returns nil.
	objType := reflect.TypeOf(obj)
	objValue := reflect.ValueOf(obj)
	if objType.Kind() == reflect.Ptr {
		objType = objType.Elem()
		objValue = objValue.Elem()
	}
	objStructRef.objName = objType.Name()
	objStructRef.objType = objType
	objStructRef.keyField = -1

	structFields := map[string]objStruct{}
	redisValueFields := []redisValue{}

	// Iterate over all available fields and read the tag value
	for i := 0; i < objType.NumField(); i++ {
		fieldValue := objValue.Field(i)
		fieldType := objType.Field(i)

		if fieldType.Type.Kind() == reflect.Struct {
			embeddedObjStructRef, err := newObjStruct(fieldValue.Interface())
			if err != nil {
				return objStructRef, err
			}

			structFields[fieldType.Name] = embeddedObjStructRef
		} else {
			if tagValue, exists := fieldType.Tag.Lookup(structTagRedisValue); exists {
				tagParts := strings.SplitN(tagValue, ",", 2)
				if len(tagParts) == 2 && strings.EqualFold(tagParts[1], "key") {
					objStructRef.keyField = i
				}

				redisValueRef, err := newRedisValue(i, tagParts[0])
				if err != nil {
					return objStruct{}, err
				}
				redisValueFields = append(redisValueFields, redisValueRef)
			}
		}
	}

	objStructRef.structFields = structFields
	objStructRef.redisValueFields = redisValueFields

	return objStructRef, nil
}
