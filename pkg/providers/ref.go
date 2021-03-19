package providers

import (
	"hash/crc32"
	"sort"
	"strconv"

	"github.com/lithammer/fuzzysearch/fuzzy"
)

type Ref struct {
	Name   string
	Type   RefType
	WebURL string

	// OriginRef can be used to store the ref onto which a GitLab environment
	// is pointing to
	OriginRef *Ref
}

type Refs map[RefKey]*Ref

type RankedRef struct {
	*Ref
	Rank int
}

type RankedRefs []*RankedRef

type RefType uint8

const (
	RefTypeBranch RefType = iota
	RefTypeCommit
	RefTypeEnvironment
	RefTypeTag
)

func (rt RefType) String() string {
	return [...]string{
		"branch",
		"commit",
		"env",
		"tag",
	}[rt]
}

// RefKey ..
type RefKey string

// Key ..
func (r Ref) Key() RefKey {
	return RefKey(strconv.Itoa(int(crc32.ChecksumIEEE([]byte(r.Type.String() + r.Name)))))
}

func (rs Refs) GetByKey(k RefKey) (r *Ref, ok bool) {
	r, ok = rs[k]
	return
}

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
