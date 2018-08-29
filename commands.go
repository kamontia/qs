package main

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
)

var GlobalFlags = []cli.Flag{}

var Commands = []cli.Command{
	{
		Name:  "ls",
		Usage: "list commits",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name: "number, n",
			},
		},
		Action: func(c *cli.Context) error {
			validate(c.String("number"))
			specifiedMsg = c.String("message")

			pick_up_squash_range(c.String("number"))
			logrus_init(c.Bool("debug"))
			get_commit_hash()
			get_commit_message()
			display_commit_hash_and_message()
			return nil
		},
	},
}

func CommandNotFound(c *cli.Context, command string) {
	fmt.Fprintf(os.Stderr, "%s: '%s' is not a %s command. See '%s --help'.", c.App.Name, command, c.App.Name, c.App.Name)
	os.Exit(2)
}
