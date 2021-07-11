package redisobj

import "errors"

var (
	ErrInvalidRedisDefinition = errors.New("provided redis defintion is not valid")
	ErrInvalidFieldType       = errors.New("invalid field type")
	ErrRedisCommandError      = errors.New("failed executing redis command")
)
