package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	gocmt "github.com/atlas-org/cmt"
	"github.com/gonuts/logger"
)

var g_fname = flag.String("f", "", "file containing package names/tags to checkout")
var g_head = flag.Bool("A", false, "checkout package HEAD/trunk/master")
var g_dry = flag.Bool("s", false, "dry run. don't checkout anything")
var g_recent = flag.Bool("r", false, "show recent packages. don't checkout anything")
var g_checkout = true

var cmt *gocmt.Cmt

var msg = logger.New("pkgco")

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
		msg.Errorf("could not initialize Cmt instance: %v\n", err)
		os.Exit(1)
	}
}

type response struct {
	pkg string
	tag string
	err error
}

func main() {
	flag.Parse()

	if *g_fname == "" && flag.NArg() <= 0 {
		msg.Errorf("you need to give a package name or a file containing a list of packages\n")
		flag.Usage()
		os.Exit(1)
	}

	if *g_dry || *g_recent {
		g_checkout = false
	}

	pkgs := make([]string, 0)
	if *g_fname != "" {
		f, err := os.Open(*g_fname)
		if err != nil {
			msg.Errorf("could not open file [%s]: %v\n", *g_fname, err)
			os.Exit(1)
		}
		defer f.Close()
		scan := bufio.NewScanner(f)
		for scan.Scan() {
			txt := strings.Trim(scan.Text(), " \r\n")
			if strings.HasPrefix(txt, "#") {
				continue
			}
			pkgs = append(pkgs, scan.Text())
		}
		err = scan.Err()
		if err != nil {
			msg.Errorf("problem parsing file [%s]: %v\n", *g_fname, err)
			os.Exit(1)
		}
	} else {
		pkgs = append(pkgs, flag.Args()...)
	}

	nch := len(pkgs)
	if nch > 8 {
		nch = 8
	}
	ch := make(chan response, len(pkgs))
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
		msg.Errorf("problem(s) checking out package(s):\n")
		for _, err := range errs {
			msg.Errorf("%s (%s): %v\n", err.pkg, err.tag, err.err)
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
		pkg = strings.SplitN(pkg, "-", 2)[0]
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
		env := os.Environ()
		gaudisvn := os.Getenv("GAUDISVN")
		if gaudisvn == "" {
			gaudisvn = "http://svnweb.cern.ch/guest/gaudi"
			env = append(env, "GAUDISVN="+gaudisvn)
		}
		svnroot := gaudisvn + "/Gaudi"
		svntrunk := "trunk"
		svntags := "tags"
		env = append(env, "SVNROOT="+svnroot)
		env = append(env, "SVNTRUNK="+svntrunk)
		env = append(env, "SVNTAGS="+svntags)
		env = append(env, "pkg="+pkg)

		args := []string{"co"}
		if *g_head {
			args = append(
				args,
				fmt.Sprintf("%s/%s/%s", svnroot, svntrunk, pkg),
				pkg,
			)
		} else {
			if len(tag) == 0 {
				tag = cmt.PackageVersion(pkg)
				if tag == "" {
					ch <- response{pkg, tag, fmt.Errorf("could not find any tag for %q", pkg)}
					return
				}
			}
			env = append(env, "tag="+tag)
			args = append(
				args,
				fmt.Sprintf("%s/%s/%s/%s", svnroot, svntags, pkg, tag),
				pkg,
			)
		}
		if g_checkout {
			msg.Infof("checkout: %s (%s)\n", pkg, tag)
			cmd := exec.Command("svn", args...)
			cmd.Env = env
			err = cmd.Run()
			ch <- response{pkg, tag, err}
			return
		} else {
			out := []string{tag, pkg}
			if *g_recent {
				var istrunk bool
				head, err := cmt.LatestPackageTag(pkg)
				if err != nil {
					istrunk = false
					head = "NONE"
				} else {
					istrunk = svn_tag_is_trunk(pkg, head)
				}
				eq := "=="
				switch istrunk {
				case true:
					eq = "=="
				case false:
					eq = "!="
				}
				out = append(
					out,
					fmt.Sprintf(" (most recent %s %s trunk)", head, eq),
				)
			}
			fmt.Printf("%s\n", strings.Join(out, " "))
			ch <- response{pkg, tag, nil}
		}
		return
	}

	// atlasoff packages
	if *g_head {
		msg.Infof("checkout: %s (%s)\n", pkg, "trunk")
		tag := ""
		err := cmt.CheckOut(pkg, tag)
		ch <- response{pkg, "HEAD", err}
		return
	}

	if tag == "" {
		tag = cmt.PackageVersion(pkg)
		if tag == "" {
			ch <- response{pkg, tag, fmt.Errorf("could not find any tag for %q", pkg)}
			return
		}

	}

	if g_checkout {
		msg.Infof("checkout: %s (%s)\n", pkg, tag)
		err := cmt.CheckOut(pkg, tag)
		ch <- response{pkg, tag, err}
		return
	} else {
		out := []string{tag, pkg}
		if *g_recent {
			var istrunk bool
			head, err := cmt.LatestPackageTag(pkg)
			if err != nil {
				istrunk = false
				head = "NONE"
			} else {
				istrunk = svn_tag_is_trunk(pkg, head)
			}
			eq := "=="
			switch istrunk {
			case true:
				eq = "=="
			case false:
				eq = "!="
			}
			out = append(
				out,
				fmt.Sprintf(" (most recent %s %s trunk)", head, eq),
			)
		}
		fmt.Printf("%s\n", strings.Join(out, " "))
		ch <- response{pkg, tag, nil}
	}
}

// svn_tag_is_trunk runs an SVN diff of pkg/tag with trunk
// and returns true if tag matches with trunk
func svn_tag_is_trunk(pkg, tag string) bool {
	env := os.Environ()
	svnroot := os.Getenv("SVNROOT")
	if svnroot == "" {
		msg.Errorf("SVNROOT not set\n")
		panic(fmt.Errorf("SVNROOT not set"))
	}

	tag_url := ""
	trunk_url := ""
	if strings.HasPrefix(pkg, "Gaudi") {
		gaudisvn := os.Getenv("GAUDISVN")
		if gaudisvn == "" {
			gaudisvn = "http://svnweb.cern.ch/guest/gaudi"
		}
		svnroot = strings.Join([]string{gaudisvn, "Gaudi"}, "/")
		env = append(env, "SVNTRUNK=trunk")
		env = append(env, "SVNTAGS=tags")
		tag_url = strings.Join([]string{svnroot, "tags", pkg, tag}, "/")
		trunk_url = strings.Join([]string{svnroot, "trunk", pkg}, "/")
	} else {
		tag_url = strings.Join([]string{svnroot, pkg, "tags", tag}, "/")
		trunk_url = strings.Join([]string{svnroot, pkg, "trunk"}, "/")
	}
	env = append(env, "SVNROOT="+svnroot)

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	args := []string{"diff", tag_url, trunk_url}
	cmd := exec.Command("svn", args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err := cmd.Run()
	if err != nil {
		msg.Errorf("problem running svn diff:\n%v\n", err)
		msg.Errorf("stderr:\n%v\n", string(stderr.Bytes()))
		panic(err)
	}

	return len(stdout.Bytes()) == 0
}

// EOF
