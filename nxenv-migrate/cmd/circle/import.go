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

package circle

import (
	"context"
	"encoding/json"
	"os"

	"github.com/nxenv/rapidship/nxenv-migrate/cmd/util"

	"golang.org/x/exp/slog"

	"github.com/alecthomas/kingpin/v2"

	"github.com/nxenv/rapidship/nxenv-migrate/internal/tracer"
	"github.com/nxenv/rapidship/nxenv-migrate/internal/types"
)

type importCommand struct {
	debug bool
	file  string

	nxenvToken   string
	nxenvAccount string
	nxenvOrg     string
	nxenvAddress string

	githubToken    string
	githubURL      string
	gitlabToken    string
	gitlabURL      string
	bitbucketToken string
	bitbucketURL   string
	skipVerify     bool
}

func (c *importCommand) run(*kingpin.ParseContext) error {

	// create the logger
	log := util.CreateLogger(c.debug)

	// attach the logger to the context
	ctx := context.Background()
	ctx = slog.NewContext(ctx, log)

	// read the data file
	data, err := os.ReadFile(c.file)
	if err != nil {
		log.Error("cannot read data file", nil)
		return err
	}

	// unmarshal the data file
	org := new(types.Org)
	if err := json.Unmarshal(data, org); err != nil {
		log.Error("cannot unmarshal data file", nil)
		return err
	}

	// create the tracer
	tracer_ := tracer.New()
	defer tracer_.Close()

	// create the importer
	importer := util.CreateImporter(
		c.nxenvAccount,
		c.nxenvOrg,
		c.nxenvToken,
		c.githubToken,
		c.gitlabToken,
		c.bitbucketToken,
		c.nxenvAddress,
	)
	importer.Tracer = tracer_

	// create a scm client to verify the token
	// and retrieve the user id.
	client := util.CreateClient(
		c.githubToken,
		c.gitlabToken,
		c.bitbucketToken,
		c.githubURL,
		c.gitlabURL,
		c.bitbucketURL,
		c.skipVerify,
	)

	// get the current user id.
	user, _, err := client.Users.Find(ctx)
	if err != nil {
		log.Error("cannot retrieve git user", nil)
		return err
	}

	// provide the user id to the importer. the user id
	// is required by the connector despite the fact that
	// it can be retrieved using the token itself (like we just did)
	importer.ScmLogin = user.Login

	log.Debug("verified user and token",
		slog.String("user", user.Login),
	)

	// execute the import routine.
	return importer.Import(ctx, org)
}

// helper function registers the import command.
func registerImport(app *kingpin.CmdClause) {
	c := new(importCommand)

	cmd := app.Command("import", "import circle data").
		Hidden().
		Action(c.run)

	cmd.Arg("file", "data file to import").
		StringVar(&c.file)

	cmd.Flag("nxenv-account", "nxenv account").
		Required().
		Envar("NXENV_ACCOUNT").
		StringVar(&c.nxenvAccount)

	cmd.Flag("nxenv-org", "nxenv organization").
		Required().
		Envar("NXENV_ORG").
		StringVar(&c.nxenvOrg)

	cmd.Flag("nxenv-token", "nxenv token").
		Required().
		Envar("NXENV_TOKEN").
		StringVar(&c.nxenvToken)

	cmd.Flag("nxenv-address", "nxenv address").
		Envar("NXENV_ADDRESS").
		Default("https://app.nxenv.io").
		StringVar(&c.nxenvAddress)

	cmd.Flag("github-token", "github token").
		Envar("GITHUB_TOKEN").
		StringVar(&c.githubToken)

	cmd.Flag("github-url", "github url").
		Envar("GITHUB_URL").
		StringVar(&c.githubURL)

	cmd.Flag("gitlab-token", "gitlab token").
		Envar("GITLAB_TOKEN").
		StringVar(&c.gitlabToken)

	cmd.Flag("gitlab-url", "gitlab url").
		Envar("GITLAB_URL").
		StringVar(&c.gitlabURL)

	cmd.Flag("bitbucket-token", "bitbucket token").
		Envar("BITBUCKET_TOKEN").
		StringVar(&c.bitbucketToken)

	cmd.Flag("bitbucket-url", "bitbucket url").
		Envar("BITBUCKET_URL").
		StringVar(&c.bitbucketURL)

	cmd.Flag("skip-tls-verify", "skip TLS verification for SCM").
		Envar("SKIP_TLS_VERIFY").
		BoolVar(&c.skipVerify)

	cmd.Flag("debug", "enable debug logging").
		BoolVar(&c.debug)
}
