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

type Package struct {
	Id        string
	Name      string
	Version   string
	ShortDest string
	FullDest  string
	Hash      string
	Deps      []string
}

func newPackage(fields []string) (*Package, error) {
	const NFIELDS = 7
	if len(fields) != NFIELDS {
		return nil, fmt.Errorf("invalid fields size (got %d. expected %s)", len(fields), NFIELDS)
	}

	pkg := &Package{
		Id:        fields[0],
		Name:      fields[1],
		Version:   fields[2],
		ShortDest: fields[4],
		FullDest:  fields[5],
		Hash:      fields[3],
		Deps:      make([]string, 0),
	}
	id := fields[1] + "-" + fields[3]
	if id != pkg.Id {
		return nil, fmt.Errorf("inconsistent fields (id=%v. name+id=%v)", pkg.Id, id)
	}

	if fields[6] != "" {
		deps := strings.Split(fields[6], ",")
		for _, dep := range deps {
			dep = strings.Trim(dep, " \r\n\t")
			if dep != "" {
				pkg.Deps = append(pkg.Deps, dep)
			}
		}
	}
	return pkg, nil
}

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
		fields := strings.Split(line, ";")
		pkg, err := newPackage(fields)
		handle_err(err)
		msg.Debugf("%v (%v) %v\n", pkg.Name, pkg.Version, pkg.Deps)
	}

	err = scan.Err()
	if err != nil && err != io.EOF {
		handle_err(err)
	}
}
