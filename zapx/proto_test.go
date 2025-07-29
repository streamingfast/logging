package zapx

import (
	"strings"
	"testing"
	"time"

	"github.com/lithammer/dedent"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func dedentString(input string) string {
	return strings.TrimSpace(dedent.Dedent(input))
}

func TestProtoStringer_String(t *testing.T) {
	fixedTime := time.Date(2023, 1, 1, 12, 0, 0, 123456789, time.UTC)
	fixedTimestamp := timestamppb.New(fixedTime)

	tests := []struct {
		name      string
		inner     proto.Message
		options   ProtoFieldOptions
		want      string
		wantError string
	}{
		{
			name:    "nil message",
			inner:   nil,
			options: ProtoFieldOptions{},
			want:    "<nil>",
		},
		{
			name:  "text marshaling - compact",
			inner: fixedTimestamp,
			options: ProtoFieldOptions{
				MarshalerKind: ProtoFieldTextMarshaler,
			},
			want: "seconds:1672574400 nanos:123456789",
		},
		{
			name:  "text marshaling - pretty",
			inner: fixedTimestamp,
			options: ProtoFieldOptions{
				MarshalerKind: ProtoFieldTextMarshaler,
				Indent:        "  ",
			},
			want: dedentString(`
				seconds: 1672574400
				nanos: 123456789
			`) + "\n",
		},
		{
			name:  "json marshaling - compact",
			inner: fixedTimestamp,
			options: ProtoFieldOptions{
				MarshalerKind: ProtoFieldJSONMarshaler,
			},
			want: `"2023-01-01T12:00:00.123456789Z"`,
		},
		{
			name:  "json marshaling - pretty",
			inner: fixedTimestamp,
			options: ProtoFieldOptions{
				MarshalerKind: ProtoFieldJSONMarshaler,
				Indent:        "  ",
			},
			want: `"2023-01-01T12:00:00.123456789Z"`,
		},
		{
			name:  "json marshaling - zero timestamp",
			inner: &timestamppb.Timestamp{},
			options: ProtoFieldOptions{
				MarshalerKind: ProtoFieldJSONMarshaler,
				AllowPartial:  true,
			},
			want: `"1970-01-01T00:00:00Z"`,
		},
		{
			name:  "text marshaling - zero timestamp",
			inner: &timestamppb.Timestamp{},
			options: ProtoFieldOptions{
				MarshalerKind: ProtoFieldTextMarshaler,
				AllowPartial:  true,
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := protoStringer{inner: tt.inner, options: tt.options}
			result := s.String()

			if tt.wantError != "" {
				require.Contains(t, result, tt.wantError)
				return
			}

			require.Equal(t, tt.want, result)
		})
	}
}
