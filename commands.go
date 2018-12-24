package main

import (
	"fmt"
	"os"

	"github.com/kamontia/qs/model"
	"github.com/kamontia/qs/utils"
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
			gci := model.SetGitExecuter(model.GitCommander{})
			utils.Validate(c.String("number"))
			specifiedMsg := c.String("message")

			beginNumber, endNumber := utils.PickupSquashRange(c.String("number"))
			utils.LogrusInit(c.Bool("debug"))
			gci.AddCommitHash()
			gci.AddCommitMessage(specifiedMsg)
			gci.DisplayCommitHashAndMessage(beginNumber, endNumber)
			return nil
		},
	},
}

func CommandNotFound(c *cli.Context, command string) {
	fmt.Fprintf(os.Stderr, "%s: '%s' is not a %s command. See '%s --help'.", c.App.Name, command, c.App.Name, c.App.Name)
	os.Exit(2)
}
