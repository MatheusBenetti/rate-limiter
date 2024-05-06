package entity

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerateValue(t *testing.T) {
	at := &ApiKey{}
	err := at.GenerateValue()

	require.NoError(t, err, "GenerateValue should not return an error")
	require.NotEmpty(t, at.Value(), "Generated value should not be empty")
	require.Len(t, at.Value(), 64, "Generated value should be 32 characters long")
}
