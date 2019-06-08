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
		if !utils.Validate(c.String("number")) {
			log.Error("invalid number flag")
			os.Exit(1)
		}

		/* Create GitCommit Info */
		// gci := model.SetGitExecuter(model.GitCommander{})

		// gci := make([]model.GitCommitInfo, endNumber-beginNumber+1)
		// for i := range gci {
		// 	gci[i].RangeArray = ""
		// 	gci[i].CommitMsgList = ""
		// 	gci[i].CommitHashList = ""
		// 	gci[i].CommitNewMsgList = ""
		// 	gci[i].CommitSpecifiedMsgList = ""
		// 	gci[i].ReflogHashList = ""
		// 	gci[i].CommitMessageRef = nil
		// 	gci[i].GitExecuter = model.SetGitExecuter(model.GitCommander{})
		// }

		gitCommand := model.SetGitExecuter(model.GitCommander{})
		gitCommand.StartHeadIndex, gitCommand.EndHeadIndex = utils.PickupSquashRange(c.String("number"))
		utils.LogrusInit(c.Bool("debug"))
		gitCommand.GitInfo = make([]model.CommitInfo, gitCommand.StartHeadIndex+1)

		specifiedMsg := c.String("message")

		model.CheckCurrentCommit(c.Bool("force"), gitCommand.StartHeadIndex, gitCommand.EndHeadIndex, gitCommand, specifiedMsg)

		/* Start: Signal Handling Process */

		wg, doneCh, sigCh := model.StartSignalHandling(gitCommand)

		/* Suppress vim editor launching */
		os.Setenv("GIT_EDITOR", ":")

		for i := gitCommand.StartHeadIndex; i >= 0; i-- {
			specifiedHead := fmt.Sprintf("HEAD~%d", i+1)
			var specifiedExec string
			// fmt.Printf("StartIndex:%d EndIndex:%d NowIndex:%d Specified:%s\n", gitCommand.StartHeadIndex, gitCommand.EndHeadIndex, i, specifiedMsg)
			if utils.NeedsChangeMessage(i, gitCommand.StartHeadIndex, gitCommand.EndHeadIndex) { // リベース対象のコミット
				if specifiedMsg == "" { // fixup パターン
					if gitCommand.MakeMesasgeOption(i, gitCommand.StartHeadIndex, "") == true {
						specifiedExec = fmt.Sprintf("--exec=git commit --amend -F .qstemp")
					} else {
						log.Error("Error")
					}
				} else { // squash パターン
					if gitCommand.MakeMesasgeOption(i, gitCommand.StartHeadIndex, specifiedMsg) == true {
						specifiedExec = fmt.Sprintf("--exec=git commit --amend -F .qstemp")
					} else {
						log.Error("Error")
					}
				}
			} else {
				// リベース対象外のコミットなので、Subjectをそのまま
				specifiedExec = fmt.Sprintf("--exec=git commit --amend -m\"%s\"", gitCommand.GitInfo[i].CommitMessageRef.Subject)
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
				model.DoRecovery(doneCh, gitCommand)
			}
		}

		/* git rebase with autosquash option */
		specifiedHead := fmt.Sprintf("HEAD~%d", gitCommand.StartHeadIndex+1)
		cmd := exec.Command("git", "rebase", "-i", "--autosquash", "--autostash", specifiedHead, "--quiet", "--preserve-merges")

		/* Transfer the command I/O to Standard I/O */
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			model.DoRecovery(doneCh, gitCommand)
		}
		log.Debug("rebase completed")

		/* STOP: Signal Handling Process */
		model.StopSignalHandling(wg, doneCh, sigCh)

		return nil
	}

	cmd := exec.Command("rm", ".qstemp")
	if err := cmd.Run(); err != nil {
		fmt.Println(err)
	}

	app.Run(os.Args)
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}
