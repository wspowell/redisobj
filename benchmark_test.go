package redisobj_test

import (
	"context"
	"redisobj"
	"strconv"
	"testing"

	"github.com/go-redis/redis/v7"
	"github.com/google/uuid"
)

type singleVariableSingleton struct {
	Value string
}

func Benchmark_redisobj_read_singleVariableSingleton(b *testing.B) {
	redisClient := NewGoRedisClient()
	redisClient.FlushAll()

	var err error
	ctx := context.Background()

	objStore := redisobj.NewStore(redisClient)

	input := singleVariableSingleton{
		Value: uuid.New().String(),
	}
	if err = objStore.Write(ctx, input, redisobj.Options{}); err != nil {
		panic(err)
	}

	output := &singleVariableSingleton{}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err = objStore.Read(ctx, output, redisobj.Options{}); err != nil {
			panic(err)
		}
	}
}

func Benchmark_redis_read_singleVariableSingleton(b *testing.B) {
	redisClient := NewGoRedisClient()
	redisClient.FlushAll()
	var err error

	if err = redisClient.Set("redisobj:singleVariableSingleton:value", uuid.New().String(), 0).Err(); err != nil {
		panic(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err = redisClient.Get("redisobj:singleVariableSingleton:value").Err(); err != nil {
			panic(err)
		}
	}
}

func Benchmark_redisobj_write_singleVariableSingleton(b *testing.B) {
	redisClient := NewGoRedisClient()
	redisClient.FlushAll()

	var err error
	ctx := context.Background()

	objStore := redisobj.NewStore(redisClient)

	input := singleVariableSingleton{
		Value: uuid.New().String(),
	}
	if err = objStore.Write(ctx, input, redisobj.Options{}); err != nil {
		panic(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err = objStore.Write(ctx, input, redisobj.Options{}); err != nil {
			panic(err)
		}
	}
}

func Benchmark_redis_write_singleVariableSingleton(b *testing.B) {
	redisClient := NewGoRedisClient()
	redisClient.FlushAll()
	var err error

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err = redisClient.Set("redisobj:singleVariableSingleton:value", uuid.New().String(), 0).Err(); err != nil {
			panic(err)
		}
	}
}

type keyedObject struct {
	Id    string `redisobj:"key"`
	Value string
}

func Benchmark_redisobj_read_keyedObject(b *testing.B) {
	redisClient := NewGoRedisClient()
	redisClient.FlushAll()

	var err error
	ctx := context.Background()

	objStore := redisobj.NewStore(redisClient)

	id := uuid.New().String()
	input := keyedObject{
		Id:    id,
		Value: "value",
	}
	if err = objStore.Write(ctx, input, redisobj.Options{}); err != nil {
		panic(err)
	}

	output := &keyedObject{
		Id: id,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err = objStore.Read(ctx, output, redisobj.Options{}); err != nil {
			panic(err)
		}
	}
}

func Benchmark_redis_read_keyedObject(b *testing.B) {
	redisClient := NewGoRedisClient()
	redisClient.FlushAll()
	ctx := context.Background()
	var err error

	key := uuid.New().String()
	if err = redisClient.HSet("redisobj:keyedObject:"+key, "Id", key, "Value", "value").Err(); err != nil {
		panic(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		pipe := redisClient.WithContext(ctx).Pipeline()

		pipe.HMGet("redisobj:keyedObject:"+key, "Id")
		pipe.HMGet("redisobj:keyedObject:"+key, "Value")

		if _, err := pipe.Exec(); err != nil {
			panic(err)
		}
	}
}

func Benchmark_redisobj_write_keyedObject(b *testing.B) {
	redisClient := NewGoRedisClient()
	redisClient.FlushAll()

	var err error
	ctx := context.Background()

	objStore := redisobj.NewStore(redisClient)

	id := uuid.New().String()
	input := keyedObject{
		Id:    id,
		Value: "value",
	}
	if err = objStore.Write(ctx, input, redisobj.Options{}); err != nil {
		panic(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err = objStore.Write(ctx, input, redisobj.Options{}); err != nil {
			panic(err)
		}
	}
}

func Benchmark_redis_write_keyedObject(b *testing.B) {
	redisClient := NewGoRedisClient()
	redisClient.FlushAll()
	var err error

	key := uuid.New().String()
	if err = redisClient.HSet("redisobj:keyedObject:"+key, "Id", key, "Value", "value").Err(); err != nil {
		panic(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err = redisClient.HSet("redisobj:keyedObject:"+key, "Id", key, "Value", "value").Err(); err != nil {
			panic(err)
		}
	}
}

func Benchmark_redisobj_read_keyedObject_nested(b *testing.B) {
	benchmark_redisobj_read_keyedObject_nested(b, false)
}

func Benchmark_redisobj_read_keyedObject_nested_cached(b *testing.B) {
	benchmark_redisobj_read_keyedObject_nested(b, true)
}

func benchmark_redisobj_read_keyedObject_nested(b *testing.B, cacheEnabled bool) {
	redisClient := NewGoRedisClient()
	redisClient.FlushAll()
	ctx := context.Background()

	type nested struct {
		NestedString string
		NestedInt    int
		NestedMap    map[int]int
		NestedSlice  []string
	}
	type root struct {
		Id     string `redisobj:"key"`
		String string
		Int    int
		Map    map[int]int
		Slice  []string
		Nested nested
	}

	var err error

	objStore := redisobj.NewStore(redisClient)

	input := &root{
		Id:     "UUID",
		String: "root_string",
		Int:    5,
		Map: map[int]int{
			111: 222,
		},
		Slice: []string{
			"one",
			"two",
			"three",
		},
		Nested: nested{
			NestedString: "nested_string",
			NestedInt:    13,
			NestedMap: map[int]int{
				333: 444,
			},
			NestedSlice: []string{
				"four",
				"five",
				"six",
			},
		},
	}

	// Perform one write to ensure all reflection has been setup.
	if err = objStore.Write(ctx, input, redisobj.Options{
		EnableCaching: true,
	}); err != nil {
		panic(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err = objStore.Read(ctx, input, redisobj.Options{
			EnableCaching: cacheEnabled,
		}); err != nil {
			panic(err)
		}
	}
}
func Benchmark_redis_read_keyedObject_nested(b *testing.B) {
	redisClient := NewGoRedisClient()
	redisClient.FlushAll()

	var err error

	type nested struct {
		NestedString string
		NestedInt    int
		NestedMap    map[int]int
		NestedSlice  []string
	}
	type root struct {
		Id     string `redisobj:"key"`
		String string
		Int    int
		Map    map[int]int
		Slice  []string
		Nested nested
	}

	input := &root{
		Id:     "UUID",
		String: "root_string",
		Int:    5,
		Map: map[int]int{
			111: 222,
		},
		Slice: []string{
			"one",
			"two",
			"three",
		},
		Nested: nested{
			NestedString: "nested_string",
			NestedInt:    13,
			NestedMap: map[int]int{
				333: 444,
			},
			NestedSlice: []string{
				"four",
				"five",
				"six",
			},
		},
	}

	if err = redisClient.HSet("{redisobj:root:UUID}",
		"Id", input.Id,
		"String", input.String,
		"Int", strconv.Itoa(input.Int),
	).Err(); err != nil {
		panic(err)
	}
	if err = redisClient.HSet("{redisobj:root:UUID}.Map",
		111, 222,
	).Err(); err != nil {
		panic(err)
	}
	if err = redisClient.ZAdd("{redisobj:root:UUID}.Slice",
		&redis.Z{
			Score:  0,
			Member: "one",
		},
		&redis.Z{
			Score:  1,
			Member: "two",
		},
		&redis.Z{
			Score:  2,
			Member: "three",
		},
	).Err(); err != nil {
		panic(err)
	}

	if err = redisClient.HSet("{redisobj:root:UUID}:nested",
		"Id", input.Id,
		"String", input.String,
		"Int", strconv.Itoa(input.Int),
	).Err(); err != nil {
		panic(err)
	}
	if err = redisClient.HSet("{redisobj:root:UUID}:nested.Map",
		111, 222,
	).Err(); err != nil {
		panic(err)
	}
	if err = redisClient.ZAdd("{redisobj:root:UUID}:nested.Slice",
		&redis.Z{
			Score:  0,
			Member: "one",
		},
		&redis.Z{
			Score:  1,
			Member: "two",
		},
		&redis.Z{
			Score:  2,
			Member: "three",
		},
	).Err(); err != nil {
		panic(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		pipe := redisClient.Pipeline()

		pipe.HGetAll("{redisobj:root:UUID}")
		pipe.HGetAll("{redisobj:root:UUID}.Map")
		pipe.ZRange("{redisobj:root:UUID}.Slice", 0, -1)
		pipe.HGetAll("{redisobj:root:UUID}:nested")
		pipe.HGetAll("{redisobj:root:UUID}:nested.Map")
		pipe.ZRange("{redisobj:root:UUID}:nested.Slice", 0, -1)

		_, err := pipe.Exec()
		if err != nil {
			panic(err)
		}
	}
}

func Benchmark_redisobj_write_keyedObject_nested(b *testing.B) {
	benchmark_redisobj_write_keyedObject_nested(b, false)
}

func Benchmark_redisobj_write_keyedObject_nested_cached(b *testing.B) {
	benchmark_redisobj_write_keyedObject_nested(b, true)
}

func benchmark_redisobj_write_keyedObject_nested(b *testing.B, cacheEnabled bool) {
	redisClient := NewGoRedisClient()
	redisClient.FlushAll()
	ctx := context.Background()

	type nested struct {
		NestedString string
		NestedInt    int
		NestedMap    map[int]int
		NestedSlice  []string
	}
	type root struct {
		Id     string `redisobj:"key"`
		String string
		Int    int
		Map    map[int]int
		Slice  []string
		Nested nested
	}

	var err error

	objStore := redisobj.NewStore(redisClient)

	input := &root{
		Id:     "UUID",
		String: "root_string",
		Int:    5,
		Map: map[int]int{
			111: 222,
		},
		Slice: []string{
			"one",
			"two",
			"three",
		},
		Nested: nested{
			NestedString: "nested_string",
			NestedInt:    13,
			NestedMap: map[int]int{
				333: 444,
			},
			NestedSlice: []string{
				"four",
				"five",
				"six",
			},
		},
	}

	// Perform one write to ensure all reflection has been setup.
	if err = objStore.Write(ctx, input, redisobj.Options{
		EnableCaching: cacheEnabled,
	}); err != nil {
		panic(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err = objStore.Write(ctx, input, redisobj.Options{
			EnableCaching: cacheEnabled,
		}); err != nil {
			panic(err)
		}
	}
}

func Benchmark_redis_write_keyedObject_nested(b *testing.B) {
	redisClient := NewGoRedisClient()
	redisClient.FlushAll()

	var err error

	type nested struct {
		NestedString string
		NestedInt    int
		NestedMap    map[int]int
		NestedSlice  []string
	}
	type root struct {
		Id     string `redisobj:"key"`
		String string
		Int    int
		Map    map[int]int
		Slice  []string
		Nested nested
	}

	input := &root{
		Id:     "UUID",
		String: "root_string",
		Int:    5,
		Map: map[int]int{
			111: 222,
		},
		Slice: []string{
			"one",
			"two",
			"three",
		},
		Nested: nested{
			NestedString: "nested_string",
			NestedInt:    13,
			NestedMap: map[int]int{
				333: 444,
			},
			NestedSlice: []string{
				"four",
				"five",
				"six",
			},
		},
	}

	if err = redisClient.HSet("{redisobj:root:UUID}",
		"Id", input.Id,
		"String", input.String,
		"Int", strconv.Itoa(input.Int),
	).Err(); err != nil {
		panic(err)
	}
	if err = redisClient.HSet("{redisobj:root:UUID}.Map",
		111, 222,
	).Err(); err != nil {
		panic(err)
	}
	if err = redisClient.ZAdd("{redisobj:root:UUID}.Slice",
		&redis.Z{
			Score:  0,
			Member: "one",
		},
		&redis.Z{
			Score:  1,
			Member: "two",
		},
		&redis.Z{
			Score:  2,
			Member: "three",
		},
	).Err(); err != nil {
		panic(err)
	}

	if err = redisClient.HSet("{redisobj:root:UUID}:nested",
		"Id", input.Id,
		"String", input.String,
		"Int", strconv.Itoa(input.Int),
	).Err(); err != nil {
		panic(err)
	}
	if err = redisClient.HSet("{redisobj:root:UUID}:nested.Map",
		111, 222,
	).Err(); err != nil {
		panic(err)
	}
	if err = redisClient.ZAdd("{redisobj:root:UUID}:nested.Slice",
		&redis.Z{
			Score:  0,
			Member: "one",
		},
		&redis.Z{
			Score:  1,
			Member: "two",
		},
		&redis.Z{
			Score:  2,
			Member: "three",
		},
	).Err(); err != nil {
		panic(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		pipe := redisClient.Pipeline()

		pipe.HSet("{redisobj:root:UUID}",
			"Id", input.Id,
			"String", input.String,
			"Int", strconv.Itoa(input.Int),
		)
		pipe.HSet("{redisobj:root:UUID}.Map",
			111, 222,
		)
		pipe.ZAdd("{redisobj:root:UUID}.Slice",
			&redis.Z{
				Score:  0,
				Member: "one",
			},
			&redis.Z{
				Score:  1,
				Member: "two",
			},
			&redis.Z{
				Score:  2,
				Member: "three",
			},
		)

		pipe.HSet("{redisobj:root:UUID}:nested",
			"Id", input.Id,
			"String", input.String,
			"Int", strconv.Itoa(input.Int),
		)
		pipe.HSet("{redisobj:root:UUID}:nested.Map",
			111, 222,
		)
		pipe.ZAdd("{redisobj:root:UUID}:nested.Slice",
			&redis.Z{
				Score:  0,
				Member: "one",
			},
			&redis.Z{
				Score:  1,
				Member: "two",
			},
			&redis.Z{
				Score:  2,
				Member: "three",
			},
		)

		_, err := pipe.Exec()
		if err != nil {
			panic(err)
		}
	}
}
