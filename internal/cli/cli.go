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
		&cli.StringFlag{
			Name:    "config",
			Aliases: []string{"c"},
			EnvVars: []string{"SGC_CONFIG"},
			Usage:   "config `file` (json or yaml format)",
			Value:   "./config.json",
		},
		&cli.StringFlag{
			Name:    "github-token",
			EnvVars: []string{"SGC_GITHUB_TOKEN"},
			Usage:   "GitHub `token`",
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
	}

	app.Action = cmd.ExecWrapper(cmd.Run)

	app.Metadata = map[string]interface{}{
		"startTime": start,
	}

	return
}
