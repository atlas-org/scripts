package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/gonuts/logger"
)

var msg = logger.New("lcgcmt")

type Release struct {
	Toolchain Toolchain
	PackageDb PackageDb
}

func (rel *Release) String() string {
	lines := make([]string, 0, 32)
	lines = append(
		lines,
		"Release{",
		fmt.Sprintf("\tToolchain: %#v,", rel.Toolchain),
		fmt.Sprintf("\tPackages: ["),
	)

	keys := make([]string, 0, len(rel.PackageDb))
	for id := range rel.PackageDb {
		keys = append(keys, id)
	}
	sort.Strings(keys)

	for _, id := range keys {
		pkg := rel.PackageDb[id]
		lines = append(
			lines,
			fmt.Sprintf("\t\t%v-%v deps=%v,", pkg.Name, pkg.Version, pkg.Deps),
		)
	}
	lines = append(
		lines,
		"\t],",
		"}",
	)
	return strings.Join(lines, "\n")
}

type Toolchain struct {
	Name    string
	Version string
}

type PackageDb map[string]*Package

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
		return nil, fmt.Errorf(
			"invalid fields size (got %d. expected %d): %#v",
			len(fields), NFIELDS, fields,
		)
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
