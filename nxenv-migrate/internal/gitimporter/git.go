package gitimporter

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/nxenv/rapidship/nxenv-migrate/internal/common"
	"github.com/nxenv/rapidship/nxenv-migrate/internal/nxenv"
	"github.com/nxenv/rapidship/nxenv-migrate/internal/tracer"
	"github.com/nxenv/rapidship/nxenv-migrate/types"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

func (m *Importer) Push(
	ctx context.Context,
	repoPath string,
	repo *nxenv.Repository,
	tracer tracer.Tracer,
) error {
	tracer.Start(common.MsgStartImportGit, repo.GitURL)
	gitPath := filepath.Join(repoPath, types.GitDir)

	gitRepo, err := git.PlainOpen(gitPath)
	if err != nil {
		tracer.Stop("failed to open git dir from %q", repoPath)
		return fmt.Errorf("failed to open the exported repository from %q: %w", repoPath, err)
	}

	const gitRemoteNxenv = "nxenvCode"
	_, err = gitRepo.CreateRemote(&config.RemoteConfig{
		Name: gitRemoteNxenv,
		URLs: []string{repo.GitURL},
	})
	if err != nil && !errors.Is(err, git.ErrRemoteExists) {
		return fmt.Errorf("failed to set remote to %q: %w", repo.GitURL, err)
	}

	var output bytes.Buffer

	err = gitRepo.PushContext(ctx, &git.PushOptions{
		RemoteName: gitRemoteNxenv,
		Auth: &http.BasicAuth{
			Username: "git-importer",
			Password: m.NxenvToken,
		},
		RefSpecs: []config.RefSpec{
			"refs/heads/*:refs/heads/*",
			"refs/tags/*:refs/tags/*",
			"refs/pullreq/*/head:refs/pullreq/*/head",
		},
		RemoteURL:       repo.GitURL,
		Force:           true,
		InsecureSkipTLS: true,
		Progress:        &output,
	})
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		tracer.Stop(common.ErrGitPush, repo.GitURL, err, output.String())
		return fmt.Errorf(common.ErrGitPush, repo.GitURL, err, output.String())
	}

	tracer.Stop(common.MsgCompleteImportGit, repo.GitURL)
	return nil
}
