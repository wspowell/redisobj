package redisobj_test

import (
	"context"
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

func Test_Store_singleton(t *testing.T) {
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
		String string
		Int    int
		Map    map[int]int
		Slice  []string
		Nested nested
	}

	objStore := redisobj.NewStore(redisClient)

	testObject := &root{
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

	actualObject := &root{}

	testCases := []struct {
		description    string
		object         *root
		options        redisobj.Options
		expectedObject *root
		expectedTtl    time.Duration
		expectedError  error
	}{
		{
			description:    "object does not exist",
			object:         nil,
			options:        redisobj.Options{},
			expectedObject: &root{},
			expectedTtl:    ttlNotExists,
			expectedError:  redisobj.ErrObjectNotFound,
		},
		{
			description: "writes and reads object successfully",
			object:      testObject,
			options: redisobj.Options{
				Ttl: time.Minute,
			},
			expectedObject: testObject,
			expectedTtl:    time.Minute,
			expectedError:  nil,
		},
		{
			description: "writes and reads object successfully - object cached",
			object:      testObject,
			options: redisobj.Options{
				EnableCaching: true,
			},
			expectedObject: testObject,
			expectedTtl:    ttlInfinite,
			expectedError:  nil,
		},
		{
			description:    "writes and reads object successfully - no ttl",
			object:         testObject,
			options:        redisobj.Options{},
			expectedObject: testObject,
			expectedTtl:    ttlInfinite,
			expectedError:  nil,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			var err error

			err = objStore.Write(ctx, testCase.object, testCase.options)
			if testCase.object == nil {
				assert.ErrorIs(t, err, redisobj.ErrInvalidObject)
			} else {
				assert.Nil(t, err)
			}

			err = objStore.Read(ctx, actualObject, testCase.options)
			assert.ErrorIs(t, err, testCase.expectedError)
			assert.Equal(t, testCase.expectedObject, actualObject)

			var actualTtl time.Duration

			actualTtl, err = redisClient.TTL("{redisobj:root}").Result()
			assert.Nil(t, err)
			assert.Equal(t, testCase.expectedTtl, actualTtl)

			actualTtl, err = redisClient.TTL("{redisobj:root}.__EXISTS__").Result()
			assert.Nil(t, err)
			assert.Equal(t, testCase.expectedTtl, actualTtl)

			if testCase.options.EnableCaching {
				actualTtl, err = redisClient.TTL("{redisobj:root}.__HASH__").Result()
				assert.Nil(t, err)
				assert.Equal(t, testCase.expectedTtl, actualTtl)
			}

			actualTtl, err = redisClient.TTL("{redisobj:root}.Map").Result()
			assert.Nil(t, err)
			assert.Equal(t, testCase.expectedTtl, actualTtl)

			actualTtl, err = redisClient.TTL("{redisobj:root}.Slice").Result()
			assert.Nil(t, err)
			assert.Equal(t, testCase.expectedTtl, actualTtl)
		})
	}
}

func Test_Store_keyed_object(t *testing.T) {
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

	objStore := redisobj.NewStore(redisClient)

	actualObject := &root{}

	testCases := []struct {
		description    string
		object         *root
		options        redisobj.Options
		expectedObject *root
		expectedTtl    time.Duration
		expectedError  error
	}{
		{
			description:    "object does not exist - no key value",
			object:         nil,
			options:        redisobj.Options{},
			expectedObject: &root{},
			expectedTtl:    ttlNotExists,
			expectedError:  redisobj.ErrObjectNotFound,
		},
		{
			description: "object does not exist",
			object:      nil,
			options:     redisobj.Options{},
			expectedObject: &root{
				Id: "UUID",
			},
			expectedTtl:   ttlNotExists,
			expectedError: redisobj.ErrObjectNotFound,
		},
		{
			description: "writes and reads object successfully",
			object: &root{
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
			},
			options: redisobj.Options{
				Ttl: time.Minute,
			},
			expectedObject: &root{
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
			},
			expectedTtl:   time.Minute,
			expectedError: nil,
		},
		{
			description: "writes and reads object successfully - object cached",
			object: &root{
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
			},
			options: redisobj.Options{
				EnableCaching: true,
			},
			expectedObject: &root{
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
			},
			expectedTtl:   ttlInfinite,
			expectedError: nil,
		},
		{
			description: "writes and reads object successfully - no ttl",
			object: &root{
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
			},
			options: redisobj.Options{},
			expectedObject: &root{
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
			},
			expectedTtl:   ttlInfinite,
			expectedError: nil,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			var err error

			err = objStore.Write(ctx, testCase.object, testCase.options)
			if testCase.object == nil {
				assert.ErrorIs(t, err, redisobj.ErrInvalidObject)
			} else {
				assert.Nil(t, err)
			}

			actualObject.Id = testCase.expectedObject.Id
			err = objStore.Read(ctx, actualObject, testCase.options)
			assert.ErrorIs(t, err, testCase.expectedError)
			assert.Equal(t, testCase.expectedObject, actualObject)

			var actualTtl time.Duration

			actualTtl, err = redisClient.TTL("{redisobj:root:UUID}").Result()
			assert.Nil(t, err)
			assert.Equal(t, testCase.expectedTtl, actualTtl)

			actualTtl, err = redisClient.TTL("{redisobj:root:UUID}.__EXISTS__").Result()
			assert.Nil(t, err)
			assert.Equal(t, testCase.expectedTtl, actualTtl)

			if testCase.options.EnableCaching {
				actualTtl, err = redisClient.TTL("{redisobj:root:UUID}.__HASH__").Result()
				assert.Nil(t, err)
				assert.Equal(t, testCase.expectedTtl, actualTtl)
			}

			actualTtl, err = redisClient.TTL("{redisobj:root:UUID}.Map").Result()
			assert.Nil(t, err)
			assert.Equal(t, testCase.expectedTtl, actualTtl)

			actualTtl, err = redisClient.TTL("{redisobj:root:UUID}.Slice").Result()
			assert.Nil(t, err)
			assert.Equal(t, testCase.expectedTtl, actualTtl)
		})
	}
}
