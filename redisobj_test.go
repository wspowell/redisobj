package redisobj_test

import (
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

func Test_Store_Value(t *testing.T) {
	redisClient := NewGoRedisClient()
	redisClient.FlushAll()

	type nested struct {
		NestedString string `redisValue:"foo"`
		NestedInt    int    `redisValue:"bar,key"`
	}
	type value struct {
		Foo    string `redisValue:"foo,key"`
		Bar    int    `redisValue:"bar"`
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
			description: "option does not exist",
			writeObject: value{
				Foo: "my_foo",
				Bar: 5,
				Nested: nested{
					NestedString: "nested_foo",
					NestedInt:    15,
				},
			},
			options: []redisobj.Option{
				redisobj.OptionIfExists,
			},
			expectedReadObject: value{
				Foo: "",
				Bar: 0,
				Nested: nested{
					NestedString: "",
					NestedInt:    0,
				},
			},
			expectedFooTtl:       ttlNotExists,
			expectedBarTtl:       ttlNotExists,
			expectedNestedFooTtl: ttlNotExists,
			expectedNestedBarTtl: ttlNotExists,
		},
		{
			description: "write/read object that does not exist in redis",
			writeObject: value{
				Foo: "my_foo",
				Bar: 5,
				Nested: nested{
					NestedString: "nested_foo",
					NestedInt:    15,
				},
			},
			options: []redisobj.Option{},
			expectedReadObject: value{
				Foo: "my_foo",
				Bar: 5,
				Nested: nested{
					NestedString: "nested_foo",
					NestedInt:    15,
				},
			},
			expectedFooTtl:       ttlInfinite,
			expectedBarTtl:       ttlInfinite,
			expectedNestedFooTtl: ttlInfinite,
			expectedNestedBarTtl: ttlInfinite,
		},
		{
			description: "write/read object that already exists in redis",
			writeObject: value{
				Foo: "my_foo",
				Bar: 5,
				Nested: nested{
					NestedString: "nested_foo",
					NestedInt:    15,
				},
			},
			options: []redisobj.Option{},
			expectedReadObject: value{
				Foo: "my_foo",
				Bar: 5,
				Nested: nested{
					NestedString: "nested_foo",
					NestedInt:    15,
				},
			},
			expectedFooTtl:       ttlInfinite,
			expectedBarTtl:       ttlInfinite,
			expectedNestedFooTtl: ttlInfinite,
			expectedNestedBarTtl: ttlInfinite,
		},
		{
			description: "option set ttl",
			writeObject: value{
				Foo: "my_foo",
				Bar: 5,
				Nested: nested{
					NestedString: "nested_foo",
					NestedInt:    15,
				},
			},
			options: []redisobj.Option{
				redisobj.OptionTtl(10 * time.Minute),
			},
			expectedReadObject: value{
				Foo: "my_foo",
				Bar: 5,
				Nested: nested{
					NestedString: "nested_foo",
					NestedInt:    15,
				},
			},
			expectedFooTtl:       10 * time.Minute,
			expectedBarTtl:       10 * time.Minute,
			expectedNestedFooTtl: 10 * time.Minute,
			expectedNestedBarTtl: 10 * time.Minute,
		},
		{
			description: "option if not exists (already exists so write should not occur)",
			writeObject: value{
				Foo: "my_foo",
				Bar: 999,
				Nested: nested{
					NestedString: "OVERRIDE",
					NestedInt:    15,
				},
			},
			options: []redisobj.Option{
				redisobj.OptionIfNotExists,
			},
			expectedReadObject: value{
				Foo: "my_foo",
				Bar: 5,
				Nested: nested{
					NestedString: "nested_foo",
					NestedInt:    15,
				},
			},
			expectedFooTtl:       10 * time.Minute,
			expectedBarTtl:       10 * time.Minute,
			expectedNestedFooTtl: 10 * time.Minute,
			expectedNestedBarTtl: 10 * time.Minute,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			var err error

			err = objStore.Write(testCase.writeObject, testCase.options...)
			assert.Nil(t, err)

			readObject := testCase.expectedReadObject
			err = objStore.Read(&readObject)
			assert.Nil(t, err)

			assert.Equal(t, testCase.expectedReadObject, readObject)

			actualTtl, err := redisClient.TTL("redisobj:value:my_foo:foo").Result()
			assert.Nil(t, err)
			assert.Equal(t, testCase.expectedFooTtl, actualTtl)
			actualTtl, err = redisClient.TTL("redisobj:value:my_foo:bar").Result()
			assert.Nil(t, err)
			assert.Equal(t, testCase.expectedBarTtl, actualTtl)
			actualTtl, err = redisClient.TTL("redisobj:value:my_foo:nested:15:foo").Result()
			assert.Nil(t, err)
			assert.Equal(t, testCase.expectedNestedFooTtl, actualTtl)
			actualTtl, err = redisClient.TTL("redisobj:value:my_foo:nested:15:bar").Result()
			assert.Nil(t, err)
			assert.Equal(t, testCase.expectedNestedBarTtl, actualTtl)
		})
	}

}
