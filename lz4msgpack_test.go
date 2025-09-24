package lz4msgpack_test

import (
	"reflect"
	"testing"

	"github.com/d-o-n-u-t-s/lz4msgpack"
	"github.com/shamaton/msgpack"
)

func Test(t *testing.T) {
	type Data struct {
		A int
		B float32
		C []string
	}

	data := Data{
		A: 4578234323,
		B: 1.46437485,
		C: []string{"Hello World", "Hello World", "Hello World", "Hello World", "Hello World"},
	}
	t.Log(data)

	tester := func(name string, marshaler func(v interface{}) ([]byte, error), unmarshaler func(data []byte, v interface{}) error) {
		b, _ := marshaler(&data)
		t.Logf("%s: %d", name, len(b))
		var data1 Data
		unmarshaler(b, &data1)
		if !reflect.DeepEqual(data, data1) {
			t.Fatal(name + " Error")
		}
	}

	tester("          msgpack.Marshal", msgpack.Encode, msgpack.Decode)
	tester("       lz4msgpack.Marshal", lz4msgpack.Marshal, lz4msgpack.Unmarshal)
	tester("lz4msgpack.MarshalAsArray", lz4msgpack.MarshalAsArray, lz4msgpack.Unmarshal)
}
