package headers

import (
	"bytes"
	"flag"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

var lintMode = flag.Bool("lint", false, "enables lint mode (useful for Git hooks)")

var allowFinish = true

func Run(cfg *Config, cmdLineFiles []string) error {

	cfg.makeSpecHeaders()
	err := cfg.makeSpecRegexp()
	if err != nil {
		return err
	}

	var files []string
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
	
	fmt.Println(files)

	for _, fname := range files {
		err = applyHeaderToFile(fname, cfg)
		if err != nil {
			return err
		}	
	}
	
	if *lintMode {
		if allowFinish {
			fmt.Fprintln(os.Stderr, "LINT: ok")
		} else {
			os.Exit(1)
		}
	}

	return nil
}

func makeFileRegexp(cfg *Config) (*regexp.Regexp, error) {
	var bx []string
	for _, rg := range cfg.Spec {
		bx = append(bx, rg.Regex)
	}
	var err error
	matchRegexp, err := regexp.Compile("(" + strings.Join(bx, ")|(:?") + ")")
	if err != nil {
		return nil, err
	}
	return matchRegexp, nil
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

		if matchRegexp.MatchString(path) {
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
		if matchRegexp.MatchString(item) {
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

func applyHeaderToFile(fpath string, cfg *Config) error {

	ogFileCont, err := ioutil.ReadFile(fpath)
	if err != nil {
		return err
	}

	var baseFilepath string
	if cfg.Options.FullFilepath {
		baseFilepath = fpath
	} else {
		baseFilepath = filepath.Base(fpath)
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
		return nil
	}

	newFileCont, err := applyHeaderToBytes(baseFilepath, cfg.HeaderText, chosenSpec, ogFileCont)
	if err != nil {
		return err
	}

	if !bytes.Equal(ogFileCont, newFileCont) {

		if *lintMode {
			fmt.Fprintf(os.Stderr, "LINT: %s has not had file headers applied\n", fpath)
			allowFinish = false
			return nil
		}

		return ioutil.WriteFile(fpath, newFileCont, 0644)
	}

	return nil
}

func applyHeaderToBytes(fname, header string, spec *Spec, file []byte) ([]byte, error) {

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
		newFile = append(rendered.Bytes(), []byte("\n")...)
		newFile = append(newFile, file...)
	} else {
		var firstDone bool
		newFile = rxp.ReplaceAllFunc(file, func(b []byte) []byte {
			if firstDone {
				return nil
			}
			firstDone = true
			return rendered.Bytes()
		})
	}

	return newFile, nil
}
