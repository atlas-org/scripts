// atl-svn is a set of commands to interact with atlasoff SVN repository.
package main

import (
	"os"

	"github.com/gonuts/commander"
	"github.com/gonuts/flag"
	"github.com/gonuts/logger"
)

var g_cmd *commander.Commander
var msg = logger.New("avn")

func init() {
	g_cmd = &commander.Commander{
		Name: os.Args[0],
		Commands: []*commander.Command{
			atl_make_cmd_diff(),
		},
		Flag: flag.NewFlagSet("avn", flag.ExitOnError),
	}
}

func handle_err(err error) {
	if err != nil {
		msg.Errorf("%v\n", err.Error())
		os.Exit(1)
	}
}

func main() {

	err := g_cmd.Flag.Parse(os.Args[1:])
	handle_err(err)

	args := g_cmd.Flag.Args()
	err = g_cmd.Run(args)
	handle_err(err)
}

// EOF
