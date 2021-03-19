package providers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComparisonCommitCount(t *testing.T) {
	c := Comparison{
		Commits: Commits{
			Commit{
				ID: "foo",
			},
			Commit{
				ID: "bar",
			},
		},
	}
	assert.Equal(t, uint(2), c.CommitCount())
}
