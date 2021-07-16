package redisobj

import (
	"errors"
)

var (
	ErrInvalidRedisDefinition = errors.New("provided redis defintion is not valid")
	ErrInvalidObject          = errors.New("invalid object")
	ErrInvalidFieldType       = errors.New("invalid field type")
	ErrObjectNotFound         = errors.New("object not found")
	ErrRedisCommandError      = errors.New("failed executing redis command")
	ErrCacheFailure           = errors.New("failure checking redis object cache")
)
