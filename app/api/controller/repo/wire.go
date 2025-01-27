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

package repo

import (
	"github.com/nxenv/rapidship/app/api/controller/limiter"
	"github.com/nxenv/rapidship/app/auth/authz"
	repoevents "github.com/nxenv/rapidship/app/events/repo"
	"github.com/nxenv/rapidship/app/services/codeowners"
	"github.com/nxenv/rapidship/app/services/importer"
	"github.com/nxenv/rapidship/app/services/instrument"
	"github.com/nxenv/rapidship/app/services/keywordsearch"
	"github.com/nxenv/rapidship/app/services/label"
	"github.com/nxenv/rapidship/app/services/locker"
	"github.com/nxenv/rapidship/app/services/protection"
	"github.com/nxenv/rapidship/app/services/publicaccess"
	"github.com/nxenv/rapidship/app/services/refcache"
	"github.com/nxenv/rapidship/app/services/rules"
	"github.com/nxenv/rapidship/app/services/settings"
	"github.com/nxenv/rapidship/app/services/usergroup"
	"github.com/nxenv/rapidship/app/sse"
	"github.com/nxenv/rapidship/app/store"
	"github.com/nxenv/rapidship/app/url"
	"github.com/nxenv/rapidship/audit"
	"github.com/nxenv/rapidship/git"
	"github.com/nxenv/rapidship/lock"
	"github.com/nxenv/rapidship/store/database/dbtx"
	"github.com/nxenv/rapidship/types"
	"github.com/nxenv/rapidship/types/check"

	"github.com/google/wire"
)

// WireSet provides a wire set for this package.
var WireSet = wire.NewSet(
	ProvideController,
)

func ProvideController(
	config *types.Config,
	tx dbtx.Transactor,
	urlProvider url.Provider,
	authorizer authz.Authorizer,
	repoStore store.RepoStore,
	spaceStore store.SpaceStore,
	pipelineStore store.PipelineStore,
	principalStore store.PrincipalStore,
	executionStore store.ExecutionStore,
	ruleStore store.RuleStore,
	checkStore store.CheckStore,
	pullReqStore store.PullReqStore,
	settings *settings.Service,
	principalInfoCache store.PrincipalInfoCache,
	protectionManager *protection.Manager,
	rpcClient git.Interface,
	spaceCache refcache.SpaceCache,
	repoFinder refcache.RepoFinder,
	importer *importer.Repository,
	codeOwners *codeowners.Service,
	repoReporter *repoevents.Reporter,
	indexer keywordsearch.Indexer,
	limiter limiter.ResourceLimiter,
	locker *locker.Locker,
	auditService audit.Service,
	mtxManager lock.MutexManager,
	identifierCheck check.RepoIdentifier,
	repoChecks Check,
	publicAccess publicaccess.Service,
	labelSvc *label.Service,
	instrumentation instrument.Service,
	userGroupStore store.UserGroupStore,
	userGroupService usergroup.SearchService,
	rulesSvc *rules.Service,
	sseStreamer sse.Streamer,
) *Controller {
	return NewController(config, tx, urlProvider,
		authorizer,
		repoStore, spaceStore, pipelineStore, executionStore,
		principalStore, ruleStore, checkStore, pullReqStore, settings,
		principalInfoCache, protectionManager, rpcClient, spaceCache, repoFinder, importer,
		codeOwners, repoReporter, indexer, limiter, locker, auditService, mtxManager, identifierCheck,
		repoChecks, publicAccess, labelSvc, instrumentation, userGroupStore, userGroupService,
		rulesSvc, sseStreamer,
	)
}

func ProvideRepoCheck() Check {
	return NewNoOpRepoChecks()
}
