package main

import (
	"fmt"
	"strings"
)

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

// EOF
