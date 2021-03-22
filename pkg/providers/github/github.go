package github

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/mvisonneau/slack-git-compare/pkg/providers"

	"github.com/google/go-github/v33/github"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

// Provider implements the Provider interface for GitHub
type Provider struct {
	ctx        context.Context
	client     *github.Client
	orgs       []string
	webBaseURL string
}

// NewProvider returns a new Provider with a new GitHub client instanciation and
// associated config
func NewProvider(ctx context.Context, token, baseURL string, orgs []string) (p Provider, err error) {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	p.ctx = ctx
	p.client = github.NewClient(tc)
	p.client.BaseURL, err = url.Parse(baseURL)
	p.orgs = orgs

	// TODO: This is probably not going to work for everyone, I suppose we should
	// consider adding a new/dedicated flag
	p.webBaseURL = strings.Replace(baseURL, "api.", "", -1)
	return
}

// Type returns the provider type
func (p Provider) Type() providers.ProviderType {
	return providers.ProviderTypeGitHub
}

// WebBaseURL returns the base URL for HTML rendered pages (non-API)
func (p Provider) WebBaseURL() string {
	return p.webBaseURL
}

// ListRepositories returns the list of all projects which belong to
// the organizations configured
func (p Provider) ListRepositories() (repos providers.Repositories, err error) {
	repos = make(providers.Repositories)
	var fetchedRepos []*github.Repository
	var resp *github.Response

	for _, org := range p.orgs {
		log.WithFields(log.Fields{
			"provider": providers.ProviderTypeGitHub,
			"org":      org,
		}).Debug("fetching projects")

		opts := &github.RepositoryListByOrgOptions{
			ListOptions: github.ListOptions{
				Page:    1,
				PerPage: 100,
			},
		}

		for {
			fetchedRepos, resp, err = p.client.Repositories.ListByOrg(p.ctx, org, opts)
			if err != nil {
				return
			}

			for _, repo := range fetchedRepos {
				r := &providers.Repository{
					ProviderType: providers.ProviderTypeGitHub,
					Name:         *repo.FullName,
					WebURL:       *repo.URL,
				}
				repos[r.Key()] = r
			}

			if resp.NextPage == 0 {
				break
			}

			opts.Page++
		}
	}

	return
}

// Compare calculates the diff between two git references
func (p Provider) Compare(project string, fromRef, toRef providers.Ref) (cmp providers.Comparison, err error) {
	projectValues := strings.Split(project, "/")
	if len(projectValues) != 2 {
		err = fmt.Errorf("invalid project name '%s'", project)
		return
	}

	var githubCompare *github.CommitsComparison
	if githubCompare, _, err = p.client.Repositories.CompareCommits(p.ctx, projectValues[0], projectValues[1], fromRef.Name, toRef.Name); err != nil {
		return
	}

	cmp.WebURL = fmt.Sprintf("%s/%s/compare/%s...%s", p.WebBaseURL(), project, fromRef.Name, toRef.Name)
	for _, commit := range githubCompare.Commits {
		cmp.Commits = append(cmp.Commits, providers.Commit{
			ID:          commit.GetSHA(),
			ShortID:     commit.GetSHA()[:9],
			AuthorName:  *commit.Commit.GetAuthor().Name,
			AuthorEmail: *commit.Commit.GetAuthor().Email,
			CreatedAt:   commit.Committer.GetCreatedAt().Time,
			Message:     commit.Commit.GetMessage(),
			WebURL:      commit.GetURL(),
		})
	}

	return
}

// ListRefs returns all the Refs for a given project
func (p Provider) ListRefs(project string) (refs providers.Refs, err error) {
	projectValues := strings.Split(project, "/")
	if len(projectValues) != 2 {
		err = fmt.Errorf("invalid project name '%s'", project)
		return
	}

	refs = make(providers.Refs)
	branches, err := p.ListRepositoryBranches(projectValues[0], projectValues[1])
	if err != nil {
		return
	}

	for k, r := range branches {
		refs[k] = r
	}

	tags, err := p.ListRepositoryTags(projectValues[0], projectValues[1])
	if err != nil {
		return
	}

	for k, r := range tags {
		refs[k] = r
	}

	return
}

// ListRepositoryBranches returns all the branches for a given repository
func (p Provider) ListRepositoryBranches(owner, repo string) (refs providers.Refs, err error) {
	refs = make(providers.Refs)
	opts := &github.BranchListOptions{
		ListOptions: github.ListOptions{
			Page:    1,
			PerPage: 100,
		},
	}

	for {
		var foundBranches []*github.Branch
		var resp *github.Response
		foundBranches, resp, err = p.client.Repositories.ListBranches(p.ctx, owner, repo, opts)
		if err != nil {
			return
		}

		for _, branch := range foundBranches {
			ref := &providers.Ref{
				Name: *branch.Name,
				Type: providers.RefTypeBranch,
				// TODO: compute something more pertinent
				WebURL: p.webBaseURL,
			}
			refs[ref.Key()] = ref
		}

		if resp.NextPage == 0 {
			break
		}

		opts.Page++
	}

	return
}

// ListRepositoryTags returns all the tags for a given repository
func (p *Provider) ListRepositoryTags(owner, repo string) (refs providers.Refs, err error) {
	refs = make(providers.Refs)
	opts := &github.ListOptions{
		Page:    1,
		PerPage: 100,
	}

	for {
		var foundTags []*github.RepositoryTag
		var resp *github.Response
		foundTags, resp, err = p.client.Repositories.ListTags(p.ctx, owner, repo, opts)
		if err != nil {
			return
		}

		for _, tag := range foundTags {
			ref := &providers.Ref{
				Name: *tag.Name,
				Type: providers.RefTypeTag,
				// TODO: Provide the correct URL
				WebURL: p.webBaseURL,
			}
			refs[ref.Key()] = ref
		}

		if resp.NextPage == 0 {
			break
		}

		opts.Page++
	}

	return
}
