package main

import (
	"fmt"
	"os"
	"strings"
	"path/filepath"
	"runtime"
)

var g_libpaths []string

func path_exists(name string) bool {
	_, err := os.Stat(name)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

// sysLibName returns the OS-native library name from an OS-independant one
func sysLibName(name string) string {
	prefix := map[string]string{
		"linux": "lib",
		"darwin": "lib",
		"windows": "",
	}[runtime.GOOS]

	suffix := map[string]string{
		"linux": ".so",
		"darwin":".dyld",
		"windows": ".dll",
	}[runtime.GOOS]

	if !strings.HasPrefix(name, prefix) {
		name = prefix + name
	}
	if !strings.HasSuffix(name, suffix) {
		name = name + suffix
	}
	return name
}

func findLib(libname string) string {
	lib := sysLibName(libname)
	for _, path := range g_libpaths {
		fname := filepath.Join(path, lib)
		if path_exists(fname) {
			return fname
		}
	}
	return ""
}

func init() {
	g_libpaths = func() []string {
		v := os.Getenv("LD_LIBRARY_PATH")
		if v == "" {
			return nil
		}
		o := make([]string, 0)
		toks := strings.Split(v, string(os.PathListSeparator))
		for _, tok := range toks {
			if tok != "" {
				o = append(o, tok)
			}
		}
		return o
	}()
}

func main() {

	if len(os.Args) <= 1 {
		fmt.Fprintf(
			os.Stderr, 
			"**error** atl-find-library takes at least one argument\nex:\n%s\n",
			"$ atl-find-library AthenaServices",
		)
		os.Exit(1)
	}


	allGood := true
	libnames := append([]string{}, os.Args[1:]...)
	for _, libname := range libnames {
		libname = strings.Trim(libname, " \t\r\n")
		lib := findLib(libname)
		if lib == "" {
			fmt.Fprintf(
				os.Stderr, 
				"**error** could not locate library [%s]\n", 
				libname,
			)
			allGood = false
			continue
		}
		fmt.Fprintf(os.Stdout, "%s\n", lib)
	}

	if !allGood {
		os.Exit(1)
	}
}
