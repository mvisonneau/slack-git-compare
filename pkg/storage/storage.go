package storage

import (
	"github.com/mvisonneau/slack-git-compare/pkg/providers"
)

// Storage is handling the data we fetch from the providers APIs in
// order to not overwhelm them and also reduce the risk to get rate-limited
// TODO: Convert it to an interface in order to other storage
// providers than in-memory
type Storage struct {
	Repositories           providers.Repositories
	SlackUserEmailMappings map[string]string
}
