// https://github.com/codemicro/headers
// Copyright (c) 2021, codemicro and contributors
// SPDX-License-Identifier: MIT
// Filename: internal/headers/headers.go

package headers

import (
	"bytes"
	"errors"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

var LintMode bool
var Verbose bool

var logger = log.New(os.Stderr, "", 0)

type transformation struct {
	Filename string
	// NewFileContents will be nil if no updates need to be made
	NewFileContents []byte
}

func (tf *transformation) apply(permCode int) error {
	if tf.NewFileContents == nil {
		return nil
	}
	return ioutil.WriteFile(tf.Filename, tf.NewFileContents, fs.FileMode(permCode))
}

func Run(cfg *Config, cmdLineFiles []string) error {

	var files []string
	var err error
	if len(cmdLineFiles) == 0 {
		files, err = discoverFiles(".", cfg)
		if err != nil {
			return nil
		}
	} else {
		filtered, err := filterFileList(cmdLineFiles, cfg)
		if err != nil {
			return err
		}
		files = filtered
	}

	if Verbose {
		logger.Println("Running against files:", files)
	}

	var transformations []*transformation

	for _, fname := range files {
		tf, err := generateUpdateTransformation(fname, cfg)
		if err != nil {
			return err
		}
		transformations = append(transformations, tf)
	}

	return executeTransformations(transformations)
}

func Replace(cfg *Config, cmdLineFiles []string, newHeader string) error {

	var files []string
	var err error
	if len(cmdLineFiles) == 0 {
		files, err = discoverFiles(".", cfg)
		if err != nil {
			return nil
		}
	} else {
		filtered, err := filterFileList(cmdLineFiles, cfg)
		if err != nil {
			return err
		}
		files = filtered
	}

	if Verbose {
		logger.Println("Running against files:", files)
	}

	var transformations []*transformation

	for _, fname := range files {
		tf, err := generateReplaceTransformation(newHeader, fname, cfg)
		if err != nil {
			return err
		}
		transformations = append(transformations, tf)
	}

	return executeTransformations(transformations)

}

func executeTransformations(transformations []*transformation) error {

	if Verbose {
		s := "Applying transformations"
		if LintMode {
			s = "Checking transformations"
		}
		logger.Println(s)
	}

	var hasLintingFailed bool
	for _, tf := range transformations {
		if tf.NewFileContents == nil {
			if Verbose {
				logger.Printf("%s: no action required\n", tf.Filename)
			}
			continue
		}

		if LintMode {
			logger.Printf("LINT: %s has not had file headers applied\n", tf.Filename)
			hasLintingFailed = true
			continue
		}

		if Verbose {
			logger.Printf("%s: updating file content\n", tf.Filename)
		}
		err := tf.apply(0644)
		if err != nil {
			return err
		}
	}

	if LintMode {
		if hasLintingFailed {
			return errors.New("linting did not pass")
		} else {
			logger.Println("LINT: ok")
		}
	}

	return nil
}

func makeFileRegexp(cfg *Config) (func(string) bool, error) {
	var bx []string
	for _, rg := range cfg.Spec {
		bx = append(bx, rg.Regex)
	}
	var err error
	matchRegexp, err := regexp.Compile("(" + strings.Join(bx, ")|(:?") + ")")
	if err != nil {
		return nil, err
	}

	if cfg.Options.ExcludeRegex != "" {
		excludeRegexp, err := regexp.Compile(cfg.Options.ExcludeRegex)
		if err != nil {
			return nil, err
		}
		return func(mx string) bool {
			return matchRegexp.MatchString(mx) && !excludeRegexp.MatchString(mx)
		}, nil
	}

	return matchRegexp.MatchString, nil
}

func discoverFiles(dir string, cfg *Config) ([]string, error) {

	var o []string

	matchRegexp, err := makeFileRegexp(cfg)
	if err != nil {
		return nil, err
	}

	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if matchRegexp(path) {
			o = append(o, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return o, nil
}

func filterFileList(flist []string, cfg *Config) ([]string, error) {
	matchRegexp, err := makeFileRegexp(cfg)
	if err != nil {
		return nil, err
	}

	var n int
	for _, item := range flist {
		if matchRegexp(item) {
			flist[n] = item
			n += 1
		}
	}
	flist = flist[:n]
	return flist, nil
}

// transformHeaderBySpec adds the spec comment to the beginning of each line in header
func transformHeaderBySpec(header string, spec *Spec) string {
	lines := strings.Split(header, "\n")
	newLines := make([]string, len(lines))
	for i, line := range lines {
		newLines[i] = spec.Comment + " " + line
		if spec.EndComment != "" {
			newLines[i] += " " + spec.EndComment
		}
	}
	return strings.Join(newLines, "\n")
}

// splitByTemplateLiteral splits a string by any `{{ .Var }}` style template literals
// This will produce weird output when run against invalid templates, run this through
func splitByTemplateLiteral(src string) []string {
	x := strings.Split(src, "{{")
	var y []string
	for _, z := range x {
		zx := strings.Split(z, "}}")
		var zxy string
		if len(zx) == 1 {
			zxy = zx[0]
		} else if len(zx) <= 2 {
			zxy = zx[1]
		}
		y = append(y, zxy)
	}
	return y
}

// replaceTemplateLiteralsWithRegexp replaces all `{{ .Var }}` style template literals
// with a replacement regexp. Any areas in src that aren't replaced are regexp escaped.
func replaceTemplateLiteralsWithRegexp(src, replacement string) string {
	x := splitByTemplateLiteral(src)
	for i, item := range x {
		x[i] = regexp.QuoteMeta(item)
	}
	return strings.Join(x, replacement)
}

func renderHeader(headerText string, tplContent *headerTemplate) (*bytes.Buffer, error) {
	tpl, err := template.New("").Parse(headerText)
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	err = tpl.Execute(buf, tplContent)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func getFileInfo(fpath string, cfg *Config) (originalFileContents []byte, templateFilename string, spec *Spec, err error) {

	originalFileContents, err = ioutil.ReadFile(fpath)
	if err != nil {
		return nil, "", nil, err
	}

	if cfg.Options.FullFilepath {
		templateFilename = fpath
	} else {
		templateFilename = filepath.Base(fpath)
	}

	for _, s := range cfg.Spec {
		if s.compRegexp.MatchString(fpath) {
			spec = s
			break
		}
	}

	return
}

func generateUpdateTransformation(fpath string, cfg *Config) (*transformation, error) {

	originalFileContents, templateFilename, chosenSpec, err := getFileInfo(fpath, cfg)
	if err != nil {
		return nil, err
	}

	if chosenSpec == nil {
		logger.Printf("WARN: Cannot find spec for file '%s'\n", fpath)
		return nil, nil
	}

	newFileContents, err := applyHeaderToBytes(templateFilename, cfg.HeaderText, originalFileContents, chosenSpec)
	if err != nil {
		return nil, err
	}

	if bytes.Equal(originalFileContents, newFileContents) {
		return &transformation{
			Filename: fpath,
		}, nil
	}

	return &transformation{
		Filename:        fpath,
		NewFileContents: newFileContents,
	}, nil
}

func applyHeaderToBytes(fname, header string, file []byte, spec *Spec) ([]byte, error) {

	specHeader := transformHeaderBySpec(header, spec)

	rendered, err := renderHeader(specHeader, newHeaderTemplate(fname))
	if err != nil {
		return nil, err
	}

	rxp, err := regexp.Compile(replaceTemplateLiteralsWithRegexp(specHeader, "(.+)"))
	if err != nil {
		return nil, err
	}

	var newFile []byte

	if !rxp.Match(file) {
		newFile = append(rendered.Bytes(), []byte("\n\n")...)
		newFile = append(newFile, file...)
	} else {
		var firstDone bool
		newFile = rxp.ReplaceAllFunc(file, func(b []byte) []byte {
			if firstDone {
				return []byte{} // removes extra/old license headers
			}
			firstDone = true
			return rendered.Bytes()
		})
	}

	return newFile, nil
}

func generateReplaceTransformation(newHeader, fpath string, cfg *Config) (*transformation, error) {

	originalFileContents, templateFilename, chosenSpec, err := getFileInfo(fpath, cfg)
	if err != nil {
		return nil, err
	}

	if chosenSpec == nil {
		logger.Printf("WARN: Cannot find spec for file '%s'\n", fpath)
		return nil, nil
	}

	newFileContents, err := replaceHeaderInBytes(originalFileContents, templateFilename, cfg.HeaderText, newHeader, chosenSpec)
	if err != nil {
		return nil, err
	}

	if bytes.Equal(originalFileContents, newFileContents) {
		return &transformation{
			Filename: fpath,
		}, nil
	}

	return &transformation{
		Filename:        fpath,
		NewFileContents: newFileContents,
	}, nil
}

func replaceHeaderInBytes(file []byte, fname, oldHeader, newHeader string, spec *Spec) ([]byte, error) {

	oldSpecHeader := transformHeaderBySpec(oldHeader, spec)
	newSpecHeader := transformHeaderBySpec(newHeader, spec)

	newRendered, err := renderHeader(newSpecHeader, newHeaderTemplate(fname))
	if err != nil {
		return nil, err
	}

	rxp, err := regexp.Compile(replaceTemplateLiteralsWithRegexp(oldSpecHeader, "(.+)"))
	if err != nil {
		return nil, err
	}

	if rxp.Match(file) {
		return rxp.ReplaceAll(file, newRendered.Bytes()), nil
	}

	return nil, nil
}
