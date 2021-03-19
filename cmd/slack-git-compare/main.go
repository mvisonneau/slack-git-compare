package main

import (
	"os"

	"github.com/mvisonneau/slack-git-compare/internal/cli"
)

var version = ""

func main() {
	cli.Run(version, os.Args)
}
