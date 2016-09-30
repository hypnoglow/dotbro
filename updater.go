package main

import (
	"path"

	"github.com/libgit2/git2go"
)

func checkDotfilesUpdates(repoPath string) error {
	repo, err := git.OpenRepository(repoPath)
	if err != nil {
		return err
	}

	headRef, err := repo.Head()
	if err != nil {
		return err
	}

	if path.Base(headRef.Name()) != "master" {
		return nil
	}

	remote, err := repo.Remotes.Lookup("origin")
	if err != nil {
		return err
	}

	rcb := &git.RemoteCallbacks{}
	var headers []string

	if err = remote.Connect(
		git.ConnectDirectionFetch,
		rcb,
		headers,
	); err != nil {
		return err
	}

	remoteHeads, err := remote.Ls("HEAD")
	if err != nil {
		return err
	}

	remoteHead := remoteHeads[0]

	if headRef.Target() == remoteHead.Id {
		return nil
	}

	// TODO: Update!

	return nil
}
