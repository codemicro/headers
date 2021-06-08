package main

import (
	"flag"
	"io/ioutil"

	"github.com/codemicro/headers/pkg/headers"
	"github.com/pelletier/go-toml/v2"
)

var inputFile = flag.String("i", "headers.toml", "input file to read configuration from")

func main() {

	flag.Parse()

	fcont, err := ioutil.ReadFile(*inputFile)
	if err != nil {
		panic(err)
	}

	cfg := new(headers.Config)
	err = toml.Unmarshal(fcont, cfg)
	if err != nil {
		if pkx, ok := err.(*toml.DecodeError); ok {
			panic(pkx.String())
		}
		panic(err)
	}

	err = headers.Run(cfg, flag.Args())
	if err != nil {
		panic(err)
	}

}
