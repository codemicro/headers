// https://github.com/codemicro/headers
// Copyright (c) 2021, codemicro and contributors
// SPDX-License-Identifier: MIT
// Filename: internal/commands/commands.go

package commands

import (
	"github.com/urfave/cli/v2"
)

const githookContent = "#!/bin/sh\nfiles=`git diff --name-only --cached`\n./headers --lint $files\n"

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