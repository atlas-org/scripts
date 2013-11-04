package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	gocmt "github.com/atlas-org/cmt"
)

var g_verbose = flag.Bool("v", false, "enable verbose output")
var g_fname = flag.String("f", "store.cmt", "path to file to load the environment from")
var g_oname = flag.String("o", "", "shell file to hold the environment")
var g_shell = flag.String("sh", "sh", "shell type (sh|csh)")

func main() {
	flag.Parse()

	if *g_verbose {
		fmt.Fprintf(os.Stderr, "::: loading up a CMT environment...\n")
	}

	switch *g_shell {
	case "sh", "csh":
		// ok
	default:
		fmt.Fprintf(os.Stderr, "**error** invalid shell mode. got [%s]. valid ones: %v\n", *g_shell, "sh|csh")
		flag.Usage()
		os.Exit(1)
	}

	var err error
	var out io.Writer = os.Stdout
	if *g_oname != "" {
		if *g_oname == "-" {
			out = os.Stdout
		} else {
			var f *os.File
			f, err = os.Create(*g_oname)
			if err != nil {
				fmt.Fprintf(os.Stderr, "**error** opening file [%s]: %v\n", *g_oname, err)
				os.Exit(1)
			}
			defer f.Close()
			out = f
		}
	}

	setup, err := gocmt.NewSetupFromCache(*g_fname, *g_verbose)
	if err != nil {
		fmt.Fprintf(os.Stderr, "**error** opening cache [%s]: %v\n", *g_fname, err)
		os.Exit(1)
	}
	defer setup.Delete()

	export := map[string]string{
		"sh":  "export",
		"csh": "set",
	}[*g_shell]

	eq := map[string]string{
		"sh":  "=",
		"csh": " ",
	}[*g_shell]

	for k, v := range setup.EnvMap() {
		if k == "_" {
			continue
		}
		_, err = fmt.Fprintf(out, fmt.Sprintf("%s %s%s%q\n", export, k, eq, v))
		if err != nil {
			fmt.Fprintf(
				os.Stderr, "**error** generating shell script [%s]: %v\n",
				*g_fname,
				err,
			)
			os.Exit(1)
		}
	}

}

// EOF
