package providers

import (
	"hash/crc32"
	"sort"
	"strconv"

	"github.com/lithammer/fuzzysearch/fuzzy"
)

// Ref holds details of a git reference
type Ref struct {
	Name   string
	Type   RefType
	WebURL string

	// OriginRef can be used to store the ref onto which a GitLab environment
	// is pointing to
	OriginRef *Ref
}

// Refs holds multiple Ref with their unique identifiers (RefKey)
type Refs map[RefKey]*Ref

// RankedRef can be used when fuzzy searching Refs, attributing a "rank"
// for the Ref given the pertinence of its attributes given the search
type RankedRef struct {
	*Ref
	Rank int
}

// RankedRefs is a slice of *RankedRef
type RankedRefs []*RankedRef

// RefType represents the type of git reference
type RefType uint8

const (
	// RefTypeBranch represent a git branch
	RefTypeBranch RefType = iota

	// RefTypeCommit represent a git commit
	RefTypeCommit

	// RefTypeEnvironment represent a git environment (only used for GitLab)
	RefTypeEnvironment

	// RefTypeTag represent a git tag
	RefTypeTag
)

// String returns the type as a readable string
func (rt RefType) String() string {
	return [...]string{
		"branch",
		"commit",
		"env",
		"tag",
	}[rt]
}

// RefKey is a unique identifier for a Ref
type RefKey string

// Key returns a unique identifier based upon the Type and Name  of the Ref
func (r Ref) Key() RefKey {
	return RefKey(strconv.Itoa(int(crc32.ChecksumIEEE([]byte(r.Type.String() + r.Name)))))
}

// GetByKey returns a Ref given its RefKey
func (rs Refs) GetByKey(k RefKey) (r *Ref, ok bool) {
	r, ok = rs[k]
	return
}

// Search looks up for references by Name in a fuzzy finding fashion, it will return
// them sorted by pertinence
func (rs Refs) Search(filter string, limit int) (refs RankedRefs) {
	for _, r := range rs {
		if rank := fuzzy.RankMatchNormalizedFold(filter, r.Name); rank >= 0 {
			refs = append(refs, &RankedRef{
				Ref:  r,
				Rank: rank,
			})
		}
	}

	sort.SliceStable(refs, func(i, j int) bool {
		if refs[i].Rank == refs[j].Rank {
			return refs[i].Name > refs[j].Name
		}
		return refs[i].Rank < refs[j].Rank
	})

	if len(refs) > limit {
		return refs[:limit]
	}

	return
}

// GetByClosestNameMatch returns the Ref which is the most pertinent
// given its Name
func (rs Refs) GetByClosestNameMatch(name string) (ref *Ref) {
	if len(name) == 0 {
		return
	}

	for _, r := range rs.Search(name, 1) {
		ref = r.Ref
		break
	}
	return
}
