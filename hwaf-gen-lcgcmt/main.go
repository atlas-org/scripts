package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/gonuts/logger"
)

var msg = logger.New("lcgcmt")
var g_out = flag.String("o", "hscript.py", "path to hscript.py file to generate")

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(
			os.Stderr,
			`$ %s [options] path/to/lcgcmt.txt

ex:
 $ %s /afs/cern.ch/sw/lcg/experimental/LCG-preview/LCG_x86_64-slc6-gcc48-opt.txt

options:
`,
			os.Args[0], os.Args[0],
		)
		flag.PrintDefaults()
	}

}

func handle_err(err error) {
	if err != nil {
		msg.Errorf("%v\n", err)
		os.Exit(1)
	}
}

func main() {
	flag.Parse()
	fname := flag.Arg(0)

	f, err := os.Open(fname)
	handle_err(err)
	defer f.Close()

	release, err := newRelease(f)
	handle_err(err)
	msg.Debugf("%v\n", release)

	out, err := os.Create(*g_out)
	handle_err(err)
	defer out.Close()

	err = render(release, out)
	handle_err(err)
}
