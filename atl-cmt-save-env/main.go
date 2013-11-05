package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	gocmt "github.com/atlas-org/cmt"
)

var g_verbose = flag.Bool("v", false, "enable verbose output")
var g_fname = flag.String("f", "store.cmt", "path to file where to store the environment")
var g_help = flag.Bool("h", false, "print help")

func main() {
	flag.Parse()

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(
			os.Stderr,
			`$ %s [options] ASETUP-TAGS

ex:
 $ %s             # no-arg: take env from shell
 $ %s rel1,devval
 $ %s 19.0.0
 $ %s -f my.setup.cmt 19.0.0

options:
`,
			os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0],
		)
		flag.PrintDefaults()

	}

	if *g_help {
		flag.Usage()
		os.Exit(1)
	}

	if *g_verbose {
		fmt.Printf("::: setting up a CMT environment...\n")
	}

	tags := strings.Join(flag.Args(), " ")
	setup, err := gocmt.NewSetup(tags, *g_verbose)
	if err != nil {
		fmt.Fprintf(os.Stderr, "**error** sourcing asetup: %v\n", err)
		os.Exit(1)
	}
	defer setup.Delete()

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
