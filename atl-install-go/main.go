package main

import (
	"archive/tar"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	//"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

var g_odir = flag.String("o", "/afs/cern.ch/atlas/offline/external/Go", "output directory where to install the go rungime")

func download(version string, platform [2]string) error {
	odir := filepath.Join(*g_odir, version, "tmp-"+platform[0]+"-"+platform[1])

	// create out-dir layout
	fmt.Printf("~~~ %s\n", odir)
	err := os.MkdirAll(odir, 0755)
	if err != nil {
		return err
	}

	// https://go.googlecode.com/files/go1.1.2.linux-amd64.tar.gz
	url := fmt.Sprintf(
		"https://go.googlecode.com/files/go%s.%s-%s.tar.gz",
		version,
		platform[0],
		platform[1],
	)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	gz, err := gzip.NewReader(resp.Body)
	if err != nil {
		return err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	// Iterate through the files in the archive.
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			// end of tar archive
			break
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "**error** %v\n", err)
			continue
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			dir := filepath.Join(odir, hdr.Name)
			fmt.Printf(">>> %s\n", dir)
			err = os.MkdirAll(dir, 0755)
			if err != nil {
				return err
			}
			continue

		case tar.TypeReg, tar.TypeRegA:
			// ok
		default:
			fmt.Fprintf(os.Stderr, "**error: %v\n", hdr.Typeflag)
			return err
		}
		oname := filepath.Join(odir, hdr.Name)
		fmt.Printf("::: %s (%s)\n", oname, string(byte(hdr.Typeflag)))
		dir := filepath.Dir(oname)
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}

		o, err := os.OpenFile(oname,
			os.O_WRONLY|os.O_CREATE,
			os.FileMode(hdr.Mode),
		)
		if err != nil {
			return err
		}
		defer o.Close()
		_, err = io.Copy(o, tr)
		if err != nil {
			return err
		}
		o.Sync()
		o.Close()
	}

	err = os.Rename(
		filepath.Join(odir, "go"),
		filepath.Join(*g_odir, version, platform[0]+"_"+platform[1]),
	)
	if err != nil {
		return err
	}
	err = os.RemoveAll(odir)
	if err != nil {
		return err
	}

	goroot := filepath.Join(*g_odir, version, platform[0]+"_"+platform[1])
	// create setup.sh
	setup_sh, err := os.Create(
		filepath.Join(goroot, "setup.sh"),
	)
	if err != nil {
		return err
	}
	defer setup_sh.Close()
	_, err = setup_sh.WriteString(fmt.Sprintf(`#!/bin/sh

export GOROOT=%s
export PATH=${GOROOT}/bin:${PATH}
export GOPATH=${HOME}/dev/gocode
export PATH=${GOPATH}/bin:${PATH}
`, goroot))
	if err != nil {
		return err
	}

	// create setup.csh
	setup_csh, err := os.Create(
		filepath.Join(*g_odir, version, platform[0]+"_"+platform[1], "setup.csh"),
	)
	if err != nil {
		return err
	}
	defer setup_csh.Close()
	_, err = setup_csh.WriteString(fmt.Sprintf(`
setenv GOROOT %s
setenv PATH ${GROOT}/bin:${PATH}
setenv GOPATH ${HOME}/dev/gocode
setenv PATH ${GOPATH}/bin:${PATH}
`))
	if err != nil {
		return err
	}

	return err
}

func main() {
	flag.Parse()
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `atl-install-go installs the go-gc runtime.
Usage: 

$ atl-install-go [options] <go-version>
`)
		flag.PrintDefaults()
	}

	if flag.NArg() <= 0 {
		flag.Usage()
		os.Exit(1)
	}

	version := flag.Arg(0)
	if version == "" {
		flag.Usage()
		os.Exit(1)
	}

	platforms := [][2]string{
		{"linux", "amd64"},
		{"linux", "386"},
	}

	all_good := true
	for _, plat := range platforms {
		fmt.Printf(":: installing %s-%s...\n", plat[0], plat[1])
		err := download(version, plat)
		if err != nil {
			fmt.Fprintf(os.Stderr, "**error** %v\n", err)
			all_good = false
			continue
		}
		fmt.Printf(":: installing %s-%s... [ok]\n", plat[0], plat[1])
	}

	if !all_good {
		os.Exit(1)
	}
}

// EOF
