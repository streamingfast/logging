package zapx

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGoTypeName_String(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  string
	}{
		{
			name:  "String type",
			value: "hello",
			want:  "string",
		},
		{
			name:  "Int type",
			value: 42,
			want:  "int",
		},
		{
			name:  "Bool type",
			value: true,
			want:  "bool",
		},
		{
			name:  "Float64 type",
			value: 3.14,
			want:  "float64",
		},
		{
			name:  "Slice type",
			value: []string{"a", "b"},
			want:  "[]string",
		},
		{
			name:  "Map type",
			value: map[string]int{"key": 1},
			want:  "map[string]int",
		},
		{
			name:  "Pointer type",
			value: &struct{}{},
			want:  "*struct {}",
		},
		{
			name:  "Nil pointer",
			value: (*string)(nil),
			want:  "*string",
		},
		{
			name:  "Interface type",
			value: error(nil),
			want:  "<nil>",
		},
		{
			name:  "Channel type",
			value: make(chan int),
			want:  "chan int",
		},
		{
			name:  "Function type",
			value: func() {},
			want:  "func()",
		},
		{
			name:  "Struct type",
			value: struct{ Name string }{Name: "test"},
			want:  "struct { Name string }",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := typeStringer{inner: tt.value}
			require.Equal(t, tt.want, s.String())
		})
	}
}
