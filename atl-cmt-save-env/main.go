package main

import (
	"flag"
	"fmt"
	"os"

	gocmt "github.com/atlas-org/cmt"
)

var g_verbose = flag.Bool("v", false, "enable verbose output")
var g_fname = flag.String("f", "store.cmt", "path to file where to store the environment")

func main() {
	flag.Parse()

	if *g_verbose {
		fmt.Printf("::: setting up a CMT environment...\n")
	}

	tags := ""
	switch flag.NArg() {
	case 1:
		tags = flag.Args()[0]
	default:
		fmt.Fprintf(os.Stderr, "%s needs an asetup-compatible set of tags\n", os.Args[0])
		flag.Usage()
		os.Exit(1)
	}

	setup, err := gocmt.NewSetup(tags, *g_verbose)
	if err != nil {
		panic(err)
	}
	defer setup.Delete()

	cmt, err := gocmt.New(setup)
	if err != nil {
		panic(err)
	}

	if *g_verbose {
		fmt.Printf("::: storing CMT environment into [%s]...\n", *g_fname)
	}

	f, err := os.Create(*g_fname)
	if err != nil {
		fmt.Fprintf(os.Stderr, "**error** opening file [%s]: %v\n", *g_fname, err)
		os.Exit(1)
	}
	defer f.Close()

	err = setup.Save(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "**error** saving file [%s]: %v\n", *g_fname, err)
		os.Exit(1)
	}

	if *g_verbose {
		fmt.Printf("::: storing CMT environment into [%s]... [done]\n", *g_fname)
	}
}

// EOF
