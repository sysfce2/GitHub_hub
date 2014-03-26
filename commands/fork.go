package commands

import (
	"fmt"
	"github.com/github/hub/github"
	"github.com/github/hub/utils"
	"os"
	"reflect"
)

var cmdFork = &Command{
	Run:   fork,
	Usage: "fork [--no-remote]",
	Short: "Make a fork of a remote repository on GitHub and add as remote",
	Long: `Forks the original project (referenced by "origin" remote) on GitHub and
adds a new remote for it under your username.
`,
}

var flagForkNoRemote bool

func init() {
	cmdFork.Flag.BoolVar(&flagForkNoRemote, "no-remote", false, "")

	CmdRunner.Use(cmdFork)
}

/*
  $ gh fork
  [ repo forked on GitHub ]
  > git remote add -f YOUR_USER git@github.com:YOUR_USER/CURRENT_REPO.git

  $ gh fork --no-remote
  [ repo forked on GitHub ]
*/
func fork(cmd *Command, args *Args) {
	localRepo := github.LocalRepo()

	project, err := localRepo.MainProject()
	utils.Check(err)

	configs := github.CurrentConfigs()
	host := configs.PromptFor(project.Host)
	forkProject := github.NewProject(host.User, project.Name, project.Host)

	client := github.NewClient(project.Host)
	existingRepo, err := client.Repository(forkProject)
	if err == nil {
		var parentURL *github.URL
		if parent := existingRepo.Parent; parent != nil {
			parentURL, _ = github.ParseURL(parent.HTMLURL)
		}
		if parentURL == nil || !reflect.DeepEqual(parentURL.Project, project) {
			err = fmt.Errorf("Error creating fork: %s already exists on %s",
				forkProject, forkProject.Host)
			utils.Check(err)
		}
	} else {
		if !args.Noop {
			_, err := client.ForkRepository(project)
			utils.Check(err)
		}
	}

	if flagForkNoRemote {
		os.Exit(0)
	} else {
		u := forkProject.GitURL("", "", true)
		args.Replace("git", "remote", "add", "-f", forkProject.Owner, u)
		args.After("echo", fmt.Sprintf("new remote: %s", forkProject.Owner))
	}
}
