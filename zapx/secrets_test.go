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
			name: "Length 5: 25% front=1, 10% end=1",
			s:    SecretString("mysec"),
			want: "m***c",
		},
		{
			name: "Length 10: 25% front=2, 10% end=1",
			s:    SecretString("mysecret12"),
			want: "my*******2",
		},
		{
			name: "Length 19: 25% front=4, 10% end=1",
			s:    SecretString("mysecret12345678901"),
			want: "myse**************1",
		},
		{
			name: "Length 36: 25% front=9, 10% end=3",
			s:    SecretString("mysecretisverylongandshouldbehidden"),
			want: "mysecret************************den",
		},
		{
			name: "API key example (42 chars): 25% front=10, 10% end=4",
			s:    SecretString("ac_fake_abcdefghijklmnopqrstuvwxyz123456"),
			want: "ac_fake_ab**************************3456",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.s.String())
		})
	}
}
