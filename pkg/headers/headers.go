package headers

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

var LintMode bool
var Verbose bool

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
		fmt.Fprintln(os.Stderr, "Running against files:", files)
	}

	var transformations []*transformation

	for _, fname := range files {
		tf, err := generateTransformationForFile(fname, cfg)
		if err != nil {
			return err
		}
		transformations = append(transformations, tf)
	}

	if Verbose {
		s := "Applying transformations"
		if LintMode {
			s = "Checking transformations"
		}
		fmt.Fprintln(os.Stderr, s)
	}

	var hasLintingFailed bool
	for _, tf := range transformations {
		if tf.NewFileContents == nil {
			if Verbose {
				fmt.Fprintf(os.Stderr, "%s: no action required\n", tf.Filename)
			}
			continue
		}

		if LintMode {
			fmt.Fprintf(os.Stderr, "LINT: %s has not had file headers applied\n", tf.Filename)
			hasLintingFailed = true
			continue
		}

		if Verbose {
			fmt.Fprintf(os.Stderr, "%s: updating file content\n", tf.Filename)
		}
		err = tf.apply(0644)
		if err != nil {
			return err
		}
	}

	if LintMode {
		if hasLintingFailed {
			return errors.New("linting did not pass")
		} else {
			fmt.Fprintln(os.Stderr, "LINT: ok")
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

func generateTransformationForFile(fpath string, cfg *Config) (*transformation, error) {

	originalFileContents, err := ioutil.ReadFile(fpath)
	if err != nil {
		return nil, err
	}

	var templateFilename string
	if cfg.Options.FullFilepath {
		templateFilename = fpath
	} else {
		templateFilename = filepath.Base(fpath)
	}

	var chosenSpec *Spec
	for _, spec := range cfg.Spec {
		if spec.compRegexp.MatchString(fpath) {
			chosenSpec = spec
			break
		}
	}

	if chosenSpec == nil {
		fmt.Fprintf(os.Stderr, "WARN: Cannot find spec for file '%s'\n", fpath)
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

	rxp, err := regexp.Compile(replaceTemplateLiteralsWithRegexp(specHeader, "(.+)"))
	if err != nil {
		return nil, err
	}

	rendered, err := renderHeader(specHeader, newHeaderTemplate(fname))
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
