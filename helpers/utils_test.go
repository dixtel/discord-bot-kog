package helpers

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetRandomString(t *testing.T) {
	require.Len(t, GetRandomString(), 64)
	require.False(t, GetRandomString() == GetRandomString())
}