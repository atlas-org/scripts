package main

import (
	"bufio"
	"fmt"
	"io"
	"sort"
	"strings"
)

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

func newRelease(r io.Reader) (*Release, error) {
	var err error
	release := &Release{
		PackageDb: make(PackageDb),
	}

	scan := bufio.NewScanner(r)
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
		if err != nil {
			return nil, err
		}
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
		return nil, err
	}

	return release, nil
}

// EOF
