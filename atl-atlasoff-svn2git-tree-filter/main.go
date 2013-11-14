// atl-atlasoff-svn2git-tree-filter is used by "git filter-tree" to reorganize atlasoff svn/git
// repositories to move all files under the repo under Dir1/Dir01/PackageName
//
// ie:
//   # before:
//   $ ls
//   ChangeLog Package src cmt wscript
//
//   # after
//   $ ls
//   Dir1
//   $ ls Dir1/Dir01/Package
//   ChangeLog Package src cmt wscript
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	flag.Parse()
	pkgdir := flag.Arg(0)
	files, err := filepath.Glob("*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "**error**: %v\n", err)
		os.Exit(1)
	}

	tmpdir := filepath.Join("__@@ATLAS@@__", pkgdir)
	err = os.MkdirAll(tmpdir, 0755)
	if err != nil {
		fmt.Fprintf(os.Stderr, "**error**: %v\n", err)
		os.Exit(1)
	}

	for _, file := range files {
		if strings.HasPrefix(file, ".git") {
			continue
		}
		err = os.Rename(file, filepath.Join(tmpdir, file))
		if err != nil {
			fmt.Fprintf(os.Stderr, "**error**: %v\n", err)
			os.Exit(1)
		}
	}

	dirs, err := filepath.Glob("__@@ATLAS@@__/*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "**error**: %v\n", err)
		os.Exit(1)
	}

	for _, dir := range dirs {
		err = os.Rename(dir, dir[len("__@@ATLAS@@__/"):])
		if err != nil {
			fmt.Fprintf(os.Stderr, "**error**: %v\n", err)
			os.Exit(1)
		}
	}

	err = os.Remove("__@@ATLAS@@__")
	if err != nil {
		fmt.Fprintf(os.Stderr, "**error**: %v\n", err)
		os.Exit(1)
	}

}
