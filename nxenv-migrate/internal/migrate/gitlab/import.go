// Copyright 2023 Nxenv, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gitlab

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/nxenv/rapidship/nxenv-migrate/internal/nxenv"
	"github.com/nxenv/rapidship/nxenv-migrate/internal/slug"
	"github.com/nxenv/rapidship/nxenv-migrate/internal/tracer"
	"github.com/nxenv/rapidship/nxenv-migrate/internal/types"
	"github.com/nxenv/rapidship/nxenv-migrate/internal/util"

	git "github.com/go-git/go-git/v5"

	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/gotidy/ptr"
)

// Importer imports data from gitlab to Nxenv.
type Importer struct {
	Nxenv      nxenv.Client
	NxenvOrg   string
	NxenvToken string

	ScmType  string // github, gitlab, bitbucket
	ScmLogin string
	ScmToken string

	Tracer tracer.Tracer
}

func (m *Importer) Import(ctx context.Context, data *types.Org) error {

	m.Tracer.Start("create organization %s", m.NxenvOrg)

	// find the nxenv organization
	org, err := m.Nxenv.FindOrg(m.NxenvOrg)
	if err != nil {
		org = &nxenv.Org{
			ID:   m.NxenvOrg,
			Name: m.NxenvOrg,
		}
		// create the organization if not exists
		if err := m.Nxenv.CreateOrg(org); err != nil {
			return err
		}
	}

	// wait for the nxenv secret manager to be created for the
	// organization. It is created async and if we do not wait, it
	// could result in failure to add secrets in subsequent steps.
	if err := nxenv.WaitNxenvSecretManagerOrg(m.Nxenv, m.NxenvOrg); err != nil {
		return err
	}

	m.Tracer.Stop("create organization %s [done]", m.NxenvOrg)

	m.Tracer.Start("create secret %s", m.ScmType)

	// find the github, gitlab or bitbucket secret or
	// create if the secret does not already exist.
	if _, err = m.Nxenv.FindSecretOrg(org.ID, m.ScmType); err != nil {
		// create the scm secret as an inline secret using
		// the nxenv secret manager.
		secret := createSecretOrg(org.ID, m.ScmType, m.ScmToken)
		// save the secret to the organization
		if err := m.Nxenv.CreateSecretOrg(secret); err != nil {
			return err
		}
	}

	m.Tracer.Stop("create secret %s [done]", m.ScmType)

	m.Tracer.Start("create connector %s", m.ScmType)

	// find the github, gitlab or bitbucket connector or
	// create if the connector does not already exist.
	if _, err = m.Nxenv.FindConnectorOrg(org.ID, m.ScmType); err != nil {
		conn := createGitlabConnector(org.ID, m.ScmType, m.ScmLogin, "org."+m.ScmType)
		if err := m.Nxenv.CreateConnectorOrg(conn); err != nil {
			return err
		}
	}

	m.Tracer.Stop("create connector %s [done]", m.ScmType)

	// create tmp dir for cloning repos
	tmpDir, err := os.MkdirTemp("", "nxenv-migrate-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir for org: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// convert each gitlab project to a nxenv project.
	for _, srcProject := range data.Projects {

		m.Tracer.Start("create project %s", srcProject.Name)

		// convert the gitlab project to a nxenv project
		// structure and convert the gitlab project name to
		// a nxenv project identifier.
		project := &nxenv.Project{
			Orgidentifier: org.ID,
			Identifier:    slug.Create(srcProject.Name),
			Name:          srcProject.Name,
		}

		// create the nxenv project.
		if err := m.Nxenv.CreateProject(project); err != nil {
			// if the error indicates the project already exists
			// we can continue with the import, else we should return
			// the error and exit the import.
			if isErrConflict(err) == false {
				return err
			}
		}

		// wait for the nxenv secret manager to be created for the
		// project. It is created async and if we do not wait, it
		// could result in failure to add secrets in subsequent steps.
		if err := nxenv.WaitNxenvSecretManager(
			m.Nxenv, m.NxenvOrg, project.Identifier); err != nil {
			return err
		}

		m.Tracer.Stop("create project %s [done]", srcProject.Name)

		m.Tracer.Start("create repository %s", project.Identifier)

		// create the nxenv repository in nxenv
		repoCreate := &nxenv.CreateRepositoryInput{
			Identifier:    project.Identifier,
			DefaultBranch: srcProject.Branch,
			IsPublic:      false, // TODO: Nxenv doesn't have private repos at the moment
		}

		repoRef := util.JoinPaths(project.Orgidentifier, project.Identifier)
		repo, err := m.Nxenv.CreateRepository(repoRef, repoCreate)
		if err != nil {
			// if the error indicates the project already exists, continue with next project.
			// This is a temporary workaround to avoid conflicts while pushing the git repo.
			// TODO: Handle conflicts properly, this only works because repo migration is the last step of a project migration.
			if isErrConflict(err) {
				m.Tracer.Stop("create repository %s [done]", project.Identifier)
				continue
			}
		}

		m.Tracer.Stop("create repository %s [done]", project.Identifier)

		m.Tracer.Start("clone git repository %s", project.Identifier)

		// create tmp dir for repo clone (use generated name to avoid issues with invalid chars)
		// NOTE: no extra clean-up required - tmpDir is already being cleaned-up.
		tmpRepoDir, err := os.MkdirTemp(tmpDir, "repo-*.git")
		if err != nil {
			return fmt.Errorf("failed to create tempo dir for repo: %w", err)
		}
		gitRepo, err := git.PlainCloneContext(ctx, tmpRepoDir, true, &git.CloneOptions{
			URL: srcProject.Repo,
			Auth: &http.BasicAuth{
				Username: m.ScmLogin,
				Password: m.ScmToken,
			},
			ReferenceName: plumbing.NewBranchReferenceName(srcProject.Branch),
			SingleBranch:  false,
			Tags:          git.AllTags,
			NoCheckout:    true,
		})
		if err != nil {
			return fmt.Errorf("failed to clone repo from '%s': %w", srcProject.Repo, err)
		}

		m.Tracer.Stop("clone git repository %s [done]", project.Identifier)

		m.Tracer.Start("push git repository %s", project.Identifier)

		// add empty nxenv repo as remote
		const gitRemoteNxenv = "nxenv"
		gitRepo.CreateRemote(&config.RemoteConfig{
			Name: gitRemoteNxenv,
			URLs: []string{repo.GitURL},
		})

		// push repo
		err = gitRepo.PushContext(ctx, &git.PushOptions{
			RemoteName: gitRemoteNxenv,
			Auth: &http.BasicAuth{
				Username: "git",
				Password: m.NxenvToken,
			},
			RefSpecs: []config.RefSpec{
				config.RefSpec("refs/remotes/origin/*:refs/heads/*"),
				config.RefSpec("refs/tags/*:refs/tags/*")},
		})
		if err != nil {
			return fmt.Errorf("failed to push repo to '%s': %w", repo.GitURL, err)
		}

		m.Tracer.Stop("push git repository %s [done]", project.Identifier)
	}

	return nil
}

//
// helper functions to simplify complex resource creation.
//

// helper function to create a secret.
func createSecret(org, project, identifier, desc, data string) *nxenv.Secret {
	return &nxenv.Secret{
		Name:              identifier,
		Identifier:        identifier,
		Orgidentifier:     org,
		Projectidentifier: project,
		Description:       desc,
		Type:              "SecretText",
		Spec: &nxenv.SecretText{
			Value:   ptr.String(data),
			Type:    "Inline",
			Manager: "nxenvSecretManager",
		},
	}
}

// helper function to create an org secret.
func createSecretOrg(org, identifier, data string) *nxenv.Secret {
	return createSecret(org, "", identifier, "", data)
}

// helper function to create a github connector
func createGitlabConnector(org, id, username, token string) *nxenv.Connector {
	return &nxenv.Connector{
		Name:          id,
		Identifier:    id,
		Orgidentifier: org,
		Type:          "Gitlab",
		Spec: &nxenv.ConnectorGitlab{
			Type: "Account",
			URL:  "https://gitlab.com",
			Authentication: &nxenv.Resource{
				Type: "Http",
				Spec: &nxenv.Resource{
					Type: "UsernameToken",
					Spec: &nxenv.ConnectorToken{
						Username: username,
						Tokenref: token,
					},
				},
			},
			Apiaccess: &nxenv.Resource{
				Type: "Token",
				Spec: &nxenv.ConnectorToken{
					Tokenref: token,
				},
			},
		},
	}
}

// helper function return true if the error message
// indicate the resource already exists.
func isErrConflict(err error) bool {
	return strings.Contains(err.Error(), "already present") ||
		strings.Contains(err.Error(), "already exists")
}
