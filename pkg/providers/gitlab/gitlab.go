package gitlab

import (
	"fmt"
	"strings"

	"github.com/mvisonneau/slack-git-compare/pkg/providers"

	log "github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
)

// Provider implements the Provider interface for GitLab
type Provider struct {
	client     *gitlab.Client
	groups     []string
	webBaseURL string
}

// NewProvider returns a new Provider with a new GitLab client instanciation and
// associated config
func NewProvider(token, baseURL string, groups []string) (p Provider, err error) {
	p.client, err = gitlab.NewClient(
		token,
		gitlab.WithBaseURL(baseURL),
		gitlab.WithoutRetries(),
	)

	p.groups = groups

	if baseURL != "" {
		p.webBaseURL = baseURL
	} else {
		p.webBaseURL = "https://gitlab.com"
	}

	return
}

// Type returns the provider type
func (p Provider) Type() providers.ProviderType {
	return providers.ProviderTypeGitLab
}

// WebBaseURL returns the base URL for HTML rendered pages (non-API)
func (p Provider) WebBaseURL() string {
	return p.webBaseURL
}

// ListRepositories returns the list of all non archived projects which belong to
// the groups configured as well as their subgroups
func (p Provider) ListRepositories() (repos providers.Repositories, err error) {
	repos = make(providers.Repositories)
	var fetchedRepos []*gitlab.Project
	var resp *gitlab.Response

	for _, group := range p.groups {
		log.WithFields(log.Fields{
			"provider": providers.ProviderTypeGitLab,
			"group":    group,
		}).Debug("fetching projects")

		opts := &gitlab.ListGroupProjectsOptions{
			ListOptions: gitlab.ListOptions{
				Page:    1,
				PerPage: 100,
			},
			Archived:         gitlab.Bool(false),
			IncludeSubgroups: gitlab.Bool(true),
			WithShared:       gitlab.Bool(false),
		}

		for {
			fetchedRepos, resp, err = p.client.Groups.ListGroupProjects(group, opts)
			if err != nil {
				return
			}

			for _, repo := range fetchedRepos {
				r := &providers.Repository{
					ProviderType: providers.ProviderTypeGitLab,
					Name:         repo.PathWithNamespace,
					WebURL:       repo.WebURL,
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
func (p Provider) Compare(project string, fromRef, toRef providers.Ref) (cmp *providers.Comparison, err error) {
	cmp = &providers.Comparison{}
	opts := &gitlab.CompareOptions{}

	if fromRef.OriginRef != nil {
		opts.From = gitlab.String(fromRef.OriginRef.Name)
	} else {
		opts.From = gitlab.String(fromRef.Name)
	}

	if toRef.OriginRef != nil {
		opts.To = gitlab.String(toRef.OriginRef.Name)
	} else {
		opts.To = gitlab.String(toRef.Name)
	}

	var gitlabCompare *gitlab.Compare
	if gitlabCompare, _, err = p.client.Repositories.Compare(project, opts); err != nil {
		return
	}

	cmp.WebURL = fmt.Sprintf("%s/%s/-/compare/%s...%s", p.WebBaseURL(), project, *opts.From, *opts.To)
	for _, commit := range gitlabCompare.Commits {
		cmp.Commits = append(cmp.Commits, providers.Commit{
			ID:      commit.ID,
			ShortID: commit.ShortID,
			Author: providers.Author{
				Name:  commit.AuthorName,
				Email: commit.AuthorEmail,
			},
			CreatedAt: *commit.CreatedAt,
			Message:   commit.Message,
			WebURL:    commit.WebURL,
		})
	}

	return
}

// ListRefs returns all the Refs for a given project
func (p Provider) ListRefs(project string) (refs providers.Refs, err error) {
	refs = make(providers.Refs)
	branches, err := p.ListProjectBranches(project)
	if err != nil {
		return
	}

	for k, r := range branches {
		refs[k] = r
	}

	tags, err := p.ListProjectTags(project)
	if err != nil {
		return
	}

	for k, r := range tags {
		refs[k] = r
	}

	envs, err := p.ListProjectEnvironments(project)
	if err != nil {
		return
	}

	for k, r := range envs {
		refs[k] = r
	}

	return
}

// ListProjectBranches returns all the branches for a given project
func (p Provider) ListProjectBranches(projectName string) (refs providers.Refs, err error) {
	refs = make(providers.Refs)
	opts := &gitlab.ListBranchesOptions{
		ListOptions: gitlab.ListOptions{
			Page:    1,
			PerPage: 100,
		},
	}

	for {
		var foundBranches []*gitlab.Branch
		var resp *gitlab.Response
		foundBranches, resp, err = p.client.Branches.ListBranches(projectName, opts)
		if err != nil {
			return
		}

		for _, branch := range foundBranches {
			ref := &providers.Ref{
				Name:   branch.Name,
				Type:   providers.RefTypeBranch,
				WebURL: branch.WebURL,
			}
			refs[ref.Key()] = ref
		}

		if resp.CurrentPage >= resp.TotalPages {
			break
		}

		opts.Page = resp.NextPage
	}

	return
}

// ListProjectTags returns all the tags for a given project
func (p *Provider) ListProjectTags(projectName string) (refs providers.Refs, err error) {
	refs = make(providers.Refs)
	opts := &gitlab.ListTagsOptions{
		ListOptions: gitlab.ListOptions{
			Page:    1,
			PerPage: 100,
		},
	}

	for {
		var foundTags []*gitlab.Tag
		var resp *gitlab.Response
		foundTags, resp, err = p.client.Tags.ListTags(projectName, opts)
		if err != nil {
			return
		}

		for _, tag := range foundTags {
			ref := &providers.Ref{
				Name: tag.Name,
				Type: providers.RefTypeTag,
				// TODO: Provide the correct URL
				WebURL: tag.Commit.WebURL,
			}
			refs[ref.Key()] = ref
		}

		if resp.CurrentPage >= resp.TotalPages {
			break
		}
		opts.Page = resp.NextPage
	}

	return
}

// ListProjectEnvironments returns all the "available" environments for a given project.
// It omits environments which start with "review/"
func (p *Provider) ListProjectEnvironments(project string) (refs providers.Refs, err error) {
	refs = make(providers.Refs)
	opts := &gitlab.ListEnvironmentsOptions{
		Page:    1,
		PerPage: 100,
	}

	for {
		var foundEnvs []*gitlab.Environment
		var resp *gitlab.Response
		foundEnvs, resp, err = p.client.Environments.ListEnvironments(project, opts)
		if err != nil {
			return
		}

		for _, env := range foundEnvs {
			if env.State == "available" && !strings.HasPrefix("review/", env.Name) {
				var envDetails *gitlab.Environment
				envDetails, resp, err = p.client.Environments.GetEnvironment(project, env.ID)
				if err != nil {
					return
				}

				if envDetails.LastDeployment != nil && envDetails.LastDeployment.Deployable.Commit != nil {
					ref := &providers.Ref{
						Name: env.Name,
						Type: providers.RefTypeEnvironment,
						// TODO: Provide the correct URL
						// WebURL: env.LastDeployment.Deployable.Commit.WebURL,
						OriginRef: &providers.Ref{
							Name: envDetails.LastDeployment.Deployable.Commit.ID,
							Type: providers.RefTypeCommit,
						},
					}
					refs[ref.Key()] = ref
				}
			}
		}

		if resp.CurrentPage >= resp.TotalPages {
			break
		}
		opts.Page = resp.NextPage
	}

	return
}
