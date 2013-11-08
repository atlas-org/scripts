package main

import (
	"os"
	"os/exec"

	"github.com/gonuts/logger"
)

var msg = logger.New("atl-mirror")

type Request struct {
	url string // git URI of mirror
	dir string // git clone of origin
}

type Response struct {
	Request
	err error
}

func do_mirror(req Request, ch chan Response) {
	msg.Infof("==> [%s]...\n", req.url)
	cmd := exec.Command("git", "fetch", "--all", "--tags")
	cmd.Dir = req.dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		ch <- Response{req, err}
		return
	}

	cmd = exec.Command("git", "push", "--mirror", req.url)
	cmd.Dir = req.dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		ch <- Response{req, err}
		return
	}

	ch <- Response{req, err}
}

func main() {
	reqs := []Request{
		{
			url: "https://:@git.cern.ch/kerberos/atlas-gaudi",
			dir: "/afs/cern.ch/user/b/binet/dev/repos/mirrors/gaudi.git",
		},
		{
			url: "https://:@git.cern.ch/kerberos/atlas-lcg",
			dir: "/afs/cern.ch/user/b/binet/dev/repos/mirrors/lcg.git",
		},
	}

	resp := make(chan Response)
	for _, req := range reqs {
		go do_mirror(req, resp)
	}

	for _ = range reqs {
		r := <-resp
		if r.err != nil {
			os.Exit(1)
		}
	}
}
