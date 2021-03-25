package providers

import (
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// Comparison holds the information of a git compare response
type Comparison struct {
	Commits Commits
	WebURL  string
	// FromRef string
	// ToRef   string
}

// Author holds details about the author of a commit
type Author struct {
	Name        string
	Email       string
	SlackUserID string
}

// Authors is a slice of Author
type Authors []Author

// Commit holds the details of a git commit
type Commit struct {
	ID        string
	ShortID   string
	Author    Author
	CreatedAt time.Time
	Message   string
	WebURL    string
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
	if len(c.Message) > 75 {
		c.Message = c.Message[:73] + ".."
	}

	if i := strings.Index(c.Message, "\n"); i != -1 {
		return c.Message[:i]
	}

	return c.Message
}

// AuthorSlackString returns a string which can be nicely
// rendered through Slack
func (c Commit) AuthorSlackString() string {
	if c.Author.SlackUserID != "" {
		return fmt.Sprintf("*<@%s>*", c.Author.SlackUserID)
	}

	return fmt.Sprintf("*%s* - _%s_", c.Author.Name, c.Author.Email)
}

// HydrateCommitsAuthorsWithSlackUserID adds the SlackUserID of an author based
// on its email address
func (c *Comparison) HydrateCommitsAuthorsWithSlackUserID(mapping map[string]string) {
	for k, commit := range c.Commits {
		if slackUserID, found := mapping[commit.Author.Email]; found {
			c.Commits[k].Author.SlackUserID = slackUserID
			log.WithFields(log.Fields{
				"commit_id":     commit.ID,
				"email":         commit.Author.Email,
				"slack_user_id": commit.Author.SlackUserID,
			}).Trace("hydrated commit author")
		} else {
			log.WithFields(log.Fields{
				"commit_id": commit.ID,
				"email":     commit.Author.Email,
			}).Trace("could not hydrate commit author")
		}
	}
}

// GetAuthors returns Authors who appeared to have make
// contribution(s) in the comparison
func (c Comparison) GetAuthors() (authors Authors) {
	slackUserIDMapping := make(map[string]Author)
	emailMapping := make(map[string]Author)
	for _, commit := range c.Commits {
		if commit.Author.SlackUserID != "" {
			if _, found := slackUserIDMapping[commit.Author.SlackUserID]; !found {
				slackUserIDMapping[commit.Author.SlackUserID] = commit.Author
			}
			continue
		}

		if _, found := emailMapping[commit.Author.Email]; !found {
			emailMapping[commit.Author.Email] = commit.Author
		}
	}

	for _, author := range slackUserIDMapping {
		authors = append(authors, author)
	}

	for _, author := range emailMapping {
		authors = append(authors, author)
	}

	return
}

// AuthorsSlackString returns a string containing authors who contributed
// within the diff, in a format which can be nicely rendered in Slack
func (c Comparison) AuthorsSlackString() (out string) {
	if c.CommitCount() > 0 {
		if c.CommitCount() > 1 {
			out = "commits from "
		} else {
			out = "commit from "
		}

		cursor := 0
		for _, author := range c.GetAuthors() {
			if cursor > 0 {
				if cursor == int(c.CommitCount())-1 || cursor == 7 {
					out += " and "
				} else {
					out += ", "
				}
			}

			if cursor == 7 {
				remainingCount := int(c.CommitCount()) - cursor
				out += fmt.Sprintf("%d other", remainingCount)
				if remainingCount > 1 {
					out += "s"
				}
				break
			}

			cursor++
			if author.SlackUserID != "" {
				out += fmt.Sprintf("<@%s>", author.SlackUserID)
				continue
			}

			out += fmt.Sprintf("_%s_", author.Email)
		}
	} else {
		out = "no authors"
	}

	return
}
