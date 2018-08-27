package main

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	log "github.com/sirupsen/logrus"
)

var GlobalFlags = []cli.Flag{}

var Commands = []cli.Command{
	{
		Name:  "ls",
		Usage: "list commits",
		Flags: []cli.Flag{
			cli.IntFlag{
				Name: "number, n",
			},
		},
		Action: func(c *cli.Context) error {
			log.Infof("ls -n %d", c.Int("number"))
			ls()
			return nil
		},
	},
}

func CommandNotFound(c *cli.Context, command string) {
	fmt.Fprintf(os.Stderr, "%s: '%s' is not a %s command. See '%s --help'.", c.App.Name, command, c.App.Name, c.App.Name)
	os.Exit(2)
}
