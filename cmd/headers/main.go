package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/codemicro/headers/pkg/headers"
)

var inputFile = flag.String("i", "headers.toml", "input file to read configuration from")

func e(err error) {
	fmt.Fprintln(os.Stderr, "ERROR:", err)
	os.Exit(1)
}

func main() {

	flag.Parse()

	cfg, err := headers.LoadConfigFromFile(*inputFile)
	if err != nil {
		e(err)
	}

	err = headers.Run(cfg, flag.Args())
	if err != nil {
		e(err)
	}

}
