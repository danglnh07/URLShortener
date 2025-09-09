package service

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncodeDecode(t *testing.T) {
	cases := 10
	for range cases {
		// Test data
		data := rand.Int63n(10000)

		// Encode the number
		encode := EncodeBase62(data)
		require.NotEmpty(t, encode)

		// Decode and compare
		require.Equal(t, data, DecodeBase62(encode))
	}
}
