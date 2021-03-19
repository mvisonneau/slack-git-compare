package providers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRepositoryKey(t *testing.T) {
	assert.Equal(t, RepositoryKey("1040242264"), Repository{
		Name:         "foo",
		ProviderType: ProviderTypeGitHub,
	}.Key())
}

func TestRepositoriesGetByKey(t *testing.T) {
	r := Repository{
		Name:         "foo",
		ProviderType: ProviderTypeGitHub,
	}
	rs := make(Repositories)
	rs[r.Key()] = &r

	foundRepository, ok := rs.GetByKey(r.Key())
	assert.True(t, ok)
	assert.Equal(t, &r, foundRepository)

	foundRepository, ok = rs.GetByKey(Repository{}.Key())
	assert.False(t, ok)
	assert.Nil(t, foundRepository)
}
