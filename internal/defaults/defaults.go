package defaults

import "github.com/urfave/cli/v2"

func RegisterDefault() *cli.Command {
	return &cli.Command{
			Name:    "defaults",
			Usage:   "create a new, default configuration",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    outputFilenameFlag,
					Usage:   fmt.Sprintf("file to write git hook to (default: \"%s\")", defaultOutputFilename),
					Aliases: []string{"o"},
				},
			},
			Action: gitHookInstall,
		}
	}
}
