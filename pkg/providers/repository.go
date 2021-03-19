package providers

import (
	"hash/crc32"
	"sort"
	"strconv"
	"time"

	"github.com/lithammer/fuzzysearch/fuzzy"
)

// Repository holds details of a git repository
type Repository struct {
	ProviderType   ProviderType
	Refs           Refs
	Name           string
	LastRefsUpdate time.Time
	WebURL         string
}

// RepositoryKey is a unique identifier for a Repository
type RepositoryKey string

// Key returns a unique identifier based upon the Name and ProviderType
// of the Repository
func (r Repository) Key() RepositoryKey {
	return RepositoryKey(strconv.Itoa(int(crc32.ChecksumIEEE([]byte(r.ProviderType.String() + r.Name)))))
}

// Repositories holds multiple Repository index with their unique identifiers (RepositoryKey)
type Repositories map[RepositoryKey]*Repository

// RankedRepository can be used when fuzzy searching Repositories, attributing a "rank"
// for the Repository given the pertinence of its attributes given the search
type RankedRepository struct {
	*Repository
	Rank int
}

// RankedRepositories is a slice of *RankedRepository
type RankedRepositories []*RankedRepository

// GetByKey returns a Repository given its RepositoryKey
func (rs Repositories) GetByKey(k RepositoryKey) (r *Repository, ok bool) {
	r, ok = rs[k]
	return
}

// Search looks up for repositories by Name in a fuzzy finding fashion, it will return
// then sorted by pertinence
func (rs Repositories) Search(filter string, limit int) (repos RankedRepositories) {
	for _, r := range rs {
		if len(filter) == 0 {
			repos = append(repos, &RankedRepository{
				Repository: r,
				Rank:       0,
			})
		} else if rank := fuzzy.RankMatchNormalizedFold(filter, r.Name); rank >= 0 {
			repos = append(repos, &RankedRepository{
				Repository: r,
				Rank:       rank,
			})
		}
	}

	sort.SliceStable(repos, func(i, j int) bool {
		if repos[i].Rank == repos[j].Rank {
			return repos[i].Name > repos[j].Name
		}
		return repos[i].Rank < repos[j].Rank
	})

	if len(repos) > limit {
		return repos[:limit]
	}

	return
}

// GetByClosestNameMatch returns the Repository which is the most pertinent
// given the Name
func (rs Repositories) GetByClosestNameMatch(name string) (repo *Repository) {
	if len(name) == 0 {
		return
	}

	for _, r := range rs.Search(name, 1) {
		repo = r.Repository
		break
	}
	return
}
