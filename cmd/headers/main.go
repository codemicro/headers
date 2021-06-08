package main

import (
	"flag"

	"github.com/codemicro/headers/pkg/headers"
)

var inputFile = flag.String("i", "headers.toml", "input file to read configuration from")

func main() {

	flag.Parse()

	cfg, err := headers.LoadConfigFromFile(*inputFile)
	if err != nil {
		panic(err)
	}

	err = headers.Run(cfg, flag.Args())
	if err != nil {
		panic(err)
	}

}
