//  Copyright 2023 Nxenv, Inc.
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

package generic

import (
	"github.com/nxenv/rapidship/app/auth/authz"
	rapidshipstore "github.com/nxenv/rapidship/app/store"
	"github.com/nxenv/rapidship/registry/app/pkg/filemanager"
	"github.com/nxenv/rapidship/registry/app/store"
	"github.com/nxenv/rapidship/store/database/dbtx"

	"github.com/google/wire"
)

func DBStoreProvider(
	imageDao store.ImageRepository,
	artifactDao store.ArtifactRepository,
	bandwidthStatDao store.BandwidthStatRepository,
	downloadStatDao store.DownloadStatRepository,
	registryDao store.RegistryRepository,
) *DBStore {
	return NewDBStore(registryDao, imageDao, artifactDao, bandwidthStatDao, downloadStatDao)
}

func ControllerProvider(
	spaceStore rapidshipstore.SpaceStore,
	authorizer authz.Authorizer,
	fileManager filemanager.FileManager,
	dBStore *DBStore,
	tx dbtx.Transactor,
) *Controller {
	return NewController(spaceStore, authorizer, fileManager, dBStore, tx)
}

var DBStoreSet = wire.NewSet(DBStoreProvider)
var ControllerSet = wire.NewSet(ControllerProvider)

var WireSet = wire.NewSet(ControllerSet, DBStoreSet)
