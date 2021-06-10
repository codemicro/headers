// https://github.com/codemicro/headers
// Copyright (c) 2021, codemicro and contributors
// SPDX-License-Identifier: MIT
// Filename: internal/commands/commands.go

package commands

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/codemicro/headers/internal/headers"
	"github.com/urfave/cli/v2"
)

const githookContent = "#!/bin/sh\nfiles=`git diff --name-only --cached`\nheaders --lint $files\n"

func RegisterGitHooks() *cli.Command {
	return &cli.Command{
		Name:  "githook",
		Usage: "commands to manage Headers git hooks",
		Subcommands: []*cli.Command{
			registerFileWriter(".git/hooks/pre-commit", "install", "install pre-commit git lint hook", []string{"i"}, []byte(githookContent)),
		},
	}
}

const defaultContent = `headerText = """https://github.com/yourUsername/yourRepo
Copyright (c) {{ .Year }}, yourUsername and contributors
SPDX-License-Identifier: MIT
Filename: {{ .Filename }}"""

spec = [
    { regex = '^.+\.py', comment = "#" },
    { regex = '^.+\.go', comment = "//" },
    { regex = '^.+\.html', comment = "<!--", endComment = "-->" },
]

[options]
fullFilepath = true
`

func RegisterDefault() *cli.Command {
	return registerFileWriter("headers.toml", "new", "create a new default config file", nil, []byte(defaultContent))
}

func RegisterReplaceHeader() *cli.Command {
	return &cli.Command{
		Name:  "replace",
		Usage: "replace header in files - new header content should be piped through stdin",
		Action: func(ctx *cli.Context) error {
			var newHeader string
			{
				r := bufio.NewReader(os.Stdin)
				buf := make([]byte, 0, 4*1024)
				for {
					n, err := r.Read(buf[:cap(buf)])
					buf = buf[:n]
					if n == 0 {
						if err == nil {
							continue
						}
						if err == io.EOF {
							break
						}
						return err
					}

					newHeader = strings.TrimSpace(string(buf))
				}
			}

			cfg, err := headers.LoadConfigFromFile(ctx.String("inputfile"))
			if err != nil {
				return err
			}

			err = headers.Replace(cfg, ctx.Args().Slice(), newHeader)
			if err != nil {
				return err
			}

			fmt.Println("Now update your headers.toml file with the new header content.")

			return nil
		},
	}
}
