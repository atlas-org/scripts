package main

import (
	"bufio"
	"fmt"
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"

	gocmt "github.com/atlas-org/cmt"
)

var g_fname = flag.String("f", "", "file containing package names/tags to checkout")
var g_head = flag.Bool("A", false, "checkout package HEAD/trunk/master")
var g_dry = flag.Bool("s", false, "dry run. don't checkout anything")
var g_recent = flag.Bool("r", false, "show recent packages. don't checkout anything")
var g_checkout = true

var cmt *gocmt.Cmt

var msg = log.New(os.Stderr, "pkgco ", 0)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(
			os.Stderr,
			`$ %s [options] PACKAGE

ex:
 $ %s AthenaServices-00-01-02
 $ %s Control/AthenaServices-00-01-02
 $ %s Control/AthenaServices
 $ %s -f pkg-list.txt

options:
`,
			os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0],
		)
		flag.PrintDefaults()
	}

	var err error
	cmt, err = gocmt.New(nil)
	if err != nil {
		errorf("could not initialize Cmt instance: %v\n", err)
		//os.Exit(1)
	}
}

func errorf(format string, args ...interface{}) {
	msg.Printf("ERROR    "+format, args...)
}

func warnf(format string, args ...interface{}) {
	msg.Printf("WARNING  "+format, args...)
}

func infof(format string, args ...interface{}) {
	msg.Printf("INFO     "+format, args...)
}

func debugf(format string, args ...interface{}) {
	msg.Printf("DEBUG    "+format, args...)
}

type response struct {
	pkg string
	tag string
	err error
}

func main() {
	fmt.Printf("::: atl-pkgco...\n")
}

// EOF
