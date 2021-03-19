package providers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProviderStrings(t *testing.T) {
	assert.Equal(t, "github", ProviderTypeGitHub.String())
	assert.Equal(t, "gitlab", ProviderTypeGitLab.String())
}
