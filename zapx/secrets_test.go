package zapx

import (
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

// fakeBody is the fake secret part of the service keys used in tests. Test keys are
// always built at runtime as `prefix + fakeBody` so that no literal in this file looks
// like a real credential to secret scanners.
const fakeBody = "0123456789abcdefghijklmnopqrstuvWXYZ"

func TestSecretString_String(t *testing.T) {
	tests := []struct {
		name string
		s    SecretString
		want string
	}{
		{"empty", "", ""},
		{"short", "my", "***"},
		{"11 chars, below threshold", "mysecret123", "***"},
		{"12 chars, suffix shown", "mysecret1234", "***1234"},
		{"long", "mysecretisverylongandshouldbehidden", "***dden"},
		{"prefix is not kept", SecretString("ac_fake_" + fakeBody), "***WXYZ"},
		{"passphrase", "blue_horse_battery_staple", "***aple"},
		{"multi-byte suffix is not split", "密码密码密码密码密码密码", "***密码密码"},
		{"multi-byte counted as characters, below threshold", "motdepasseé", "***"},
		{"multi-byte counted as characters, at threshold", "motdepasseéé", "***seéé"},
		{"4 multi-byte characters stay masked", "密码密码", "***"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.s.String()

			assert.Equal(t, tt.want, got)
			assert.True(t, utf8.ValidString(got), "output must be valid UTF-8")
		})
	}
}

func TestSecretPrefixedString_String(t *testing.T) {
	tests := []struct {
		name string
		s    SecretPrefixedString
		want string
	}{
		{"empty", "", ""},
		{"short", "my", "***"},
		{"no prefix, below threshold", "mysecret123", "***"},
		{"no prefix", "mysecretisverylongandshouldbehidden", "***dden"},
		{"two segments prefix", SecretPrefixedString("ac_fake_" + fakeBody), "ac_fake_***WXYZ"},
		{"prefix kept with exactly 12 remaining", "ac_fake_abcdefgh1234", "ac_fake_***1234"},
		{"prefix dropped with 11 remaining", "ac_fake_abcdefg1234", "***1234"},
		{"prefix dropped with short secret", "ac_fake_abc", "***"},
		{"prefix only", "ac_fake_", "***"},
		{"almost all prefix", "ab_cdefg_k", "***"},
		{"short words password drops prefix", "my_pass_is_hunt", "***hunt"},
		{"only first two segments kept", "a_b_c_abcdefghijklmnop", "a_b_***mnop"},
		{"segment longer than 5 letters is not a prefix", "abcdef_abcdefghijklmnop", "***mnop"},
		{"second segment too long keeps first", "sk_abcdef_abcdefghijklmnop", "sk_***mnop"},
		{"digit stops prefix", "m3_fam!l!yHas_good_value", "***alue"},
		{"digits are not a prefix", "sk1_abcdefghijklmnop", "***mnop"},
		{"leading underscore is not a prefix", "_abcdefghijklmnop", "***mnop"},
		{"multi-byte secret after prefix", "ab_cd_密码密码密码密码密码密码", "ab_cd_***密码密码"},
		{"multi-byte secret after prefix below threshold falls back to strict", "ab_cd_密码密码密码", "***密码密码"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.s.String()

			assert.Equal(t, tt.want, got)
			assert.True(t, utf8.ValidString(got), "output must be valid UTF-8")
		})
	}
}

// TestSecretPrefixedString_WellKnownServices checks the redaction of the key formats
// of well-known services. The plain [SecretString] always gives `***WXYZ` for them.
func TestSecretPrefixedString_WellKnownServices(t *testing.T) {
	tests := []struct {
		service string
		prefix  string
		want    string
	}{
		// Stripe
		{"Stripe secret key live", "sk_live_", "sk_live_***WXYZ"},
		{"Stripe secret key test", "sk_test_", "sk_test_***WXYZ"},
		{"Stripe publishable key live", "pk_live_", "pk_live_***WXYZ"},
		{"Stripe publishable key test", "pk_test_", "pk_test_***WXYZ"},
		{"Stripe restricted key live", "rk_live_", "rk_live_***WXYZ"},
		{"Stripe restricted key test", "rk_test_", "rk_test_***WXYZ"},
		{"Stripe webhook secret", "whsec_", "whsec_***WXYZ"},

		// GitHub
		{"GitHub personal access token (classic)", "ghp_", "ghp_***WXYZ"},
		{"GitHub OAuth access token", "gho_", "gho_***WXYZ"},
		{"GitHub user-to-server token", "ghu_", "ghu_***WXYZ"},
		{"GitHub server-to-server token", "ghs_", "ghs_***WXYZ"},
		{"GitHub refresh token", "ghr_", "ghr_***WXYZ"},
		// `github` is longer than 5 letters, the prefix is not recognized.
		{"GitHub fine-grained personal access token", "github_pat_", "***WXYZ"},

		// Other `_` separated formats
		{"npm access token", "npm_", "npm_***WXYZ"},
		{"Hugging Face token", "hf_", "hf_***WXYZ"},
		{"Shopify access token", "shpat_", "shpat_***WXYZ"},
		{"Shopify shared secret", "shpss_", "shpss_***WXYZ"},
		{"Linear API key", "lin_api_", "lin_api_***WXYZ"},
		{"Supabase access token", "sbp_", "sbp_***WXYZ"},
		// `v1` contains a digit, only the first segment is recognized.
		{"DigitalOcean personal access token", "dop_v1_", "dop_***WXYZ"},

		// `-` separated formats are not recognized.
		{"Slack bot token", "xoxb-", "***WXYZ"},
		{"OpenAI project key", "sk-proj-", "***WXYZ"},
		{"Anthropic API key", "sk-ant-api03-", "***WXYZ"},
		{"GitLab personal access token", "glpat-", "***WXYZ"},
		{"PyPI token", "pypi-", "***WXYZ"},

		// Formats without separator are not recognized.
		{"AWS access key id", "AKIA", "***WXYZ"},
		{"Google API key", "AIza", "***WXYZ"},
	}
	for _, tt := range tests {
		t.Run(tt.service, func(t *testing.T) {
			value := tt.prefix + fakeBody

			assert.Equal(t, tt.want, SecretPrefixedString(value).String())
			assert.Equal(t, "***WXYZ", SecretString(value).String())
		})
	}
}

func TestSecretFields(t *testing.T) {
	value := "ac_fake_" + fakeBody

	assert.Equal(t, "***WXYZ", Secret("key", value).String)
	assert.Equal(t, "ac_fake_***WXYZ", SecretPrefixed("key", value).String)
}
