package zapx

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSecretString_String(t *testing.T) {
	tests := []struct {
		name string
		s    SecretString
		want string
	}{
		{
			name: "Empty Secret",
			s:    SecretString(""),
			want: "",
		},
		{
			name: "Shorter than 3",
			s:    SecretString("my"),
			want: "**",
		},
		{
			name: "Equal to 3",
			s:    SecretString("mys"),
			want: "***",
		},
		{
			name: "Shorter than 6",
			s:    SecretString("mysec"),
			want: "m****",
		},
		{
			name: "Equal to 6",
			s:    SecretString("mysecr"),
			want: "m*****",
		},
		{
			name: "Shorter than 9",
			s:    SecretString("mysecret"),
			want: "my******",
		},
		{
			name: "Equal to 9",
			s:    SecretString("mysecreti"),
			want: "my*******",
		},
		{
			name: "Longer than 9",
			s:    SecretString("mysecretisverylongandshouldbehidden"),
			want: "mysecret***************************",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.s.String())
		})
	}
}
