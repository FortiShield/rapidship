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

package githook

import (
	"github.com/nxenv/rapidship/app/api/controller/limiter"
	"github.com/nxenv/rapidship/app/auth/authz"
	eventsgit "github.com/nxenv/rapidship/app/events/git"
	eventsrepo "github.com/nxenv/rapidship/app/events/repo"
	"github.com/nxenv/rapidship/app/services/protection"
	"github.com/nxenv/rapidship/app/services/settings"
	"github.com/nxenv/rapidship/app/sse"
	"github.com/nxenv/rapidship/app/store"
	"github.com/nxenv/rapidship/app/url"
	"github.com/nxenv/rapidship/git"
	"github.com/nxenv/rapidship/git/hook"

	"github.com/google/wire"
)

var WireSet = wire.NewSet(
	ProvideController,
	ProvideFactory,
)

func ProvideFactory() hook.ClientFactory {
	return &ControllerClientFactory{
		githookCtrl: nil,
	}
}

func ProvideController(
	authorizer authz.Authorizer,
	principalStore store.PrincipalStore,
	repoStore store.RepoStore,
	gitReporter *eventsgit.Reporter,
	repoReporter *eventsrepo.Reporter,
	git git.Interface,
	pullreqStore store.PullReqStore,
	urlProvider url.Provider,
	protectionManager *protection.Manager,
	githookFactory hook.ClientFactory,
	limiter limiter.ResourceLimiter,
	settings *settings.Service,
	preReceiveExtender PreReceiveExtender,
	updateExtender UpdateExtender,
	postReceiveExtender PostReceiveExtender,
	sseStreamer sse.Streamer,
) *Controller {
	ctrl := NewController(
		authorizer,
		principalStore,
		repoStore,
		gitReporter,
		repoReporter,
		git,
		pullreqStore,
		urlProvider,
		protectionManager,
		limiter,
		settings,
		preReceiveExtender,
		updateExtender,
		postReceiveExtender,
		sseStreamer,
	)

	// TODO: improve wiring if possible
	if fct, ok := githookFactory.(*ControllerClientFactory); ok {
		fct.githookCtrl = ctrl
		fct.git = git
	}

	return ctrl
}

var ExtenderWireSet = wire.NewSet(
	ProvidePreReceiveExtender,
	ProvideUpdateExtender,
	ProvidePostReceiveExtender,
)

func ProvidePreReceiveExtender() (PreReceiveExtender, error) {
	return NewPreReceiveExtender(), nil
}

func ProvideUpdateExtender() (UpdateExtender, error) {
	return NewUpdateExtender(), nil
}

func ProvidePostReceiveExtender() (PostReceiveExtender, error) {
	return NewPostReceiveExtender(), nil
}
