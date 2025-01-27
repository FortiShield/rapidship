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

package gitspace

import (
	gitspaceevents "github.com/nxenv/rapidship/app/events/gitspace"
	"github.com/nxenv/rapidship/app/gitspace/orchestrator"
	"github.com/nxenv/rapidship/app/gitspace/scm"
	"github.com/nxenv/rapidship/app/services/infraprovider"
	"github.com/nxenv/rapidship/app/store"
	"github.com/nxenv/rapidship/store/database/dbtx"
	"github.com/nxenv/rapidship/types"

	"github.com/google/wire"
)

var WireSet = wire.NewSet(
	ProvideGitspace,
)

func ProvideGitspace(
	tx dbtx.Transactor,
	gitspaceStore store.GitspaceConfigStore,
	gitspaceInstanceStore store.GitspaceInstanceStore,
	eventReporter *gitspaceevents.Reporter,
	gitspaceEventStore store.GitspaceEventStore,
	spaceStore store.SpaceStore,
	infraProviderSvc *infraprovider.Service,
	orchestrator orchestrator.Orchestrator,
	scm *scm.SCM,
	config *types.Config,
) *Service {
	return NewService(tx, gitspaceStore, gitspaceInstanceStore, eventReporter,
		gitspaceEventStore, spaceStore, infraProviderSvc, orchestrator, scm, config)
}
