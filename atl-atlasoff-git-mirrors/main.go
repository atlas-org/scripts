package main

import (
	"encoding/json"
	"flag"
	"os"
	"os/exec"

	"github.com/gonuts/logger"
)

var msg = logger.New("atl-mirror")
var g_fname = flag.String("f", "mirrors.json", "path to file containing a list of mirrors to sync")

type Request struct {
	Url string // git URI of mirror
	Dir string // git clone of origin
}

type Response struct {
	Request
	err error
}

func do_mirror(req Request, ch chan Response) {
	msg.Infof("==> [%s]...\n", req.Url)
	cmd := exec.Command("git", "fetch", "--all", "--tags")
	cmd.Dir = req.Dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		ch <- Response{req, err}
		return
	}

	cmd = exec.Command("git", "push", "--mirror", req.Url)
	cmd.Dir = req.Dir
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
	flag.Parse()
	f, err := os.Open(*g_fname)
	if err != nil {
		msg.Errorf("problem opening file [%s]: %v\n", *g_fname, err)
		os.Exit(1)
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	reqs := make([]Request, 0, 2)
	err = dec.Decode(&reqs)
	if err != nil {
		msg.Errorf("problem decoding JSON: %v\n", err)
		os.Exit(1)
	}

	// reqs := []Request{
	// 	{
	// 		Url: "https://:@git.cern.ch/kerberos/atlas-gaudi",
	// 		Dir: "/afs/cern.ch/user/b/binet/dev/repos/mirrors/gaudi.git",
	// 	},
	// 	{
	// 		Url: "https://:@git.cern.ch/kerberos/atlas-lcg",
	// 		Dir: "/afs/cern.ch/user/b/binet/dev/repos/mirrors/lcg.git",
	// 	},
	// }

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
