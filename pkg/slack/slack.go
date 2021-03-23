package slack

import (
	"context"
	"fmt"
	"strings"

	"github.com/mvisonneau/slack-git-compare/pkg/config"
	"github.com/mvisonneau/slack-git-compare/pkg/providers"
	"github.com/mvisonneau/slack-git-compare/pkg/providers/github"
	"github.com/mvisonneau/slack-git-compare/pkg/providers/gitlab"
	"github.com/mvisonneau/slack-git-compare/pkg/storage"

	log "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
)

// Slack holds clients and configuration to perform operations
// upon the Slack API, git providers and storage/cache endpoints
type Slack struct {
	Client        *slack.Client
	SigningSecret string
	Providers     providers.Providers
	Storage       storage.Storage
	CustomUsers   config.Users
}

// NewOptions holds configuration parameters to create a new and
// operational Slack object
type NewOptions struct {
	ProviderGitHubToken  string
	ProviderGitHubURL    string
	ProviderGitHubOrgs   []string
	ProviderGitLabToken  string
	ProviderGitLabURL    string
	ProviderGitLabGroups []string
	SigningSecret        string
	Token                string
}

// New creates and configures a new Slack object
func New(ctx context.Context, slackConfig config.Slack, providersConfig config.Providers, customUsers config.Users) (s Slack, err error) {
	s.Client = slack.New(slackConfig.Token)
	s.SigningSecret = slackConfig.SigningSecret
	s.CustomUsers = customUsers
	if err = s.configureProviders(ctx, providersConfig); err != nil {
		return
	}

	s.Storage.Repositories, err = s.Providers.ListRepositories()
	if err != nil {
		return
	}

	s.Storage.SlackUserEmailMappings, err = s.ListSlackUserEmailMappings()

	return
}

func (s *Slack) configureProviders(ctx context.Context, providersConfig config.Providers) error {
	s.Providers = make(providers.Providers)

	if len(providersConfig) == 0 {
		return fmt.Errorf("you must configure at least one git provider, none given")
	}

	for _, p := range providersConfig {
		if len(p.Owners) == 0 {
			return fmt.Errorf("you must define at least one 'owners', none given")
		}

		pt, err := providers.GetProviderTypeFromString(p.Type)
		if err != nil {
			return err
		}

		switch pt {
		case providers.ProviderTypeGitHub:
			s.Providers[pt], err = github.NewProvider(ctx, p.Token, p.URL, p.Owners)
		case providers.ProviderTypeGitLab:
			s.Providers[pt], err = gitlab.NewProvider(p.Token, p.URL, p.Owners)
		}

		if err != nil {
			return err
		}

		log.WithFields(log.Fields{
			"provider": pt.String(),
			"orgs":     p.Owners,
		}).Debug("configured provider")
	}

	return nil
}

func (s Slack) sendMessage(channelID, msg string) (err error) {
	_, _, err = s.Client.PostMessage(channelID,
		slack.MsgOptionText(msg, false),
		slack.MsgOptionAttachments())
	return
}

func stripRankFromValue(value string) string {
	values := strings.Split(value, "/")
	if len(values) != 2 {
		return ""
	}
	return values[1]
}
