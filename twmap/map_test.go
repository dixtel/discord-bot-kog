package twmap

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsMapNameValid(t *testing.T) {
	require.Equal(t, IsValidMapFileName("abc123.map"), true)
	require.Equal(t, IsValidMapFileName("abc 123.map"), false)
	require.Equal(t, IsValidMapFileName("abc 123.dmap"), false)
}