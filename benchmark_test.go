package redisobj_test

import (
	"context"
	"redisobj"
	"testing"

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
	if err = objStore.Write(ctx, input, redisobj.Options{
		EnableCaching: true,
	}); err != nil {
		panic(err)
	}

	output := &singleVariableSingleton{}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err = objStore.Read(ctx, output, redisobj.Options{
			EnableCaching: true,
		}); err != nil {
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
	var err error

	key := uuid.New().String()
	if err = redisClient.Set("redisobj:keyedObject:"+key, "value", 0).Err(); err != nil {
		panic(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err = redisClient.Get("redisobj:keyedObject:" + key).Err(); err != nil {
			panic(err)
		}
	}
}

// func Benchmark_redisobj_write_singleVariableRedisValue(b *testing.B) {
// 	redisClient := NewGoRedisClient()
// 	redisClient.FlushAll()
// 	var err error

// 	objStore := redisobj.NewStore(redisClient)

// 	input := singleVariableRedisValue{
// 		Value: uuid.New().String(),
// 	}

// 	// Perform one write to ensure all reflection has been setup.
// 	if err = objStore.Write(input); err != nil {
// 		panic(err)
// 	}

// 	b.ResetTimer()

// 	for i := 0; i < b.N; i++ {
// 		if err = objStore.Write(input); err != nil {
// 			panic(err)
// 		}
// 	}
// }

// func Benchmark_redis_write_singleVariableRedisValue(b *testing.B) {
// 	redisClient := NewGoRedisClient()
// 	redisClient.FlushAll()
// 	var err error

// 	value := uuid.New().String()

// 	b.ResetTimer()

// 	for i := 0; i < b.N; i++ {
// 		if err = redisClient.Set("redisobj:singleVariableRedisValue:value", value, 0).Err(); err != nil {
// 			panic(err)
// 		}
// 	}
// }

// type singleVariableRedisHash struct {
// 	Hash map[string]interface{} `redisHash:"hash"`
// }

// func Benchmark_redisobj_read_singleVariableRedisHash(b *testing.B) {
// 	redisClient := NewGoRedisClient()
// 	redisClient.FlushAll()
// 	var err error

// 	objStore := redisobj.NewStore(redisClient)

// 	input := singleVariableRedisHash{
// 		Hash: map[string]interface{}{
// 			"key": "value",
// 		},
// 	}
// 	if err = objStore.Write(input); err != nil {
// 		panic(err)
// 	}

// 	output := &singleVariableRedisHash{}

// 	b.ResetTimer()

// 	for i := 0; i < b.N; i++ {
// 		if err = objStore.Read(output); err != nil {
// 			panic(err)
// 		}
// 	}
// }
