package main

import (
	"fmt"
	"io"
	"strings"
	"text/template"
)

const tmpl = `
## -*- python -*-

## waf imports
import waflib.Logs as msg

PACKAGE = {
    "name":    "LCG_Configuration",
    "authors": ["ATLAS Collaboration"],
}

### ---------------------------------------------------------------------------
def pkg_deps(ctx):
    
    ## public dependencies
    ctx.use_pkg("LCG_Platforms", version="", public=True)
    
    ## no private dependencies
    ## no runtime dependencies
    return # pkg_deps


### ---------------------------------------------------------------------------
def options(ctx):
    
    return # options


### ---------------------------------------------------------------------------
def configure(ctx):
    

    macro = ctx.hwaf_declare_macro
    
    ## projects.
{{. | gen_lcg_config}}
    ctx.msg("LCG", ctx.env["LCG_config_version"])

{{. | gen_lcgapp_config}}

    ## externals
{{. | gen_lcgext_config}}

    return # configure


def build(ctx):
    return # build

## EOF ##
`

const INDENT = "    "

func render(rel *Release, w io.Writer) error {
	t := template.New("hscript")
	t.Funcs(template.FuncMap{
		"gen_lcg_config":    gen_lcg,
		"gen_lcgapp_config": gen_lcgapp,
		"gen_lcgext_config": gen_lcgext,
		//"gen_config": gen_config,
	})
	template.Must(t.Parse(tmpl))
	return t.Execute(w, rel)
}

func gen_lcg(rel *Release) string {
	return fmt.Sprintf(
		`%smacro("%s_config_version", "%s")`,
		INDENT,
		rel.lcg.Name,
		rel.lcg.Version,
	)
}

func gen_lcgapp(rel *Release) string {
	return gen_config(rel.LcgApps())
}

func gen_lcgext(rel *Release) string {
	return gen_config(rel.LcgExternals())
}

func gen_config(pkgs []*Package) string {
	var str []string
	for _, pkg := range pkgs {
		key := fmt.Sprintf(`"%s_config_version",`, pkg.Name)
		str = append(
			str,
			fmt.Sprintf(
				"%smacro(%-40s \"%s\")",
				INDENT,
				key,
				pkg.Version,
			),
		)
	}
	return strings.Join(str, "\n")
}

// EOF
