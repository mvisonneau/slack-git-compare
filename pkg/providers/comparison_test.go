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

func TestAuthorsSlackString(t *testing.T) {
	// No commits
	c := Comparison{}
	assert.Equal(t, "no authors", c.AuthorsSlackString())

	// Single commit with SlackUserID
	c.Commits = append(c.Commits, Commit{
		ID: "commit1",
		Author: Author{
			Email:       "alice@foo.bar",
			SlackUserID: "U123456789",
		},
	})
	assert.Equal(t, "commit from <@U123456789>", c.AuthorsSlackString())

	// 2 commits with SlackUserID
	c.Commits = append(c.Commits, Commit{
		ID: "commit2",
		Author: Author{
			Email:       "bob@foo.bar",
			SlackUserID: "U234567891",
		},
	})
	assert.Equal(t, "commits from <@U123456789> and <@U234567891>", c.AuthorsSlackString())

	// 3 commits
	c.Commits = append(c.Commits, Commit{
		ID: "commit3",
		Author: Author{
			Email: "alice@foo.baz",
		},
	})
	assert.Equal(t, "commits from <@U123456789>, <@U234567891> and _alice@foo.baz_", c.AuthorsSlackString())
}
