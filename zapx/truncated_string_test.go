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

func TestTruncatedString(t *testing.T) {
	tests := []struct {
		name      string
		fieldName string
		value     string
		maxLength int
		wantKey   string
		wantValue string
	}{
		{
			name:      "Basic truncation",
			fieldName: "message",
			value:     "this is a long message that should be truncated",
			maxLength: 20,
			wantKey:   "message",
			wantValue: "this is a long messa",
		},
		{
			name:      "No truncation needed",
			fieldName: "short",
			value:     "short",
			maxLength: 10,
			wantKey:   "short",
			wantValue: "short",
		},
		{
			name:      "Empty field name",
			fieldName: "",
			value:     "test value",
			maxLength: 5,
			wantKey:   "",
			wantValue: "test ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := TruncatedString(tt.fieldName, tt.value, tt.maxLength)

			// Verify the field key
			require.Equal(t, tt.wantKey, field.Key)

			// The actual string value should be accessible through the Interface field
			// which contains the stringer implementation
			stringer, ok := field.Interface.(truncatedStringField)
			require.True(t, ok, "Field.Interface should be a truncatedStringField")
			require.Equal(t, tt.wantValue, stringer.String())
		})
	}
}

func TestTruncatedString_LazyEvaluation(t *testing.T) {
	// This test verifies that the truncation is performed lazily
	originalValue := "this is a very long string that will be truncated"
	maxLength := 10

	field := TruncatedString("test", originalValue, maxLength)

	// The field should contain the stringer, not the truncated value directly
	stringer, ok := field.Interface.(truncatedStringField)
	require.True(t, ok)

	// The original value should be preserved in the stringer
	require.Equal(t, originalValue, stringer.value)
	require.Equal(t, maxLength, stringer.maxLength)

	// Only when String() is called should the truncation happen
	require.Equal(t, "this is a ", stringer.String())
}
