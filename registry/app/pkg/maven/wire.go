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

package maven

import (
	"github.com/nxenv/rapidship/app/auth/authz"
	corestore "github.com/nxenv/rapidship/app/store"
	"github.com/nxenv/rapidship/registry/app/pkg/filemanager"
	"github.com/nxenv/rapidship/registry/app/remote/controller/proxy/maven"
	"github.com/nxenv/rapidship/registry/app/store"
	"github.com/nxenv/rapidship/secret"
	"github.com/nxenv/rapidship/store/database/dbtx"

	"github.com/google/wire"
)

func LocalRegistryProvider(
	dBStore *DBStore,
	tx dbtx.Transactor,
	fileManager filemanager.FileManager,
) *LocalRegistry {
	return NewLocalRegistry(dBStore,
		tx,
		fileManager,
	).(*LocalRegistry)
}

func RemoteRegistryProvider(
	dBStore *DBStore,
	tx dbtx.Transactor,
	local *LocalRegistry,
	proxyController maven.Controller,
) *RemoteRegistry {
	return NewRemoteRegistry(dBStore, tx, local, proxyController).(*RemoteRegistry)
}

func ControllerProvider(
	local *LocalRegistry,
	remote *RemoteRegistry,
	authorizer authz.Authorizer,
	dBStore *DBStore,
) *Controller {
	return NewController(local, remote, authorizer, dBStore)
}

func DBStoreProvider(
	registryDao store.RegistryRepository,
	imageDao store.ImageRepository,
	artifactDao store.ArtifactRepository,
	spaceStore corestore.SpaceStore,
	bandwidthStatDao store.BandwidthStatRepository,
	downloadStatDao store.DownloadStatRepository,
	nodeDao store.NodesRepository,
	upstreamProxyDao store.UpstreamProxyConfigRepository,
) *DBStore {
	return NewDBStore(registryDao, imageDao, artifactDao, spaceStore, bandwidthStatDao,
		downloadStatDao,
		nodeDao,
		upstreamProxyDao)
}

func ProvideProxyController(
	registry *LocalRegistry, secretService secret.Service,
	spacePathStore corestore.SpacePathStore,
) maven.Controller {
	return maven.NewProxyController(registry, secretService, spacePathStore)
}

var ControllerSet = wire.NewSet(ControllerProvider)
var DBStoreSet = wire.NewSet(DBStoreProvider)
var RegistrySet = wire.NewSet(LocalRegistryProvider, RemoteRegistryProvider)
var ProxySet = wire.NewSet(ProvideProxyController)
var WireSet = wire.NewSet(ControllerSet, DBStoreSet, RegistrySet, ProxySet)
