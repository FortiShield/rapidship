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

package exporter

import (
	"github.com/nxenv/rapidship/app/sse"
	"github.com/nxenv/rapidship/app/store"
	"github.com/nxenv/rapidship/app/url"
	"github.com/nxenv/rapidship/encrypt"
	"github.com/nxenv/rapidship/git"
	"github.com/nxenv/rapidship/job"

	"github.com/google/wire"
)

var WireSet = wire.NewSet(
	ProvideSpaceExporter,
)

func ProvideSpaceExporter(
	urlProvider url.Provider,
	git git.Interface,
	repoStore store.RepoStore,
	scheduler *job.Scheduler,
	executor *job.Executor,
	encrypter encrypt.Encrypter,
	sseStreamer sse.Streamer,
) (*Repository, error) {
	exporter := &Repository{
		urlProvider: urlProvider,
		git:         git,
		repoStore:   repoStore,
		scheduler:   scheduler,
		encrypter:   encrypter,
		sseStreamer: sseStreamer,
	}

	err := executor.Register(jobType, exporter)
	if err != nil {
		return nil, err
	}

	return exporter, nil
}
