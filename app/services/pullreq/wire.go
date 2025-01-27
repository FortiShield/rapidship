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

package pullreq

import (
	"context"

	"github.com/nxenv/rapidship/app/auth/authz"
	gitevents "github.com/nxenv/rapidship/app/events/git"
	pullreqevents "github.com/nxenv/rapidship/app/events/pullreq"
	"github.com/nxenv/rapidship/app/services/codecomments"
	"github.com/nxenv/rapidship/app/services/label"
	"github.com/nxenv/rapidship/app/services/protection"
	"github.com/nxenv/rapidship/app/sse"
	"github.com/nxenv/rapidship/app/store"
	"github.com/nxenv/rapidship/app/url"
	"github.com/nxenv/rapidship/events"
	"github.com/nxenv/rapidship/git"
	"github.com/nxenv/rapidship/pubsub"
	"github.com/nxenv/rapidship/store/database/dbtx"
	"github.com/nxenv/rapidship/types"

	"github.com/google/wire"
)

var WireSet = wire.NewSet(
	ProvideService,
	ProvideListService,
)

func ProvideService(ctx context.Context,
	config *types.Config,
	gitReaderFactory *events.ReaderFactory[*gitevents.Reader],
	pullReqEvFactory *events.ReaderFactory[*pullreqevents.Reader],
	pullReqEvReporter *pullreqevents.Reporter,
	git git.Interface,
	repoGitInfoCache store.RepoGitInfoCache,
	repoStore store.RepoStore,
	pullreqStore store.PullReqStore,
	activityStore store.PullReqActivityStore,
	principalInfoCache store.PrincipalInfoCache,
	codeCommentView store.CodeCommentView,
	codeCommentMigrator *codecomments.Migrator,
	fileViewStore store.PullReqFileViewStore,
	pubsub pubsub.PubSub,
	urlProvider url.Provider,
	sseStreamer sse.Streamer,
) (*Service, error) {
	return New(ctx,
		config,
		gitReaderFactory,
		pullReqEvFactory,
		pullReqEvReporter,
		git,
		repoGitInfoCache,
		repoStore,
		pullreqStore,
		activityStore,
		codeCommentView,
		codeCommentMigrator,
		fileViewStore,
		principalInfoCache,
		pubsub,
		urlProvider,
		sseStreamer,
	)
}

func ProvideListService(
	tx dbtx.Transactor,
	git git.Interface,
	authorizer authz.Authorizer,
	spaceStore store.SpaceStore,
	repoStore store.RepoStore,
	repoGitInfoCache store.RepoGitInfoCache,
	pullreqStore store.PullReqStore,
	checkStore store.CheckStore,
	labelSvc *label.Service,
	protectionManager *protection.Manager,
) *ListService {
	return NewListService(
		tx,
		git,
		authorizer,
		spaceStore,
		repoStore,
		repoGitInfoCache,
		pullreqStore,
		checkStore,
		labelSvc,
		protectionManager,
	)
}
