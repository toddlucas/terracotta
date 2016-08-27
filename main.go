package main

import (
	"flag"
	"fmt"

	"github.com/toddlucas/terracotta/pre"
)

func main() {
	var defines symbols
	var undefs symbols

	flag.Var(&defines, "define", "Define one or more preprocessor symbols")
	flag.Var(&undefs, "undef", "Undefine one or more preprocessor symbols")

	source := flag.String("source", ".", "The source directory")
	output := flag.String("output", ".", "The output directory")
	version := flag.Bool("version", false, "The version")

	flag.Parse()

	if *version {
		fmt.Println(pre.GetVersion())
		return
	}

	p := pre.Preprocessor{}

	p.ProcessDirectory(*source, *output, defines, undefs)
}
