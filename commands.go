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
			gitCommand := model.SetGitExecuter(model.GitCommander{})
			utils.Validate(c.String("number"))
			specifiedMsg := c.String("message")
			gitCommand.StartHeadIndex, gitCommand.EndHeadIndex = utils.PickupSquashRange(c.String("number"))
			gitCommand.GitInfo = make([]model.CommitInfo, gitCommand.StartHeadIndex+2)
			gitCommand.AddCommitHash()
			gitCommand.AddCommitMessage(specifiedMsg)

			gitCommand.DisplayCommitHashAndMessage(gitCommand.EndHeadIndex, gitCommand.StartHeadIndex)
			return nil
		},
	},
}

func CommandNotFound(c *cli.Context, command string) {
	fmt.Fprintf(os.Stderr, "%s: '%s' is not a %s command. See '%s --help'.", c.App.Name, command, c.App.Name, c.App.Name)
	os.Exit(2)
}
