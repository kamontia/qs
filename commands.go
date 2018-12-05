package main

import (
	"fmt"
	"os"

	"github.com/kamontia/qs/model"
	"github.com/urfave/cli"
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
			gci := new(model.GitCommitInfo)
			validate(c.String("number"))
			specifiedMsg := c.String("message")

			beginNumber, endNumber := pickupSquashRange(c.String("number"))
			logrusInit(c.Bool("debug"))
			getCommitHash(gci)
			getCommitMessage(gci, specifiedMsg)
			displayCommitHashAndMessage(gci, beginNumber, endNumber)
			return nil
		},
	},
}

func CommandNotFound(c *cli.Context, command string) {
	fmt.Fprintf(os.Stderr, "%s: '%s' is not a %s command. See '%s --help'.", c.App.Name, command, c.App.Name, c.App.Name)
	os.Exit(2)
}
