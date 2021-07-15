package redisobj_test

import (
	"context"
	"fmt"
	"redisobj"
	"testing"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/stretchr/testify/assert"
)

func NewGoRedisClient() *redis.Client {
	options := &redis.Options{
		Addr: "redis:6379",
	}
	return redis.NewClient(options)
}

var (
	ttlInfinite  = time.Duration(-1)
	ttlNotExists = time.Duration(-2)
)

func Test_Store(t *testing.T) {
	redisClient := NewGoRedisClient()
	redisClient.FlushAll()
	ctx := context.Background()

	type nested struct {
		NestedString string
		NestedInt    int `redisobj:"key"`
	}
	type value struct {
		Foo    string `redisobj:"key"`
		Bar    int
		Har    map[int]int
		Lar    []string
		Nested nested
	}

	objStore := redisobj.NewStore(redisClient)

	testCases := []struct {
		description          string
		writeObject          value
		options              []redisobj.Option
		expectedReadObject   value
		expectedFooTtl       time.Duration
		expectedBarTtl       time.Duration
		expectedNestedFooTtl time.Duration
		expectedNestedBarTtl time.Duration
	}{
		{
			description: "",
			writeObject: value{
				Foo: "my_foo",
				Bar: 5,
				Har: map[int]int{
					111: 222,
				},
				Lar: []string{
					"one",
					"two",
					"three",
				},
				Nested: nested{
					NestedString: "nested_foo",
					NestedInt:    15,
				},
			},
			options: []redisobj.Option{},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			var err error

			err = objStore.Write(ctx, testCase.writeObject, testCase.options...)
			assert.Nil(t, err)

			err = objStore.Write(ctx, testCase.writeObject, testCase.options...)
			assert.Nil(t, err)

			{
				readObject := value{
					Foo: "my_foo",
					Nested: nested{
						NestedInt: 15,
					},
				}
				err = objStore.Read(ctx, &readObject)
				assert.Nil(t, err)
				err = objStore.Read(ctx, &readObject)
				assert.Nil(t, err)

				fmt.Printf("%+v\n", readObject)
			}
			{
				readObject := nested{
					NestedInt: 15,
				}
				err = objStore.Read(ctx, &readObject)
				assert.Nil(t, err)
				err = objStore.Read(ctx, &readObject)
				assert.Nil(t, err)

				fmt.Printf("%+v\n", readObject)
			}

			{
				readObject := nested{}
				err = objStore.Read(ctx, &readObject)
				assert.ErrorIs(t, err, redisobj.ErrObjectNotFound)

				fmt.Printf("%+v\n", readObject)
			}
			t.Fail()
		})
	}
}

// func Test_Store_Value(t *testing.T) {
// 	redisClient := NewGoRedisClient()
// 	redisClient.FlushAll()

// 	type nested struct {
// 		NestedString string
// 		NestedInt    int `redisobj:"key"`
// 	}
// 	type value struct {
// 		Foo    string `redisobj:"key"`
// 		Bar    int
// 		Nested nested
// 	}

// 	objStore := redisobj.NewStore(redisClient)

// 	testCases := []struct {
// 		description          string
// 		writeObject          value
// 		options              []redisobj.Option
// 		expectedReadObject   value
// 		expectedFooTtl       time.Duration
// 		expectedBarTtl       time.Duration
// 		expectedNestedFooTtl time.Duration
// 		expectedNestedBarTtl time.Duration
// 	}{
// 		{
// 			description: "option if exists",
// 			writeObject: value{
// 				Foo: "my_foo",
// 				Bar: 5,
// 				Nested: nested{
// 					NestedString: "nested_foo",
// 					NestedInt:    15,
// 				},
// 			},
// 			options: []redisobj.Option{
// 				redisobj.OptionIfExists,
// 			},
// 			expectedReadObject: value{
// 				Foo: "",
// 				Bar: 0,
// 				Nested: nested{
// 					NestedString: "",
// 					NestedInt:    0,
// 				},
// 			},
// 			expectedFooTtl:       ttlNotExists,
// 			expectedBarTtl:       ttlNotExists,
// 			expectedNestedFooTtl: ttlNotExists,
// 			expectedNestedBarTtl: ttlNotExists,
// 		},
// 		{
// 			description: "write/read object that does not exist in redis",
// 			writeObject: value{
// 				Foo: "my_foo",
// 				Bar: 5,
// 				Nested: nested{
// 					NestedString: "nested_foo",
// 					NestedInt:    15,
// 				},
// 			},
// 			options: []redisobj.Option{},
// 			expectedReadObject: value{
// 				Foo: "my_foo",
// 				Bar: 5,
// 				Nested: nested{
// 					NestedString: "nested_foo",
// 					NestedInt:    15,
// 				},
// 			},
// 			expectedFooTtl:       ttlInfinite,
// 			expectedBarTtl:       ttlInfinite,
// 			expectedNestedFooTtl: ttlInfinite,
// 			expectedNestedBarTtl: ttlInfinite,
// 		},
// 		{
// 			description: "write/read object that already exists in redis",
// 			writeObject: value{
// 				Foo: "my_foo",
// 				Bar: 5,
// 				Nested: nested{
// 					NestedString: "nested_foo",
// 					NestedInt:    15,
// 				},
// 			},
// 			options: []redisobj.Option{},
// 			expectedReadObject: value{
// 				Foo: "my_foo",
// 				Bar: 5,
// 				Nested: nested{
// 					NestedString: "nested_foo",
// 					NestedInt:    15,
// 				},
// 			},
// 			expectedFooTtl:       ttlInfinite,
// 			expectedBarTtl:       ttlInfinite,
// 			expectedNestedFooTtl: ttlInfinite,
// 			expectedNestedBarTtl: ttlInfinite,
// 		},
// 		{
// 			description: "option set ttl",
// 			writeObject: value{
// 				Foo: "my_foo",
// 				Bar: 5,
// 				Nested: nested{
// 					NestedString: "nested_foo",
// 					NestedInt:    15,
// 				},
// 			},
// 			options: []redisobj.Option{
// 				redisobj.OptionTtl(10 * time.Minute),
// 			},
// 			expectedReadObject: value{
// 				Foo: "my_foo",
// 				Bar: 5,
// 				Nested: nested{
// 					NestedString: "nested_foo",
// 					NestedInt:    15,
// 				},
// 			},
// 			expectedFooTtl:       10 * time.Minute,
// 			expectedBarTtl:       10 * time.Minute,
// 			expectedNestedFooTtl: 10 * time.Minute,
// 			expectedNestedBarTtl: 10 * time.Minute,
// 		},
// 		{
// 			description: "option if not exists (already exists so write should not occur)",
// 			writeObject: value{
// 				Foo: "my_foo",
// 				Bar: 999,
// 				Nested: nested{
// 					NestedString: "OVERRIDE",
// 					NestedInt:    15,
// 				},
// 			},
// 			options: []redisobj.Option{
// 				redisobj.OptionIfNotExists,
// 			},
// 			expectedReadObject: value{
// 				Foo: "my_foo",
// 				Bar: 5,
// 				Nested: nested{
// 					NestedString: "nested_foo",
// 					NestedInt:    15,
// 				},
// 			},
// 			expectedFooTtl:       10 * time.Minute,
// 			expectedBarTtl:       10 * time.Minute,
// 			expectedNestedFooTtl: 10 * time.Minute,
// 			expectedNestedBarTtl: 10 * time.Minute,
// 		},
// 	}
// 	for _, testCase := range testCases {
// 		t.Run(testCase.description, func(t *testing.T) {
// 			var err error

// 			err = objStore.Write(testCase.writeObject, testCase.options...)
// 			assert.Nil(t, err)

// 			readObject := value{
// 				Foo: testCase.expectedReadObject.Foo,
// 				Nested: nested{
// 					NestedInt: testCase.expectedReadObject.Nested.NestedInt,
// 				},
// 			}
// 			err = objStore.Read(&readObject)
// 			assert.Nil(t, err)

// 			assert.Equal(t, testCase.expectedReadObject, readObject)

// 			actualTtl, err := redisClient.TTL("redisobj:value:my_foo:foo").Result()
// 			assert.Nil(t, err)
// 			assert.Equal(t, testCase.expectedFooTtl, actualTtl)
// 			actualTtl, err = redisClient.TTL("redisobj:value:my_foo:bar").Result()
// 			assert.Nil(t, err)
// 			assert.Equal(t, testCase.expectedBarTtl, actualTtl)
// 			actualTtl, err = redisClient.TTL("redisobj:value:my_foo:nested:15:foo").Result()
// 			assert.Nil(t, err)
// 			assert.Equal(t, testCase.expectedNestedFooTtl, actualTtl)
// 			actualTtl, err = redisClient.TTL("redisobj:value:my_foo:nested:15:bar").Result()
// 			assert.Nil(t, err)
// 			assert.Equal(t, testCase.expectedNestedBarTtl, actualTtl)
// 		})
// 	}
// }

// func Test_Store_Hash(t *testing.T) {
// 	redisClient := NewGoRedisClient()
// 	redisClient.FlushAll()

// 	type nested struct {
// 		NestedString string `redisHash:"foo"`
// 		NestedInt    int    `redisHash:"bar,key"`
// 	}
// 	type value struct {
// 		Foo    string      `redisHash:"foo,key"`
// 		Bar    int         `redisHash:"bar"`
// 		Har    map[int]int `redisHash:"har"`
// 		Nested nested
// 	}

// 	objStore := redisobj.NewStore(redisClient)

// 	testCases := []struct {
// 		description        string
// 		writeObject        value
// 		options            []redisobj.Option
// 		expectedReadObject value
// 		expectedValueTtl   time.Duration
// 		expectedHarTtl     time.Duration
// 		expectedNestedTtl  time.Duration
// 	}{
// 		{
// 			description: "option if exists (command not available for HSET so this will set the value)",
// 			writeObject: value{
// 				Foo: "my_foo",
// 				Bar: 5,
// 				Har: map[int]int{
// 					1: 2,
// 					3: 4,
// 					5: 6,
// 				},
// 				Nested: nested{
// 					NestedString: "nested_foo",
// 					NestedInt:    15,
// 				},
// 			},
// 			options: []redisobj.Option{
// 				redisobj.OptionIfExists,
// 			},
// 			expectedReadObject: value{
// 				Foo: "my_foo",
// 				Bar: 5,
// 				Har: map[int]int{
// 					1: 2,
// 					3: 4,
// 					5: 6,
// 				},
// 				Nested: nested{
// 					NestedString: "nested_foo",
// 					NestedInt:    15,
// 				},
// 			},
// 			expectedValueTtl:  ttlInfinite,
// 			expectedHarTtl:    ttlInfinite,
// 			expectedNestedTtl: ttlInfinite,
// 		},
// 		{
// 			description: "write/read object that does not exist in redis",
// 			writeObject: value{
// 				Foo: "my_foo",
// 				Bar: 5,
// 				Har: map[int]int{
// 					1: 2,
// 					3: 4,
// 					5: 6,
// 				},
// 				Nested: nested{
// 					NestedString: "nested_foo",
// 					NestedInt:    15,
// 				},
// 			},
// 			options: []redisobj.Option{},
// 			expectedReadObject: value{
// 				Foo: "my_foo",
// 				Bar: 5,
// 				Har: map[int]int{
// 					1: 2,
// 					3: 4,
// 					5: 6,
// 				},
// 				Nested: nested{
// 					NestedString: "nested_foo",
// 					NestedInt:    15,
// 				},
// 			},
// 			expectedValueTtl:  ttlInfinite,
// 			expectedHarTtl:    ttlInfinite,
// 			expectedNestedTtl: ttlInfinite,
// 		},
// 		{
// 			description: "write/read object that already exists in redis",
// 			writeObject: value{
// 				Foo: "my_foo",
// 				Bar: 5,
// 				Har: map[int]int{
// 					1: 2,
// 					3: 4,
// 					5: 6,
// 				},
// 				Nested: nested{
// 					NestedString: "nested_foo",
// 					NestedInt:    15,
// 				},
// 			},
// 			options: []redisobj.Option{},
// 			expectedReadObject: value{
// 				Foo: "my_foo",
// 				Bar: 5,
// 				Har: map[int]int{
// 					1: 2,
// 					3: 4,
// 					5: 6,
// 				},
// 				Nested: nested{
// 					NestedString: "nested_foo",
// 					NestedInt:    15,
// 				},
// 			},
// 			expectedValueTtl:  ttlInfinite,
// 			expectedHarTtl:    ttlInfinite,
// 			expectedNestedTtl: ttlInfinite,
// 		},
// 		{
// 			description: "option set ttl",
// 			writeObject: value{
// 				Foo: "my_foo",
// 				Bar: 5,
// 				Har: map[int]int{
// 					1: 2,
// 					3: 4,
// 					5: 6,
// 				},
// 				Nested: nested{
// 					NestedString: "nested_foo",
// 					NestedInt:    15,
// 				},
// 			},
// 			options: []redisobj.Option{
// 				redisobj.OptionTtl(10 * time.Minute),
// 			},
// 			expectedReadObject: value{
// 				Foo: "my_foo",
// 				Bar: 5,
// 				Har: map[int]int{
// 					1: 2,
// 					3: 4,
// 					5: 6,
// 				},
// 				Nested: nested{
// 					NestedString: "nested_foo",
// 					NestedInt:    15,
// 				},
// 			},
// 			expectedValueTtl:  10 * time.Minute,
// 			expectedHarTtl:    10 * time.Minute,
// 			expectedNestedTtl: 10 * time.Minute,
// 		},
// 		{
// 			description: "option if not exists (already exists so write should not occur)",
// 			writeObject: value{
// 				Foo: "my_foo",
// 				Bar: 999,
// 				Har: map[int]int{
// 					999: 999,
// 				},
// 				Nested: nested{
// 					NestedString: "OVERRIDE",
// 					NestedInt:    15,
// 				},
// 			},
// 			options: []redisobj.Option{
// 				redisobj.OptionIfNotExists,
// 			},
// 			expectedReadObject: value{
// 				Foo: "my_foo",
// 				Bar: 5,
// 				Har: map[int]int{
// 					1: 2,
// 					3: 4,
// 					5: 6,
// 				},
// 				Nested: nested{
// 					NestedString: "nested_foo",
// 					NestedInt:    15,
// 				},
// 			},
// 			expectedValueTtl:  10 * time.Minute,
// 			expectedHarTtl:    10 * time.Minute,
// 			expectedNestedTtl: 10 * time.Minute,
// 		},
// 	}
// 	for _, testCase := range testCases {
// 		t.Run(testCase.description, func(t *testing.T) {
// 			var err error

// 			err = objStore.Write(testCase.writeObject, testCase.options...)
// 			assert.Nil(t, err)

// 			readObject := value{
// 				Foo: testCase.expectedReadObject.Foo,
// 				Nested: nested{
// 					NestedInt: testCase.expectedReadObject.Nested.NestedInt,
// 				},
// 			}
// 			err = objStore.Read(&readObject)
// 			assert.Nil(t, err)

// 			assert.Equal(t, testCase.expectedReadObject, readObject)

// 			actualTtl, err := redisClient.TTL("redisobj:value:my_foo").Result()
// 			assert.Nil(t, err)
// 			assert.Equal(t, testCase.expectedValueTtl, actualTtl)

// 			actualTtl, err = redisClient.TTL("redisobj:value:my_foo:har").Result()
// 			assert.Nil(t, err)
// 			assert.Equal(t, testCase.expectedHarTtl, actualTtl)

// 			actualTtl, err = redisClient.TTL("redisobj:value:my_foo:nested:15").Result()
// 			assert.Nil(t, err)
// 			assert.Equal(t, testCase.expectedNestedTtl, actualTtl)
// 		})
// 	}
// }
