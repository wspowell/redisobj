package redisobj_test

// import (
// 	"redisobj"
// 	"testing"

// 	"github.com/google/uuid"
// )

// type singleVariableRedisValue struct {
// 	Value string `redisValue:"value"`
// }

// func Benchmark_redisobj_read_singleVariableRedisValue(b *testing.B) {
// 	redisClient := NewGoRedisClient()
// 	redisClient.FlushAll()
// 	var err error

// 	objStore := redisobj.NewStore(redisClient)

// 	input := singleVariableRedisValue{
// 		Value: uuid.New().String(),
// 	}
// 	if err = objStore.Write(input); err != nil {
// 		panic(err)
// 	}

// 	output := &singleVariableRedisValue{}

// 	b.ResetTimer()

// 	for i := 0; i < b.N; i++ {
// 		if err = objStore.Read(output); err != nil {
// 			panic(err)
// 		}
// 	}
// }

// func Benchmark_redis_read_singleVariableRedisValue(b *testing.B) {
// 	redisClient := NewGoRedisClient()
// 	redisClient.FlushAll()
// 	var err error

// 	if err = redisClient.Set("redisobj:singleVariableRedisValue:value", uuid.New().String(), 0).Err(); err != nil {
// 		panic(err)
// 	}

// 	b.ResetTimer()

// 	for i := 0; i < b.N; i++ {
// 		if err = redisClient.Get("redisobj:singleVariableRedisValue:value").Err(); err != nil {
// 			panic(err)
// 		}
// 	}
// }

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
