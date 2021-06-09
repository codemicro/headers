package hooks

import (
	"github.com/urfave/cli/v2"
)

func RegisterGitHooks() *cli.Command {
	return &cli.Command{
		Name:  "githook",
		Usage: "commands to manage Headers git hooks",
		Subcommands: []*cli.Command{
			registerGitHookInstall(),
		},
	}
}

