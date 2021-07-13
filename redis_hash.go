package redisobj

const (
	// structTagRedisHash defines the struct tag key for a "hash" (i.e. uses HSET/HGETALL).
	structTagRedisHash = "redisHash"
)

// redisHash defines any value stored by SET and retrieved by GET.
type redisHash struct {
	// reflection information

	fieldNum int

	// redis information

	key string
}

func newRedisHash(fieldNum int, tagValue string) (redisHash, error) {
	return redisHash{
		fieldNum: fieldNum,
		key:      tagValue,
	}, nil
}
