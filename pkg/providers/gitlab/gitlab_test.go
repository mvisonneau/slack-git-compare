package gitlab

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mvisonneau/slack-git-compare/pkg/providers"
	"github.com/stretchr/testify/assert"
	"github.com/xanzy/go-gitlab"
)

// Mocking helpers
func getMockedProvider() (*http.ServeMux, *httptest.Server, Provider) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)

	gc, _ := gitlab.NewClient(
		"foo",
		gitlab.WithBaseURL(server.URL),
		gitlab.WithoutRetries(),
	)

	return mux, server, Provider{
		client:     gc,
		groups:     []string{"foo"},
		webBaseURL: server.URL,
	}
}

func TestNewProvider(t *testing.T) {
	groups := []string{"foo", "bar"}
	p, err := NewProvider("foo", "http://foo", groups)
	assert.NoError(t, err)
	assert.Equal(t, groups, p.groups)
}

func TestType(t *testing.T) {
	p := Provider{}
	assert.Equal(t, providers.ProviderTypeGitLab, p.Type())
}

func TestWebBaseURL(t *testing.T) {
	p := Provider{webBaseURL: "http://foo"}
	assert.Equal(t, p.webBaseURL, p.WebBaseURL())
}

func TestListRepositories(t *testing.T) {
	mux, server, p := getMockedProvider()
	defer server.Close()

	mux.HandleFunc("/api/v4/groups/foo/projects",
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, r.Method, "GET")
			fmt.Fprint(w, `
			[
				{
					"id": 1,
					"path_with_namespace": "foo/bar"
				},
				{
					"id": 2,
					"path_with_namespace": "foo/baz"
				}
			]`)
		})

	repos, err := p.ListRepositories()
	assert.NoError(t, err)
	assert.Len(t, repos, 2)
}
