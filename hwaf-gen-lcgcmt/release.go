package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	goyaml "github.com/gonuts/yaml"
	"github.com/hwaf/gas"
)

type Release struct {
	Toolchain Toolchain
	PackageDb PackageDb
	lcg       *Package
	lcgApps   map[string]struct{}
	lcgExts   map[string]struct{}
}

func (rel *Release) String() string {
	lines := make([]string, 0, 32)
	lines = append(
		lines,
		"Release{",
		fmt.Sprintf("\tToolchain: %#v,", rel.Toolchain),
		fmt.Sprintf("\tPackages: ["),
	)

	keys := rel.PackageIds()
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

func (rel *Release) PackageNames() []string {
	pkgs := make([]string, 0, len(rel.PackageDb))
	for _, pkg := range rel.PackageDb {
		pkgs = append(pkgs, pkg.Name)
	}
	sort.Strings(pkgs)
	return pkgs
}

func (rel *Release) PackageIds() []string {
	keys := make([]string, 0, len(rel.PackageDb))
	for id := range rel.PackageDb {
		keys = append(keys, id)
	}
	sort.Strings(keys)
	return keys
}

func (rel *Release) LcgApps() []*Package {
	pkgids := make([]string, 0, len(rel.PackageDb)/2)
	for _, pkg := range rel.PackageDb {
		if rel.isLcgAppPkg(pkg) {
			pkgids = append(pkgids, pkg.Id)
		}
	}
	pkgs := make([]*Package, 0, len(pkgids))
	sort.Strings(pkgids)
	for _, id := range pkgids {
		pkgs = append(pkgs, rel.PackageDb[id])
	}
	return pkgs
}

func (rel *Release) LcgExternals() []*Package {
	pkgids := make([]string, 0, len(rel.PackageDb)/2)
	for _, pkg := range rel.PackageDb {
		if rel.isLcgExtPkg(pkg) {
			pkgids = append(pkgids, pkg.Id)
		}
	}
	sort.Strings(pkgids)
	pkgs := make([]*Package, 0, len(pkgids))
	for _, id := range pkgids {
		pkgs = append(pkgs, rel.PackageDb[id])
	}
	return pkgs
}

func (rel *Release) isLcgExtPkg(pkg *Package) bool {
	_, ok := rel.lcgExts[pkg.Name]
	return ok
}

func (rel *Release) isLcgAppPkg(pkg *Package) bool {
	_, ok := rel.lcgApps[pkg.Name]
	return ok
}

func locateLcgExtAsset(version string) (io.ReadCloser, error) {
	dir, err := gas.Abs("github.com/atlas-org/scripts/hwaf-gen-lcgcmt")
	if err != nil {
		return nil, err
	}
	fname := filepath.Join(dir, fmt.Sprintf("lcgexternals_%s.txt", version))
	f, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	return f, err
}

func newRelease(r io.Reader) (*Release, error) {
	var err error
	release := &Release{
		PackageDb: make(PackageDb),
		lcgApps:   make(map[string]struct{}),
		lcgExts:   make(map[string]struct{}),
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

		if pkg.Name == "LCGCMT" {
			release.lcg = pkg
		}
		switch pkg.Name {
		case "ROOT", "RELAX", "COOL", "CORAL":
			release.lcgApps[pkg.Name] = struct{}{}
		}
	}

	err = scan.Err()
	if err != nil && err != io.EOF {
		return nil, err
	}

	{
		f, err := locateLcgExtAsset(release.lcg.Version)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		exts := []string{}
		buf, err := ioutil.ReadAll(f)
		if err != nil {
			return nil, err
		}
		err = goyaml.Unmarshal(buf, &exts)
		if err != nil {
			return nil, err
		}

		for _, ext := range exts {
			release.lcgExts[ext] = struct{}{}
		}
	}

	return release, nil
}

// EOF
