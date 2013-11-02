// atl-get-tag-diff gets the list of tag differences between 2 releases (CERN centric)
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/atlas-org/cmt"
)

func main() {
	v := flag.Bool("v", false, "enable verbose mode")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(
			os.Stderr,
			`$ %s [options] old-setup new-setup

ex:
 $ %s rel1,devval rel2,devval
 $ %s 19.0.0 rel2,devval

options:
`,
			os.Args[0], os.Args[0], os.Args[0],
		)
		flag.PrintDefaults()

	}
	flag.Parse()

	old := ""
	new := ""

	switch flag.NArg() {
	default:
		fmt.Fprintf(os.Stderr, "**error** you need to give 2 releases/nightlies setup-strings\n")
		flag.Usage()
		os.Exit(1)
	case 2:
		old = flag.Args()[0]
		new = flag.Args()[1]
	}

	diffs, err := cmt.TagDiff(old, new, *v)
	if err != nil {
		fmt.Fprintf(os.Stderr, "**error** %v\n", err)
		os.Exit(1)
	}
	if len(diffs) > 0 {
		os.Exit(1)
	}
	os.Exit(0)
}
