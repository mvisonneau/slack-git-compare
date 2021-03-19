package providers

import (
	"strings"
	"time"
)

// Comparison holds the information of a git compare response
type Comparison struct {
	Commits Commits
	WebURL  string
	// FromRef string
	// ToRef   string
}

// Commit holds the details of a git commit
type Commit struct {
	ID          string
	ShortID     string
	AuthorName  string
	AuthorEmail string
	CreatedAt   time.Time
	Message     string
	WebURL      string
}

// Commits is a slice of Commit
type Commits []Commit

// CommitCount returns the amount of commits
func (c Comparison) CommitCount() uint {
	return uint(len(c.Commits))
}

// ShortMessage truncates commit messages down to 80 chars
// and omits return carriages
func (c Commit) ShortMessage() string {
	if len(c.Message) > 80 {
		c.Message = c.Message[:80] + "..."
	}

	if i := strings.Index(c.Message, "\n"); i != -1 {
		return c.Message[:i]
	}

	return c.Message
}
