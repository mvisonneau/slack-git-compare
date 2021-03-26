package providers

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

// Provider is used to represent any kind of git provider
type Provider interface {
	WebBaseURL() string
	Type() ProviderType
	Compare(string, Ref, Ref) (*Comparison, error)
	ListRepositories() (Repositories, error)
	ListRefs(string) (Refs, error)
}

// ProviderType represents the type of git provider
type ProviderType uint8

// Providers can store multiple Provider based on their types
type Providers map[ProviderType]Provider

const (
	// ProviderTypeGitHub for GitHub provider
	ProviderTypeGitHub ProviderType = iota

	// ProviderTypeGitLab for GitLab provider
	ProviderTypeGitLab
)

// String returns the name of the provider (lowercase)
func (pt ProviderType) String() string {
	return [...]string{"github", "gitlab"}[pt]
}

// StringPretty returns the name of the provider using their capitalization
// attributes
func (pt ProviderType) StringPretty() string {
	return [...]string{"GitHub", "GitLab"}[pt]
}

// ListRepositories aggregates the repositories for all configured providers
func (ps Providers) ListRepositories() (repos Repositories, err error) {
	repos = make(Repositories)
	for _, p := range ps {
		foundRepos, err := p.ListRepositories()
		if err != nil {
			return repos, err
		}

		for k, r := range foundRepos {
			repos[k] = r
		}

		log.WithFields(log.Fields{
			"provider": p.Type().String(),
			"count":    len(foundRepos),
		}).Info("fetched repositories from provider")
	}

	log.WithFields(log.Fields{
		"total": len(repos),
	}).Debug("done fetching repositories")

	return repos, nil
}

// GetProviderTypeFromString returns a ProviderType based onto a given string
func GetProviderTypeFromString(p string) (pt ProviderType, err error) {
	mapping := map[string]ProviderType{
		"github": ProviderTypeGitHub,
		"gitlab": ProviderTypeGitLab,
	}

	var found bool
	pt, found = mapping[p]
	if !found {
		err = fmt.Errorf("invalid provider type '%s'", p)
	}

	return
}
