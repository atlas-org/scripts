package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/gonuts/logger"
)

var msg = logger.New("lcgcmt")

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

	release := &Release{
		PackageDb: make(PackageDb),
	}

	scan := bufio.NewScanner(f)
	for scan.Scan() {
		line := scan.Text()
		line = strings.Trim(line, " \r\n\t")
		if line == "" {
			continue
		}
		if line[0] == '#' {
			continue
		}
		if strings.HasPrefix(line, "COMPILER: ") {
			fields := strings.Split(line, " ")
			name := fields[1]
			vers := fields[2]
			release.Toolchain = Toolchain{Name: name, Version: vers}
			continue
		}

		fields := strings.Split(line, ";")
		pkg, err := newPackage(fields)
		handle_err(err)
		msg.Debugf("%v (%v) %v\n", pkg.Name, pkg.Version, pkg.Deps)

		if _, exists := release.PackageDb[pkg.Id]; exists {
			handle_err(
				fmt.Errorf("package %v already in package-db:\nold: %#v\nnew: %#v\n",
					pkg.Id,
					release.PackageDb[pkg.Id],
					pkg,
				),
			)
		}
		release.PackageDb[pkg.Id] = pkg
	}

	err = scan.Err()
	if err != nil && err != io.EOF {
		handle_err(err)
	}

	msg.Infof("%v\n", release)
}
