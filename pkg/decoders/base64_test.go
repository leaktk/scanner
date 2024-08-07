package decoders

import (
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecodeBase64(t *testing.T) {
	tests := []struct {
		chunk    []byte
		expected []byte
		name     string
	}{
		{
			name:     "only b64 chunk",
			chunk:    []byte(`bG9uZ2VyLWVuY29kZWQtc2VjcmV0LXRlc3Q=`),
			expected: []byte(`longer-encoded-secret-test`),
		},
		{
			name:     "mixed content",
			chunk:    []byte(`token: bG9uZ2VyLWVuY29kZWQtc2VjcmV0LXRlc3Q=`),
			expected: []byte(`token: longer-encoded-secret-test`),
		},
		{
			name:     "no chunk",
			chunk:    []byte(``),
			expected: nil,
		},
		{
			name:     "env var (looks like all b64 decodable but has `=` in the middle)",
			chunk:    []byte(`some-encoded-secret=dGVzdHNlY3JldA==`), // notsecret
			expected: []byte(`some-encoded-secret=testsecret`),       // notsecret
		},
		{
			name:     "has longer b64 inside",
			chunk:    []byte(`some-encoded-secret="bG9uZ2VyLWVuY29kZWQtc2VjcmV0LXRlc3Q="`), // notsecret
			expected: []byte(`some-encoded-secret="longer-encoded-secret-test"`),           // notsecret
		},
		{
			name: "many possible substrings",
			chunk: []byte(`Many substrings in this slack message could be base64 decoded
				but only dGhpcyBlbmNhcHN1bGF0ZWQgc2VjcmV0 should be decoded.`),
			expected: []byte(`Many substrings in this slack message could be base64 decoded
				but only this encapsulated secret should be decoded.`),
		},
		{
			name:     "b64-url-safe: only b64 chunk",
			chunk:    []byte(`bG9uZ2VyLWVuY29kZWQtc2VjcmV0LXRlc3Q`),
			expected: []byte(`longer-encoded-secret-test`),
		},
		{
			name:     "b64-url-safe: mixed content",
			chunk:    []byte(`token: bG9uZ2VyLWVuY29kZWQtc2VjcmV0LXRlc3Q`),
			expected: []byte(`token: longer-encoded-secret-test`),
		},
		{
			name:     "b64-url-safe: env var (looks like all b64 decodable but has `=` in the middle)",
			chunk:    []byte(`some-encoded-secret=dGVzdHNlY3JldA`), // notsecret
			expected: []byte(`some-encoded-secret=testsecret`),     // notsecret
		},
		{
			name:     "b64-url-safe: has longer b64 inside",
			chunk:    []byte(`some-encoded-secret="bG9uZ2VyLWVuY29kZWQtc2VjcmV0LXRlc3Q"`), // notsecret
			expected: []byte(`some-encoded-secret="longer-encoded-secret-test"`),          // notsecret
		},
		{
			name:     "b64-url-safe: hyphen url b64",
			chunk:    []byte(`dHJ1ZmZsZWhvZz4-ZmluZHMtc2VjcmV0cw`),
			expected: []byte(`trufflehog>>finds-secrets`),
		},
		{
			name:     "b64-url-safe: underscore url b64",
			chunk:    []byte(`YjY0dXJsc2FmZS10ZXN0LXNlY3JldC11bmRlcnNjb3Jlcz8_`),
			expected: []byte(`b64urlsafe-test-secret-underscores??`),
		},
		{
			name:     "invalid base64 string",
			chunk:    []byte(`a3d3fa7c2bb99e469ba55e5834ce79ee4853a8a3`),
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, string(tt.expected), string(DecodeBase64(tt.chunk)))
		})
	}
}

func MustGetBenchmarkData() map[string][]byte {
	sizes := map[string]int{
		"xsmall":  10,          // 10 bytes
		"small":   100,         // 100 bytes
		"medium":  1024,        // 1KB
		"large":   10 * 1024,   // 10KB
		"xlarge":  100 * 1024,  // 100KB
		"xxlarge": 1024 * 1024, // 1MB
	}
	data := make(map[string][]byte)

	for key, size := range sizes {
		// Generating a byte slice of a specific size with random data.
		content := make([]byte, size)
		for i := 0; i < size; i++ {
			randomByte, err := rand.Int(rand.Reader, big.NewInt(256))
			if err != nil {
				panic(err)
			}
			content[i] = byte(randomByte.Int64())
		}
		data[key] = content
	}

	return data
}

func BenchmarkFromChunkSmall(b *testing.B) {
	data := MustGetBenchmarkData()["small"]

	for n := 0; n < b.N; n++ {
		DecodeBase64(data)
	}
}

func BenchmarkFromChunkMedium(b *testing.B) {
	data := MustGetBenchmarkData()["medium"]

	for n := 0; n < b.N; n++ {
		DecodeBase64(data)
	}
}

func BenchmarkFromChunkLarge(b *testing.B) {
	data := MustGetBenchmarkData()["big"]

	for n := 0; n < b.N; n++ {
		DecodeBase64(data)
	}
}
