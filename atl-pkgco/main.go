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
	flag.Parse()

	if *g_fname == "" && flag.NArg() <= 0 {
		errorf("you need to give a package name or a file containing a list of packages\n")
		flag.Usage()
		os.Exit(1)
	}

	pkgs := make([]string, 0)
	if *g_fname != "" {
		f, err := os.Open(*g_fname)
		if err != nil {
			errorf("could not open file [%s]: %v\n", *g_fname, err)
			os.Exit(1)
		}
		defer f.Close()
		scan := bufio.NewScanner(f)
		for scan.Scan() {
			pkgs = append(pkgs, scan.Text())
		}
		err = scan.Err()
		if err != nil {
			errorf("problem parsing file [%s]: %v\n", *g_fname, err)
			os.Exit(1)
		}
	} else {
		pkgs = append(pkgs, flag.Args()...)
	}

	ch := make(chan response)
	for _, pkg := range pkgs {
		go checkout(pkg, ch)
	}

	errs := []response{}
	for _ = range pkgs {
		resp := <-ch
		if resp.err != nil {
			errs = append(errs, resp)
		}
	}
	close(ch)

	if len(errs) != 0 {
		errorf("problem(s) checking out package(s):\n")
		for _, err := range errs {
			errorf("%s (%s): %v\n", err.pkg, err.tag, err.err)
		}
		os.Exit(1)
	}
}

func checkout(pkg string, ch chan response) {
	var err error
	tag := ""
	// if - in pkg, tag was given
	if strings.Count(pkg, "-") > 0 {
		tag = filepath.Base(pkg)
		pkg = strings.SplitN(pkg, "-", 1)[0]
	}
	
	// if no '/' in pkg, need to find full package name
	if strings.Count(pkg, "/") <= 0 {
		p, err := cmt.Package(pkg)
		if err != nil {
			ch <- response{pkg, tag, err}
			return
		}
		pkg = p.Name
	}

	// remove leading '/' for cmt
	pkg = strings.TrimLeft(pkg, "/")
	
	// special case of Gaudi packages
	if strings.HasPrefix(pkg, "Gaudi") {
		
		return
	}

	// atlasoff packages
	debugf("checkout: %s (%s)\n", pkg, tag)
	ch <- response{pkg, tag, err}
}

// EOF
