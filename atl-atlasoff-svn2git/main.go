package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gonuts/logger"
)

var msg = logger.New("svn2git")

const svn = "file:///data/binet/dev/atlasoff/svn"

func cnv(pkg string) error {
	git_pkg := filepath.Join("atlasoff-git", pkg)
	start := time.Now()
	msg.Infof("converting [%s]...\n", pkg)

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

	cmd := exec.Command("go-svn2git",
		"-verbose",
		"-revision", "1",
		strings.Join([]string{svn, pkg}, "/"),
	)
	cmd.Stdout = f
	cmd.Stderr = f
	cmd.Dir = git_pkg
	err = cmd.Run()
	if err != nil {
		msg.Errorf("could not run svn2git on package [%s]: %v\n", pkg, err)
		return fmt.Errorf("problem running svn2git. logfile [%s]. %v\n", fname, err)
	}
	msg.Infof("converting [%s]... (%v)\n", pkg, time.Since(start))
	return err
}

var g_pkglist = flag.String("f", "", "path to file containing packages to convert")

func main() {
	msg.Infof("::: atl-atlasoff-svn2git\n")
	flag.Parse()

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

	msg.Debugf("pkgs: %v\n", pkgs)

	throttle := make(chan struct{}, 10)
	ch := make(chan error, 10)
	for _, pkg := range pkgs {
		go func(pkg string) {
			throttle <- struct{}{}
			ch <- cnv(pkg)
			<-throttle
		}(pkg)
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
}

// EOF
