// https://github.com/codemicro/headers
// Copyright (c) 2021, codemicro and contributors
// SPDX-License-Identifier: MIT
// Filename: cmd/headers/main.go

package main

import (
	"fmt"
	"os"

	"github.com/codemicro/headers/internal/commands"
	"github.com/codemicro/headers/internal/headers"
	"github.com/urfave/cli/v2"
)

func main() {

	app := &cli.App{
		UseShortOptionHandling: true,

		Name:  "headers",
		Usage: "apply headers to source code files",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "inputfile",
				Usage:   "input file to read configuration from (default \"headers.toml\")",
				Aliases: []string{"i"},
			},
			&cli.BoolFlag{
				Name:        "lint",
				Usage:       "enables lint mode (useful for Git hooks)",
				Destination: &headers.LintMode,
				Aliases:     []string{"l"},
			},
			&cli.BoolFlag{
				Name:        "verbose",
				Usage:       "enables verbose mode",
				Destination: &headers.Verbose,
				Aliases:     []string{"v"},
			},
		},
		Action: func(ctx *cli.Context) error {

			cfg, err := headers.LoadConfigFromFile(ctx.String("inputfile"))
			if err != nil {
				return err
			}

			err = headers.Run(cfg, ctx.Args().Slice())
			if err != nil {
				return err
			}

			return nil

		},

		Commands: []*cli.Command{
			commands.RegisterGitHooks(),
			commands.RegisterDefault(),
			commands.RegisterReplaceHeader(),
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR:", err)
		os.Exit(1)
	}

}
