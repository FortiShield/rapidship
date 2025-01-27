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

package space

import (
	"github.com/nxenv/rapidship/app/api/controller/limiter"
	"github.com/nxenv/rapidship/app/api/controller/repo"
	"github.com/nxenv/rapidship/app/auth/authz"
	"github.com/nxenv/rapidship/app/services/exporter"
	"github.com/nxenv/rapidship/app/services/gitspace"
	"github.com/nxenv/rapidship/app/services/importer"
	"github.com/nxenv/rapidship/app/services/instrument"
	"github.com/nxenv/rapidship/app/services/label"
	"github.com/nxenv/rapidship/app/services/publicaccess"
	"github.com/nxenv/rapidship/app/services/pullreq"
	"github.com/nxenv/rapidship/app/services/refcache"
	"github.com/nxenv/rapidship/app/services/rules"
	"github.com/nxenv/rapidship/app/sse"
	"github.com/nxenv/rapidship/app/store"
	"github.com/nxenv/rapidship/app/url"
	"github.com/nxenv/rapidship/audit"
	"github.com/nxenv/rapidship/store/database/dbtx"
	"github.com/nxenv/rapidship/types"
	"github.com/nxenv/rapidship/types/check"

	"github.com/google/wire"
)

// WireSet provides a wire set for this package.
var WireSet = wire.NewSet(
	ProvideController,
)

func ProvideController(config *types.Config, tx dbtx.Transactor, urlProvider url.Provider, sseStreamer sse.Streamer,
	identifierCheck check.SpaceIdentifier, authorizer authz.Authorizer, spacePathStore store.SpacePathStore,
	pipelineStore store.PipelineStore, secretStore store.SecretStore,
	connectorStore store.ConnectorStore, templateStore store.TemplateStore,
	spaceStore store.SpaceStore, repoStore store.RepoStore, principalStore store.PrincipalStore,
	repoCtrl *repo.Controller, membershipStore store.MembershipStore, prListService *pullreq.ListService,
	spaceCache refcache.SpaceCache,
	importer *importer.Repository, exporter *exporter.Repository,
	limiter limiter.ResourceLimiter, publicAccess publicaccess.Service,
	auditService audit.Service, gitspaceService *gitspace.Service,
	labelSvc *label.Service, instrumentation instrument.Service, executionStore store.ExecutionStore,
	rulesSvc *rules.Service, usageMetricStore store.UsageMetricStore,
) *Controller {
	return NewController(config, tx, urlProvider,
		sseStreamer, identifierCheck, authorizer,
		spacePathStore, pipelineStore, secretStore,
		connectorStore, templateStore,
		spaceStore, repoStore, principalStore,
		repoCtrl, membershipStore, prListService,
		spaceCache,
		importer, exporter, limiter, publicAccess,
		auditService, gitspaceService,
		labelSvc, instrumentation, executionStore,
		rulesSvc, usageMetricStore,
	)
}
