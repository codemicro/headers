package hooks

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"
)

const githookContent = "#!/bin/sh\nfiles=`git diff --name-only --cached`\n./headers --lint $files\n"

func registerGitHookInstall() *cli.Command {

	const (
		outputFilenameFlag         = "output"
		overwriteExistingFilesFlag = "overwrite"

		defaultOutputFilename = ".git/hooks/pre-commit"
	)

	gitHookInstall := func(ctx *cli.Context) error {

		var outputFilename string
		if x := ctx.String(outputFilenameFlag); x != "" {
			outputFilename = x
		} else {
			outputFilename = defaultOutputFilename
		}

		if _, err := os.Stat(outputFilename); err == nil && !ctx.Bool(overwriteExistingFilesFlag) {
			return cli.Exit("Target output file already exists. Select another output file using --output or forcibly overwrite the current file using --overwrite", 1)
		} else if err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}

		dir := filepath.Dir(outputFilename)
		err := os.MkdirAll(dir, os.ModeDir)
		if err != nil {
			return err
		}

		err = ioutil.WriteFile(outputFilename, []byte(githookContent), 0644)
		if err != nil {
			return err
		}

		fmt.Printf("Installed Git hook to %s\n", outputFilename)

		return nil
	}

	return &cli.Command{
		Name:    "install",
		Aliases: []string{"i"},
		Usage:   "install pre-commit git lint hook",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    outputFilenameFlag,
				Usage:   fmt.Sprintf("file to write git hook to (default: \"%s\")", defaultOutputFilename),
				Aliases: []string{"o"},
			},
			&cli.BoolFlag{
				Name:  overwriteExistingFilesFlag,
				Usage: "overwrite preexisting files",
			},
		},
		Action: gitHookInstall,
	}
}
