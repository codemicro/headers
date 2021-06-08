package headers

import (
	"regexp"
)

type Spec struct {
	Regex      string
	Comment    string
	compRegexp *regexp.Regexp
}

type Options struct {
	FullFilepath bool
	Include string
	Exclude string
}

type Config struct {
	HeaderText  string
	Spec        []*Spec
	specHeaders []string
	Options     *Options
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
