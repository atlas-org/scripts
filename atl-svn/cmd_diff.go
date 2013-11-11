package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	gocmt "github.com/atlas-org/cmt"
	"github.com/gonuts/commander"
	"github.com/gonuts/flag"
	"github.com/gonuts/logger"
)

func atl_make_cmd_diff() *commander.Command {
	cmd := &commander.Command{
		Run:       atl_run_cmd_diff,
		UsageLine: "diff [options] OLD-TAG NEW-TAG",
		Short:     "diff between 2 tags or revs",
		Long: `
diff displays the diff between 2 svn tags or revs.

ex:
 $ atl-svn diff AthenaServices-00-01-02 Control/AthenaServices-00-01-03
 $ atl-svn diff AthenaServices-00-01-02 AthenaServices-00-01-03
 $ atl-svn diff AthenaServices-00-01-02 AthenaServices-HEAD
 $ atl-svn diff AthenaServices-00-01-02 AthenaServices-trunk
`,
		Flag: *flag.NewFlagSet("atl-svn-diff", flag.ExitOnError),
	}
	return cmd
}

func atl_run_cmd_diff(cmd *commander.Command, args []string) {
	var err error
	n := "atl-svn-" + cmd.Name()
	msg := logger.New(n)
	if len(args) != 2 {
		msg.Errorf("you need to give *2* tags to %s\n", n)
		flag.Usage()
		os.Exit(1)
	}

	old_tag := args[0]
	new_tag := args[1]

	cmt, err := gocmt.New(nil)
	if err != nil {
		msg.Errorf("could not initialize Cmt instance: %v\n", err)
		os.Exit(1)
	}

	var p_old *gocmt.Package
	var p_new *gocmt.Package

	if strings.Count(old_tag, "-") > 0 {
		pkg_tag := filepath.Base(old_tag)
		pkg := filepath.Base(strings.SplitN(old_tag, "-", 2)[0])
		p_old, err = cmt.Package(pkg)
		if err != nil {
			msg.Errorf("could not find package [%s]: %v\n", old_tag, err)
			os.Exit(1)
		}
		if strings.HasSuffix(pkg_tag, "-HEAD") || strings.HasSuffix(pkg_tag, "-trunk") {
			pkg_tag = "trunk"
		}
		old_tag = pkg_tag
	} else {
		msg.Errorf("invalid tag version [%s]\n", old_tag)
		os.Exit(1)
	}

	if strings.Count(new_tag, "-") > 0 {
		pkg_tag := filepath.Base(new_tag)
		pkg := filepath.Base(strings.SplitN(new_tag, "-", 2)[0])
		p_new, err = cmt.Package(pkg)
		if err != nil {
			msg.Errorf("could not find package [%s]: %v\n", new_tag, err)
			os.Exit(1)
		}
		if strings.HasSuffix(pkg_tag, "-HEAD") || strings.HasSuffix(pkg_tag, "-trunk") {
			pkg_tag = "trunk"
		}
		new_tag = pkg_tag
	} else {
		msg.Errorf("invalid tag version [%s]\n", new_tag)
		os.Exit(1)
	}

	svnroot := os.Getenv("SVNROOT")
	if svnroot == "" {
		msg.Errorf("SVNROOT not set\n")
		os.Exit(1)
	}

	url_old := fmt.Sprintf("%s/%s/%s/%s", svnroot, p_old.Name, "tags", old_tag)
	if old_tag == "trunk" {
		url_old = fmt.Sprintf("%s/%s/%s", svnroot, p_old.Name, "trunk")
	}

	url_new := fmt.Sprintf("%s/%s/%s/%s", svnroot, p_new.Name, "tags", new_tag)
	if new_tag == "trunk" {
		url_new = fmt.Sprintf("%s/%s/%s", svnroot, p_new.Name, "trunk")
	}

	svn := exec.Command("svn", "diff", url_old, url_new)
	svn.Stdout = os.Stdout
	svn.Stderr = os.Stderr
	err = svn.Run()
	if err != nil {
		msg.Errorf("error running svn-diff: %v\n", err)
		os.Exit(1)
	}

}

