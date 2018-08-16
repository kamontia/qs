# qs
## Description
'qs' is the git support command without the interactive editor.  
You can squash some commits very quickly with the one-liner.

## Usage
Easy to execute.

```bash
$ qs n..m [ -f force | -m commit message or commit index ]

(Example) 
// You can see in git-rebase-to-do.
[4]   pick   ff2ec6a Add file-A 
[3]   pick   bbe19f3 Add file-B
[2]   squash 5544b4e Add file-C      // squash to index number 3
[1]   squash 29d02e7 Add file-D      // squash to index number 3
[0]   pick   76f6a9b Add file-E  

// In this case, you type ...
$ qs 1..3 -f
$ ...(some logs)
$ Success!

$ git log --oneline 
 bd28afa Add file-E
 823bad4 Add file-B
 ff2ec6a Add file-A

Congratulations !
qs command can squash some commits very quickly!
Wao!
```

If conflicts occur, qs can NOT squash automatically.  
You must rebase manually.
## Install

To install, use `go get`:

```bash
$ go get -d github.com/kamontia/git-fixup
```

## Contribution

1. Fork ([https://github.com/kamontia/git-fixup/fork](https://github.com/kamontia/git-fixup/fork))
1. Create a feature branch
1. Commit your changes
1. Rebase your local changes against the master branch
1. Run test suite with the `go test ./...` command and confirm that it passes
1. Run `gofmt -s`
1. Create a new Pull Request

## Author

[Tatsuya Kamohara](https://github.com/kamontia)  
[Takeshi Kondo](https://github.com/chaspy)
