package lz4msgpack_test

import (
	"encoding/binary"
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/d-o-n-u-t-s/lz4msgpack"
	"github.com/pierrec/lz4/v4"
	"github.com/shamaton/msgpack/v2"
)

func mapsEqual(a, b map[string]interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if bv, ok := b[k]; !ok || !valuesEqual(v, bv) {
			return false
		}
	}
	return true
}

func valuesEqual(a, b interface{}) bool {
	if reflect.DeepEqual(a, b) {
		return true
	}
	
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)
	
	if av.Kind() != bv.Kind() {
		if (av.Kind() >= reflect.Int && av.Kind() <= reflect.Int64) &&
		   (bv.Kind() >= reflect.Int && bv.Kind() <= reflect.Int64 ||
		    bv.Kind() >= reflect.Uint && bv.Kind() <= reflect.Uint64) {
			return av.Int() == int64(bv.Uint()) || av.Int() == bv.Int()
		}
		if (av.Kind() >= reflect.Uint && av.Kind() <= reflect.Uint64) &&
		   (bv.Kind() >= reflect.Int && bv.Kind() <= reflect.Int64 ||
		    bv.Kind() >= reflect.Uint && bv.Kind() <= reflect.Uint64) {
			return av.Uint() == bv.Uint() || int64(av.Uint()) == bv.Int()
		}
		if av.Kind() == reflect.Slice && bv.Kind() == reflect.Slice {
			if av.Len() != bv.Len() {
				return false
			}
			for i := 0; i < av.Len(); i++ {
				if !valuesEqual(av.Index(i).Interface(), bv.Index(i).Interface()) {
					return false
				}
			}
			return true
		}
		if av.Kind() == reflect.Map && bv.Kind() == reflect.Map {
			if av.Len() != bv.Len() {
				return false
			}
			for _, key := range av.MapKeys() {
				aval := av.MapIndex(key).Interface()
				bval := bv.MapIndex(key).Interface()
				if !valuesEqual(aval, bval) {
					return false
				}
			}
			return true
		}
	}
	
	return false
}

func Test(t *testing.T) {
	type Data struct {
		A int
		B int8
		C int16
		D int32
		E int64
		F uint
		G uint8
		H uint16
		I uint32
		J uint64
		// K uintptr // unsupported
		L float32
		M float64
		N []string
		O time.Time
		P []rune
		Q []byte
	}

	data := Data{
		A: 4578234323,
		B: math.MaxInt8,
		C: math.MaxInt16,
		D: math.MaxInt32,
		E: math.MaxInt64,
		F: ^uint(0),
		G: ^uint8(0),
		H: ^uint16(0),
		I: ^uint32(0),
		J: ^uint64(0),
		// K: ^uintptr(0),
		L: math.MaxFloat32,
		M: math.MaxFloat64,
		N: []string{"Hello World", "Hello World", "Hello World", "Hello World", "Hello World"},
		O: time.Date(1999, 12, 31, 7, 7, 7, 77777, time.Local),
		P: []rune("Hello World"),
		Q: []byte("Hello World"),
	}
	t.Log(data)

	tester := func(name string, marshaler func(v interface{}) ([]byte, error), unmarshaler func(data []byte, v interface{}) error) {
		b, err := marshaler(&data)
		if err != nil {
			t.Fatal("marshal", err)
		}
		t.Logf("%s: %d", name, len(b))
		var data1 Data
		if err = unmarshaler(b, &data1); err != nil {
			t.Fatal("unmarshal", err)
		}
		if !reflect.DeepEqual(data, data1) {
			t.Fatal("error", name)
		}
	}

	tester("          msgpack.Marshal", msgpack.Marshal, msgpack.Unmarshal)
	tester("   msgpack.MarshalAsArray", msgpack.MarshalAsArray, msgpack.UnmarshalAsArray)
	tester("       lz4msgpack.Marshal", lz4msgpack.Marshal, lz4msgpack.Unmarshal)
	tester("lz4msgpack.MarshalAsArray", lz4msgpack.MarshalAsArray, lz4msgpack.UnmarshalAsArray)
}

func TestMarshal(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		wantErr  bool
	}{
		{
			name:    "string",
			input:   "hello world",
			wantErr: false,
		},
		{
			name:    "int",
			input:   42,
			wantErr: false,
		},
		{
			name:    "slice of strings",
			input:   []string{"a", "b", "c"},
			wantErr: false,
		},
		{
			name:    "map",
			input:   map[string]int{"key1": 1, "key2": 2},
			wantErr: false,
		},
		{
			name:    "struct as map",
			input:   map[string]interface{}{"Name": "John", "Age": 30},
			wantErr: false,
		},
		{
			name:    "nil",
			input:   nil,
			wantErr: false,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: false,
		},
		{
			name:    "empty slice",
			input:   []string{},
			wantErr: false,
		},
		{
			name:    "large data",
			input:   make([]byte, 10000),
			wantErr: false,
		},
		{
			name:    "boolean",
			input:   true,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := lz4msgpack.Marshal(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Marshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			if len(data) == 0 {
				t.Error("Marshal() returned empty data")
				return
			}

			switch v := tt.input.(type) {
			case string:
				var result string
				if err := lz4msgpack.Unmarshal(data, &result); err != nil {
					t.Errorf("Failed to unmarshal: %v", err)
					return
				}
				if result != v {
					t.Errorf("Round-trip failed: input %v, got %v", v, result)
				}
			case int:
				var result int
				if err := lz4msgpack.Unmarshal(data, &result); err != nil {
					t.Errorf("Failed to unmarshal: %v", err)
					return
				}
				if result != v {
					t.Errorf("Round-trip failed: input %v, got %v", v, result)
				}
			case []string:
				var result []string
				if err := lz4msgpack.Unmarshal(data, &result); err != nil {
					t.Errorf("Failed to unmarshal: %v", err)
					return
				}
				if !reflect.DeepEqual(result, v) {
					t.Errorf("Round-trip failed: input %v, got %v", v, result)
				}
			case map[string]int:
				var result map[string]int
				if err := lz4msgpack.Unmarshal(data, &result); err != nil {
					t.Errorf("Failed to unmarshal: %v", err)
					return
				}
				if !reflect.DeepEqual(result, v) {
					t.Errorf("Round-trip failed: input %v, got %v", v, result)
				}
			case map[string]interface{}:
				var result map[string]interface{}
				if err := lz4msgpack.Unmarshal(data, &result); err != nil {
					t.Errorf("Failed to unmarshal: %v", err)
					return
				}
				if !mapsEqual(v, result) {
					t.Errorf("Round-trip failed: input %v, got %v", v, result)
				}
			case bool:
				var result bool
				if err := lz4msgpack.Unmarshal(data, &result); err != nil {
					t.Errorf("Failed to unmarshal: %v", err)
					return
				}
				if result != v {
					t.Errorf("Round-trip failed: input %v, got %v", v, result)
				}
			case []byte:
				var result []byte
				if err := lz4msgpack.Unmarshal(data, &result); err != nil {
					t.Errorf("Failed to unmarshal: %v", err)
					return
				}
				if !reflect.DeepEqual(result, v) {
					t.Errorf("Round-trip failed: input %v, got %v", v, result)
				}
			default:
				var result interface{}
				if err := lz4msgpack.Unmarshal(data, &result); err != nil {
					t.Errorf("Failed to unmarshal: %v", err)
					return
				}
				if !reflect.DeepEqual(result, tt.input) {
					t.Errorf("Round-trip failed: input %v, got %v", tt.input, result)
				}
			}
		})
	}
}

func TestMarshalErrorHandling(t *testing.T) {
	unsupportedInput := make(chan int)
	_, err := lz4msgpack.Marshal(unsupportedInput)
	if err == nil {
		t.Error("Marshal() should return error for unsupported type")
	}
}

func TestMarshalCompression(t *testing.T) {
	largeData := make([]string, 1000)
	for i := range largeData {
		largeData[i] = "This is a repeated string that should compress well with LZ4"
	}

	compressed, err := lz4msgpack.Marshal(largeData)
	if err != nil {
		t.Fatalf("Marshal() failed: %v", err)
	}

	uncompressed, err := msgpack.Marshal(largeData)
	if err != nil {
		t.Fatalf("msgpack.Marshal() failed: %v", err)
	}

	if len(compressed) >= len(uncompressed) {
		t.Logf("Compression may not be effective: compressed=%d, uncompressed=%d", len(compressed), len(uncompressed))
	} else {
		t.Logf("Compression effective: compressed=%d, uncompressed=%d, ratio=%.2f%%", 
			len(compressed), len(uncompressed), float64(len(compressed))/float64(len(uncompressed))*100)
	}
}

func TestExtUnmarshal(t *testing.T) {
	data := "qwertyuioppasdfghjkl;'zxcvbnm,./"
	msgpackData, err := msgpack.Marshal(data)
	if err != nil {
		t.Fatal("msgpack", err)
	}
	msgpackLength := len(msgpackData)

	lz4Data := make([]byte, lz4.CompressBlockBound(msgpackLength))
	lz4Length, _ := lz4.CompressBlockHC(msgpackData, lz4Data, 0, nil, nil)
	if err != nil {
		t.Fatal("lz4", err)
	}

	// ext8
	ext8 := []byte{0xc7, 0, 99, 0xd2}
	ext8 = binary.BigEndian.AppendUint32(ext8, uint32(msgpackLength))
	ext8 = append(ext8, lz4Data...)
	var ext8umarshaled string
	if err = lz4msgpack.Unmarshal(ext8[:8+lz4Length], &ext8umarshaled); err != nil {
		t.Fatal("unmarshal ext8", err)
	}
	if !reflect.DeepEqual(data, ext8umarshaled) {
		t.Fatal("error ext8")
	}

	// ext16
	ext16 := []byte{0xc8, 0, 0, 99, 0xd2}
	ext16 = binary.BigEndian.AppendUint32(ext16, uint32(msgpackLength))
	ext16 = append(ext16, lz4Data...)
	var ext16umarshaled string
	if err = lz4msgpack.Unmarshal(ext16[:9+lz4Length], &ext16umarshaled); err != nil {
		t.Fatal("unmarshal ext16", err)
	}
	if !reflect.DeepEqual(data, ext16umarshaled) {
		t.Fatal("error ext16")
	}

	// ext32
	ext32 := []byte{0xc9, 0, 0, 0, 0, 99, 0xd2}
	ext32 = binary.BigEndian.AppendUint32(ext32, uint32(msgpackLength))
	ext32 = append(ext32, lz4Data...)
	var ext32umarshaled string
	if err = lz4msgpack.Unmarshal(ext32[:11+lz4Length], &ext32umarshaled); err != nil {
		t.Fatal("unmarshal ext32", err)
	}
	if !reflect.DeepEqual(data, ext32umarshaled) {
		t.Fatal("error ext32")
	}
}
