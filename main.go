package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/kamontia/qs/model"
	"github.com/kamontia/qs/utils"
	"github.com/labstack/gommon/log"

	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = Name
	app.Version = Version
	app.Author = "Tatsuya Kamohara<kamontia@gmail.com>\n   Takeshi Kondo<take.she12@gmail.com>"
	app.Email = ""
	app.Usage = ""

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "number, n",
			Usage: "Specify suqash number",
		},
		cli.BoolFlag{
			Name:  "force, f",
			Usage: "Force update",
		},
		cli.BoolFlag{
			Name:  "debug, d",
			Usage: "Show verbose logging",
		},
		cli.StringFlag{
			Name:  "message, m",
			Usage: "Commit message",
		},
	}
	app.Commands = Commands
	app.CommandNotFound = CommandNotFound

	app.Action = func(c *cli.Context) error {
		/* Create GitCommit Info */
		gci := model.SetGitExecuter(model.GitCommander{})

		if !utils.Validate(c.String("number")) {
			log.Error("invalid number flag")
			os.Exit(1)
		}

		specifiedMsg := c.String("message")

		beginNumber, endNumber := utils.PickupSquashRange(c.String("number"))
		utils.LogrusInit(c.Bool("debug"))
		model.CheckCurrentCommit(c.Bool("force"), beginNumber, endNumber, gci, specifiedMsg)

		/* Start: Signal Handling Process */
		wg, doneCh, sigCh := model.StartSignalHandling(gci)

		/* Suppress vim editor launching */
		os.Setenv("GIT_EDITOR", ":")

		for i := beginNumber; i >= 0; i-- {
			specifiedHead := fmt.Sprintf("HEAD~%d", i+1)
			var specifiedExec string

			if specifiedMsg != "" {
				if beginNumber == i {
					specifiedExec = fmt.Sprintf("--exec=git commit --amend -m\"%s\"", specifiedMsg)
				} else if utils.NeedsChangeMessage(i, beginNumber, endNumber) {
					specifiedExec = fmt.Sprintf("--exec=git commit --amend -m\"%s\"", gci.CommitSpecifiedMsgList[beginNumber])
				} else {
					specifiedExec = fmt.Sprintf("--exec=git commit --amend -m\"%s\"", gci.CommitMsgList[i])
				}
			} else {
				if utils.NeedsChangeMessage(i, beginNumber, endNumber) {
					specifiedExec = fmt.Sprintf("--exec=git commit --amend -m\"%s\"", gci.CommitNewMsgList[beginNumber])
				} else {
					specifiedExec = fmt.Sprintf("--exec=git commit --amend -m\"%s\"", gci.CommitMsgList[i])
				}
			}

			cmd := exec.Command("git", "rebase", specifiedHead, specifiedExec)
			log.Debugf("git rebase HEAD~%d %s\n", i, specifiedExec)

			if c.Bool("debug") {
				cmd.Stdin = os.Stdin
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
			} else {
				cmd.Stdin = nil
				cmd.Stdout = nil
				cmd.Stderr = nil
			}
			if err := cmd.Run(); err != nil {
				model.DoRecovery(doneCh, gci)
			}
		}

		/* git rebase with autosquash option */
		specifiedHead := fmt.Sprintf("HEAD~%d", beginNumber+1)
		cmd := exec.Command("git", "rebase", "-i", "--autosquash", "--autostash", specifiedHead, "--quiet", "--preserve-merges")

		/* Transfer the command I/O to Standard I/O */
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			model.DoRecovery(doneCh, gci)
		}
		log.Debug("rebase completed")

		/* STOP: Signal Handling Process */
		model.StopSignalHandling(wg, doneCh, sigCh)

		return nil
	}
	app.Run(os.Args)
}
