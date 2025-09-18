package zapx

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTruncatedStringField_String(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		maxLength int
		want      string
	}{
		{
			name:      "Empty string",
			value:     "",
			maxLength: 10,
			want:      "",
		},
		{
			name:      "String shorter than max length",
			value:     "hello",
			maxLength: 10,
			want:      "hello",
		},
		{
			name:      "String equal to max length",
			value:     "hello world",
			maxLength: 11,
			want:      "hello world",
		},
		{
			name:      "String longer than max length",
			value:     "hello world this is a long string",
			maxLength: 11,
			want:      "hello world",
		},
		{
			name:      "Max length zero",
			value:     "hello",
			maxLength: 0,
			want:      "",
		},
		{
			name:      "Max length one",
			value:     "hello",
			maxLength: 1,
			want:      "h",
		},
		{
			name:      "Unicode string truncation",
			value:     "héllo wörld",
			maxLength: 6,
			want:      "héllo",
		},
		{
			name:      "Very long string",
			value:     "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua",
			maxLength: 20,
			want:      "Lorem ipsum dolor si",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := truncatedStringField{
				value:     tt.value,
				maxLength: tt.maxLength,
			}
			require.Equal(t, tt.want, field.String())
		})
	}
}


