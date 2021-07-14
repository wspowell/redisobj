package redisobj

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testObj struct {
	MyValueString string `redisValue:"my_string_value,10s,nx,keepttl,get"`
}

func Test_parseObject(t *testing.T) {
	test := &testObj{
		MyValueString: "test",
	}
	objRef, err := newObjStruct(test)
	assert.Nil(t, err)

	fmt.Printf("%+v\n", objRef)

	t.Fail()
}
