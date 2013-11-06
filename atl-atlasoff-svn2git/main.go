package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gonuts/logger"
)

var msg = logger.New("svn2git")

const svn = "file:///data/binet/dev/atlasoff/svn"

func cnv(pkg string, i, nmax int) error {
	git_pkg := filepath.Join("atlasoff-git", pkg)
	start := time.Now()
	msg.Infof("[%04d/%04d] converting [%s]...\n", i, nmax, pkg)

	err := os.MkdirAll(git_pkg, 0755)
	if err != nil {
		msg.Errorf("could not create directory [%s]: %v\n", git_pkg, err)
		return err
	}

	fname := filepath.Join("logs", strings.Replace(git_pkg, "/", "-", -1)+".log.txt")
	f, err := os.Create(fname)
	if err != nil {
		msg.Errorf("could not create log-file [%s]: %v\n", fname, err)
		return err
	}
	defer f.Close()

	var out io.Writer
	if *g_verbose {
		out = io.MultiWriter(f, os.Stdout)
	} else {
		out = f
	}

	cmd := exec.Command("go-svn2git",
		"-verbose",
		"-revision", "1",
		strings.Join([]string{svn, pkg}, "/"),
	)
	cmd.Stdout = out
	cmd.Stderr = out
	cmd.Dir = git_pkg
	err = cmd.Run()
	if err != nil {
		msg.Errorf("could not run svn2git on package [%s]: %v\n", pkg, err)
		return fmt.Errorf("problem running svn2git. logfile [%s]. %v\n", fname, err)
	}
	msg.Infof("[%04d/%04d] converting [%s]... (%v)\n", i, nmax, pkg, time.Since(start))
	return err
}

var g_pkglist = flag.String("f", "", "path to file containing packages to convert")
var g_verbose = flag.Bool("v", false, "enable verbose output")
var g_njobs = flag.Int("j", 4, "number of goroutines to spawn")

func main() {
	msg.Infof("::: atl-atlasoff-svn2git\n")
	flag.Parse()

	if *g_njobs <= 0 {
		msg.Errorf("invalid number of goroutines (%d)\n", *g_njobs)
		os.Exit(1)
	}

	err := os.MkdirAll("logs", 0755)
	if err != nil {
		msg.Errorf("could not create logs dir: %v\n", err)
		os.Exit(1)
	}

	pkgs := []string{}
	if *g_pkglist != "" {
		f, err := os.Open(*g_pkglist)
		if err != nil {
			msg.Errorf("could not open package list file [%s]: %v\n", *g_pkglist, err)
			os.Exit(1)
		}
		scan := bufio.NewScanner(f)
		for scan.Scan() {
			line := scan.Text()
			line = strings.Trim(line, " \r\t\n")
			if line == "" {
				continue
			}
			if strings.HasPrefix(line, "#") {
				continue
			}
			if strings.HasSuffix(line, "/trunk/") {
				line = line[:len(line)-len("/trunk/")]
			}
			if strings.HasSuffix(line, "/trunk") {
				line = line[:len(line)-len("/trunk")]
			}
			pkgs = append(pkgs, line)
		}
		err = scan.Err()
		if err != nil {
			msg.Errorf("problem reading package list file [%s]: %v\n", *g_pkglist, err)
			os.Exit(1)
		}
	} else {
		pkgs = append(pkgs, flag.Args()...)
	}

	npkgs := len(pkgs)
	msg.Infof("# of goroutines: %d\n", *g_njobs)
	msg.Infof("# of packages:   %d\n", npkgs)
	msg.Debugf("pkgs: %v\n", pkgs)

	throttle := make(chan struct{}, *g_njobs)
	ch := make(chan error)
	for i, pkg := range pkgs {
		go func(pkg string, i int) {
			throttle <- struct{}{}
			ch <- cnv(pkg, i, npkgs)
			<-throttle
		}(pkg, i)
	}

	allgood := true
	for _ = range pkgs {
		err := <-ch
		if err != nil {
			msg.Errorf("%v\n", err)
			allgood = false
		}
	}

	if !allgood {
		os.Exit(1)
	}

	msg.Infof("::: atl-atlasoff-svn2git [done]\n")
}

// EOF
