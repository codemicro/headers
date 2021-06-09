package headers

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	"github.com/pelletier/go-toml/v2"
)

type Spec struct {
	Regex      string
	compRegexp *regexp.Regexp
	
	Comment    string
	EndComment string
}

type Options struct {
	FullFilepath bool
	ExcludeRegex string
}

type Config struct {
	HeaderText  string
	Spec        []*Spec
	specHeaders []string
	Options     *Options
}

func LoadConfigFromFile(filename string) (*Config, error) {

	if filename == "" {

		if Verbose {
			logger.Println("No config filename provded, searching parent directories")
		}

		fn, err := locateFile("headers.toml")
		if err != nil {
			return nil, err
		}
		filename = fn
	}

	if Verbose {
		logger.Printf("Loading config file %s\n", filename)
	}

	fcont, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	cfg := new(Config)

	err = toml.Unmarshal(fcont, cfg)
	if err != nil {
		// if pkx, ok := err.(*toml.DecodeError); ok {
		// 	return err
		// }
		return nil, err
	}

	cfg.makeSpecHeaders()
	err = cfg.makeSpecRegexp()
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

// locateFile walks up parents directories and searches for a
// headers.toml file, returning an error if one cannot be found
func locateFile(filename string) (string, error) {

	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	searchDir := cwd

	for searchDir != "/" {

		if Verbose {
			logger.Printf("Searching for config file in %s\n", searchDir)
		}

		fp := filepath.Join(searchDir, filename)
		if _, err := os.Stat(fp); err == nil {
			// file found in this directory
			if Verbose {
				logger.Printf("Found for config file %s\n", fp)
			}
			return fp, nil
		}
		
		searchDir = filepath.Dir(searchDir)

	}

	return "", errors.New("could not find configuration file")
}

// makeSpecHeaders generates header strings for each Spec in Config
func (cfg *Config) makeSpecHeaders() {
	cfg.specHeaders = make([]string, len(cfg.Spec))
	for i, spec := range cfg.Spec {
		cfg.specHeaders[i] = transformHeaderBySpec(cfg.HeaderText, spec)
	}
}

// makeSpecRegexp generates the compRegexp field for each Spec in Config
func (cfg *Config) makeSpecRegexp() error {

	for _, spec := range cfg.Spec {
		comp, err := regexp.Compile(spec.Regex)
		if err != nil {
			return err
		}
		spec.compRegexp = comp
	}

	return nil
}
