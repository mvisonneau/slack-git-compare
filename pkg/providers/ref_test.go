package providers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRefTypeString(t *testing.T) {
	assert.Equal(t, "branch", RefTypeBranch.String())
	assert.Equal(t, "commit", RefTypeCommit.String())
	assert.Equal(t, "env", RefTypeEnvironment.String())
	assert.Equal(t, "tag", RefTypeTag.String())
}

func TestRefKey(t *testing.T) {
	assert.Equal(t, RefKey("920939608"), Ref{
		Name: "foo",
		Type: RefTypeBranch,
	}.Key())
}

func TestRefsGetByKey(t *testing.T) {
	r := Ref{
		Name: "foo",
		Type: RefTypeBranch,
	}
	rs := make(Refs)
	rs[r.Key()] = &r

	foundRef, ok := rs.GetByKey(r.Key())
	assert.True(t, ok)
	assert.Equal(t, &r, foundRef)

	foundRef, ok = rs.GetByKey(Ref{}.Key())
	assert.False(t, ok)
	assert.Nil(t, foundRef)
}
