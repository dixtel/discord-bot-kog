package twmap

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFixMapName(t *testing.T) {
	require.Equal(t, FixMapName("abc 123"), "abc123")
	require.Equal(t, FixMapName("abc ABC 123"), "abcABC123")
	require.Equal(t, FixMapName("ćą ća 1 22"), "a122")
	require.Equal(t, FixMapName("ab _de"), "ab_de")
	require.Equal(t, FixMapName("ab -de"), "ab-de")
	require.Equal(t, FixMapName("ab .de"), "ab.de")
}