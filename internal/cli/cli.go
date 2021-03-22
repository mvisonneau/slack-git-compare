package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/mvisonneau/slack-git-compare/internal/cmd"
	"github.com/urfave/cli/v2"
)

// Run handles the instanciation of the CLI application
func Run(version string, args []string) {
	err := NewApp(version, time.Now()).Run(args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// NewApp configures the CLI application
func NewApp(version string, start time.Time) (app *cli.App) {
	app = cli.NewApp()
	app.Name = "slack-git-compare"
	app.Version = version
	app.Usage = "Compare git refs within Slack"
	app.EnableBashCompletion = true

	app.Flags = cli.FlagsByName{
		&cli.StringSliceFlag{
			Name:    "github-orgs",
			EnvVars: []string{"SGC_GITHUB_ORGS"},
			Usage:   "GitHub `organizations` to list repositories from (can be set multiple times)",
		},
		&cli.StringFlag{
			Name:    "github-url",
			EnvVars: []string{"SGC_GITHUB_URL"},
			Usage:   "GitHub `url`",
			Value:   "https://api.github.com/",
		},
		&cli.StringFlag{
			Name:    "github-token",
			EnvVars: []string{"SGC_GITHUB_TOKEN"},
			Usage:   "GitHub `token`",
		},
		&cli.StringSliceFlag{
			Name:    "gitlab-groups",
			EnvVars: []string{"SGC_GITLAB_GROUPS"},
			Usage:   "GitLab `groups` to list repositories from (can be set multiple times)",
		},
		&cli.StringFlag{
			Name:    "gitlab-url",
			EnvVars: []string{"SGC_GITLAB_URL"},
			Usage:   "GitLab `url`",
			Value:   "https://gitlab.com",
		},
		&cli.StringFlag{
			Name:    "gitlab-token",
			EnvVars: []string{"SGC_GITLAB_TOKEN"},
			Usage:   "GitLab `token`",
		},
		&cli.StringFlag{
			Name:    "slack-token",
			EnvVars: []string{"SGC_SLACK_TOKEN"},
			Usage:   "Slack `token`",
		},
		&cli.StringFlag{
			Name:    "slack-signing-secret",
			EnvVars: []string{"SGC_SLACK_SIGNING_SECRET"},
			Usage:   "Slack `signing-secret`",
		},
		&cli.StringFlag{
			Name:    "listen-address",
			EnvVars: []string{"SGC_LISTEN_ADDRESS"},
			Usage:   "`address` to bind our http server upon",
			Value:   ":8080",
		},
		&cli.StringFlag{
			Name:    "log-level",
			EnvVars: []string{"SGC_LOG_LEVEL"},
			Usage:   "log `level` (debug,info,warn,fatal,panic)",
			Value:   "info",
		},
		&cli.StringFlag{
			Name:    "log-format",
			EnvVars: []string{"SGC_LOG_FORMAT"},
			Usage:   "log `format` (json,text)",
			Value:   "text",
		},
	}

	app.Action = cmd.ExecWrapper(cmd.Run)

	app.Metadata = map[string]interface{}{
		"startTime": start,
	}

	return
}
