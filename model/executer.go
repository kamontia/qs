package model

import "os/exec"

// GitCommander executer exec.Command(git **)
type GitCommander struct{}

// Reflog execute ("git" "reflog" opt)
func (g GitCommander) Reflog(opt string) ([]byte, error) {
	return exec.Command("git", "reflog", opt).Output()
}

// Execute git log opts
func (g GitCommander) Commitlog(opt string) ([]byte, error) {
	return exec.Command("git", "log", opt).Output()
}

// TODO: Implemantation
// --------------------------------------
//
// out, err := exec.Command("git", "log", "--oneline", "--format=%h").Output()
// out, err := exec.Command("git", "log", "--oneline", "--format=%s").Output()
// out, err := exec.Command("git", "log", "--oneline", "-n", sNum).Output()

// cmd = exec.Command("git", "reset", "--hard", gci.ReflogHashList[0])

// cmd := exec.Command("git", "rebase", "--abort")
// cmd := exec.Command("git", "rebase", specifiedHead, specifiedExec)
// cmd := exec.Command("git", "rebase", "-i", "--autosquash", "--autostash", specifiedHead, "--quiet", "--preserve-merges")
