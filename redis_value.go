package redisobj

const (
	// structTagRedisValue defines the struct tag key for a "value" (i.e. uses SET/GET).
	structTagRedisValue = "redisValue"
)

// redisValue defines any value stored by SET and retrieved by GET.
type redisValue struct {
	// reflection information

	fieldNum int

	// redis information

	key string
}

func newRedisValue(fieldNum int, tagValue string) (redisValue, error) {
	return redisValue{
		fieldNum: fieldNum,
		key:      tagValue,
	}, nil
}
