package commands

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"
)

func registerFileWriter(defaultOutputFilename, commandName, description string, alias []string, fileContent []byte) *cli.Command {

	const (
		outputFilenameFlag         = "output"
		overwriteExistingFilesFlag = "overwrite"
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

		err = ioutil.WriteFile(outputFilename, fileContent, 0644)
		if err != nil {
			return err
		}

		fmt.Printf("Written to %s\n", outputFilename)

		return nil
	}

	return &cli.Command{
		Name:    commandName,
		Aliases: alias,
		Usage:   description,
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
