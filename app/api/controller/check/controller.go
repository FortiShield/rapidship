// Copyright 2023 Nxenv, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package check

import (
	"context"
	"fmt"

	apiauth "github.com/nxenv/rapidship/app/api/auth"
	"github.com/nxenv/rapidship/app/api/controller/space"
	"github.com/nxenv/rapidship/app/api/usererror"
	"github.com/nxenv/rapidship/app/auth"
	"github.com/nxenv/rapidship/app/auth/authz"
	"github.com/nxenv/rapidship/app/services/refcache"
	"github.com/nxenv/rapidship/app/sse"
	"github.com/nxenv/rapidship/app/store"
	"github.com/nxenv/rapidship/git"
	"github.com/nxenv/rapidship/store/database/dbtx"
	"github.com/nxenv/rapidship/types"
	"github.com/nxenv/rapidship/types/enum"
)

type Controller struct {
	tx          dbtx.Transactor
	authorizer  authz.Authorizer
	spaceStore  store.SpaceStore
	checkStore  store.CheckStore
	spaceCache  refcache.SpaceCache
	repoFinder  refcache.RepoFinder
	git         git.Interface
	sanitizers  map[enum.CheckPayloadKind]func(in *ReportInput, s *auth.Session) error
	sseStreamer sse.Streamer
}

func NewController(
	tx dbtx.Transactor,
	authorizer authz.Authorizer,
	spaceStore store.SpaceStore,
	checkStore store.CheckStore,
	spaceCache refcache.SpaceCache,
	repoFinder refcache.RepoFinder,
	git git.Interface,
	sanitizers map[enum.CheckPayloadKind]func(in *ReportInput, s *auth.Session) error,
	sseStreamer sse.Streamer,
) *Controller {
	return &Controller{
		tx:          tx,
		authorizer:  authorizer,
		spaceStore:  spaceStore,
		checkStore:  checkStore,
		spaceCache:  spaceCache,
		repoFinder:  repoFinder,
		git:         git,
		sanitizers:  sanitizers,
		sseStreamer: sseStreamer,
	}
}

//nolint:unparam
func (c *Controller) getRepoCheckAccess(
	ctx context.Context,
	session *auth.Session,
	repoRef string,
	reqPermission enum.Permission,
	allowedRepoStates ...enum.RepoState,
) (*types.Repository, error) {
	if repoRef == "" {
		return nil, usererror.BadRequest("A valid repository reference must be provided.")
	}

	repo, err := c.repoFinder.FindByRef(ctx, repoRef)
	if err != nil {
		return nil, fmt.Errorf("failed to find repository: %w", err)
	}

	if err := apiauth.CheckRepoState(ctx, session, repo, reqPermission, allowedRepoStates...); err != nil {
		return nil, err
	}

	if err = apiauth.CheckRepo(ctx, c.authorizer, session, repo, reqPermission); err != nil {
		return nil, fmt.Errorf("access check failed: %w", err)
	}

	return repo, nil
}

func (c *Controller) getSpaceCheckAccess(
	ctx context.Context,
	session *auth.Session,
	spaceRef string,
	permission enum.Permission,
) (*types.Space, error) {
	return space.GetSpaceCheckAuth(ctx, c.spaceCache, c.authorizer, session, spaceRef, permission)
}
