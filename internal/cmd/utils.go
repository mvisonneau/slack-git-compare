package cmd

import (
	"os"
	"time"

	"github.com/mvisonneau/go-helpers/logger"
	"github.com/mvisonneau/slack-git-compare/pkg/slack"
	"github.com/urfave/cli/v2"

	log "github.com/sirupsen/logrus"
)

var start time.Time

type config struct {
	ListenAddress string
	Slack         slack.NewOptions
}

func configure(ctx *cli.Context) (c config) {
	start = ctx.App.Metadata["startTime"].(time.Time)

	if err := logger.Configure(logger.Config{
		Level:  ctx.String("log-level"),
		Format: ctx.String("log-format"),
	}); err != nil {
		_ = cli.ShowAppHelp(ctx)
		log.Errorf("incorrect logging configuration")
		os.Exit(2)
	}

	for _, i := range []string{"slack-token", "slack-signing-secret", "listen-address"} {
		assertStringVariableDefined(ctx, i)
	}

	if len(ctx.String("github-token")) > 0 {
		assertStringSliceVariableNotEmpty(ctx, "github-orgs")
	}

	if len(ctx.String("gitlab-token")) > 0 {
		assertStringSliceVariableNotEmpty(ctx, "gitlab-groups")
	}

	if len(ctx.String("github-token")) == 0 && len(ctx.String("gitlab-token")) == 0 {
		_ = cli.ShowAppHelp(ctx)
		log.Errorf("you must configure at least one git provider using --git(hub|lab)-token")
		os.Exit(2)
	}

	c.ListenAddress = ctx.String("listen-address")
	c.Slack.ProviderGitHubToken = ctx.String("github-token")
	c.Slack.ProviderGitHubURL = ctx.String("github-url")
	c.Slack.ProviderGitHubOrgs = ctx.StringSlice("github-orgs")
	c.Slack.ProviderGitLabToken = ctx.String("gitlab-token")
	c.Slack.ProviderGitLabURL = ctx.String("gitlab-url")
	c.Slack.ProviderGitLabGroups = ctx.StringSlice("gitlab-groups")
	c.Slack.SigningSecret = ctx.String("slack-signing-secret")
	c.Slack.Token = ctx.String("slack-token")
	return
}

func exit(exitCode int, err error) cli.ExitCoder {
	defer log.WithFields(
		log.Fields{
			"execution-time": time.Since(start),
		},
	).Debug("exited..")

	if err != nil {
		log.Error(err.Error())
	}

	return cli.NewExitError("", exitCode)
}

// ExecWrapper gracefully logs and exits our `run` functions
func ExecWrapper(f func(ctx *cli.Context) (int, error)) cli.ActionFunc {
	return func(ctx *cli.Context) error {
		return exit(f(ctx))
	}
}

func assertStringVariableDefined(ctx *cli.Context, k string) {
	if len(ctx.String(k)) == 0 {
		_ = cli.ShowAppHelp(ctx)
		log.Errorf("'--%s' must be set!", k)
		os.Exit(2)
	}
}

func assertStringSliceVariableNotEmpty(ctx *cli.Context, k string) {
	if len(ctx.StringSlice(k)) == 0 {
		_ = cli.ShowAppHelp(ctx)
		log.Errorf("'--%s' must be set at least once!", k)
		os.Exit(2)
	}
}
