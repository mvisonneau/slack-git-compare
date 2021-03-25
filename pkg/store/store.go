package store

import (
	"sync"
	"time"

	"github.com/mvisonneau/slack-git-compare/pkg/providers"
)

// Store is handling the data we fetch from the providers APIs in
// order to not overwhelm them and also reduce the risk to get rate-limited
// TODO: Convert it to an interface in order to other storage
// providers than in-memory
type Store struct {
	repositories           providers.Repositories
	repositoriesLastUpdate time.Time
	repositoriesMutex      sync.RWMutex

	slackUsersEmails           map[string]string
	slackUsersEmailsLastUpdate time.Time
	slackUsersEmailsMutex      sync.RWMutex
}

// UpdateRepositories ..
func (s *Store) UpdateRepositories(repos providers.Repositories) {
	s.repositoriesMutex.Lock()
	defer s.repositoriesMutex.Unlock()

	// We do not want to lose refs details
	for k, v := range s.repositories {
		if _, found := repos[k]; found {
			repos[k] = v
		}
	}

	s.repositories = repos
	s.repositoriesLastUpdate = time.Now()
}

// GetRepositories ..
func (s *Store) GetRepositories() providers.Repositories {
	s.repositoriesMutex.RLock()
	defer s.repositoriesMutex.RUnlock()
	return s.repositories
}

// GetRepositoriesLastUpdate ..
func (s *Store) GetRepositoriesLastUpdate() time.Time {
	s.repositoriesMutex.RLock()
	defer s.repositoriesMutex.RUnlock()
	return s.repositoriesLastUpdate
}

// UpdateRepository ..
func (s *Store) UpdateRepository(r providers.Repository) {
	s.repositoriesMutex.Lock()
	defer s.repositoriesMutex.Unlock()
	s.repositories[r.Key()] = r
}

// GetRepository ..
func (s *Store) GetRepository(rk providers.RepositoryKey) (r providers.Repository, found bool) {
	s.repositoriesMutex.RLock()
	defer s.repositoriesMutex.RUnlock()
	r, found = s.repositories[rk]
	return
}

// UpdateSlackUsersEmails ..
func (s *Store) UpdateSlackUsersEmails(sue map[string]string) {
	s.slackUsersEmailsMutex.Lock()
	defer s.slackUsersEmailsMutex.Unlock()
	s.slackUsersEmails = sue
	s.slackUsersEmailsLastUpdate = time.Now()
}

// GetSlackUsersEmails ..
func (s *Store) GetSlackUsersEmails() map[string]string {
	s.slackUsersEmailsMutex.RLock()
	defer s.slackUsersEmailsMutex.RUnlock()
	return s.slackUsersEmails
}

// GetSlackUsersEmailsLastUpdate ..
func (s *Store) GetSlackUsersEmailsLastUpdate() time.Time {
	s.slackUsersEmailsMutex.RLock()
	defer s.slackUsersEmailsMutex.RUnlock()
	return s.slackUsersEmailsLastUpdate
}
