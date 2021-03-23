package cmd

import (
	"os"
	"time"

	"github.com/mvisonneau/go-helpers/logger"
	"github.com/mvisonneau/slack-git-compare/pkg/config"
	"github.com/mvisonneau/slack-git-compare/pkg/providers"
	"github.com/urfave/cli/v2"

	log "github.com/sirupsen/logrus"
)

var start time.Time

func configure(ctx *cli.Context) config.Config {
	start = ctx.App.Metadata["startTime"].(time.Time)

	assertStringVariableDefined(ctx, "config")

	cfg, err := config.ParseFile(ctx.String("config"))
	if err != nil {
		log.WithError(err).Fatal("loading config file")
	}

	configCliOverrides(ctx, &cfg)

	if err = cfg.Validate(); err != nil {
		log.WithError(err).Fatal("invalid config")
	}

	// Configure logger
	if err := logger.Configure(logger.Config{
		Level:  cfg.Log.Level,
		Format: cfg.Log.Format,
	}); err != nil {
		log.WithError(err).Fatal("invalid logging config")
	}

	return cfg
}

func configCliOverrides(ctx *cli.Context, cfg *config.Config) {
	// Override Slack config if necessary
	if ctx.String("slack-token") != "" {
		cfg.Slack.Token = ctx.String("slack-token")
	}

	if ctx.String("slack-signing-secret") != "" {
		cfg.Slack.SigningSecret = ctx.String("slack-signing-secret")
	}

	// Override providers config if necessary
	for f, t := range map[string]providers.ProviderType{
		"github-token": providers.ProviderTypeGitHub,
		"gitlab-token": providers.ProviderTypeGitLab,
	} {
		if ctx.String(f) != "" {
			for k, p := range cfg.Providers {
				if p.Type == t.String() {
					cfg.Providers[k].Token = ctx.String(f)
					break
				}
			}
		}
	}
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
