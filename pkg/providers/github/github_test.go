package github

import (
	"testing"

	"github.com/mvisonneau/slack-git-compare/pkg/providers"

	"github.com/stretchr/testify/assert"
)

func TestType(t *testing.T) {
	p := Provider{}
	assert.Equal(t, providers.ProviderTypeGitHub, p.Type())
}
