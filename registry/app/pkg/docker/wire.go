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

package docker

import (
	"github.com/nxenv/rapidship/app/auth/authz"
	rapidshipstore "github.com/nxenv/rapidship/app/store"
	storagedriver "github.com/nxenv/rapidship/registry/app/driver"
	"github.com/nxenv/rapidship/registry/app/event"
	"github.com/nxenv/rapidship/registry/app/manifest/manifestlist"
	"github.com/nxenv/rapidship/registry/app/manifest/schema2"
	"github.com/nxenv/rapidship/registry/app/pkg"
	proxy2 "github.com/nxenv/rapidship/registry/app/remote/controller/proxy"
	"github.com/nxenv/rapidship/registry/app/storage"
	"github.com/nxenv/rapidship/registry/app/store"
	"github.com/nxenv/rapidship/registry/gc"
	"github.com/nxenv/rapidship/secret"
	"github.com/nxenv/rapidship/store/database/dbtx"
	"github.com/nxenv/rapidship/types"

	"github.com/google/wire"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

func LocalRegistryProvider(
	app *App, ms ManifestService, blobRepo store.BlobRepository,
	registryDao store.RegistryRepository, manifestDao store.ManifestRepository,
	registryBlobDao store.RegistryBlobRepository,
	mtRepository store.MediaTypesRepository,
	tagDao store.TagRepository, imageDao store.ImageRepository, artifactDao store.ArtifactRepository,
	bandwidthStatDao store.BandwidthStatRepository, downloadStatDao store.DownloadStatRepository,
	gcService gc.Service, tx dbtx.Transactor,
) *LocalRegistry {
	return NewLocalRegistry(
		app, ms, manifestDao, registryDao, registryBlobDao, blobRepo,
		mtRepository, tagDao, imageDao, artifactDao, bandwidthStatDao, downloadStatDao, gcService, tx,
	).(*LocalRegistry)
}

func ManifestServiceProvider(
	registryDao store.RegistryRepository,
	manifestDao store.ManifestRepository, blobRepo store.BlobRepository, mtRepository store.MediaTypesRepository,
	manifestRefDao store.ManifestReferenceRepository, tagDao store.TagRepository, imageDao store.ImageRepository,
	artifactDao store.ArtifactRepository, layerDao store.LayerRepository,
	gcService gc.Service, tx dbtx.Transactor, reporter event.Reporter, spacePathStore rapidshipstore.SpacePathStore,
	ociImageIndexMappingDao store.OCIImageIndexMappingRepository,
) ManifestService {
	return NewManifestService(
		registryDao, manifestDao, blobRepo, mtRepository, tagDao, imageDao,
		artifactDao, layerDao, manifestRefDao, tx, gcService, reporter, spacePathStore,
		ociImageIndexMappingDao,
	)
}

func RemoteRegistryProvider(
	local *LocalRegistry, app *App, upstreamProxyConfigRepo store.UpstreamProxyConfigRepository,
	spacePathStore rapidshipstore.SpacePathStore, secretService secret.Service, proxyCtrl proxy2.Controller,
) *RemoteRegistry {
	return NewRemoteRegistry(local, app, upstreamProxyConfigRepo, spacePathStore, secretService,
		proxyCtrl).(*RemoteRegistry)
}

func ControllerProvider(
	local *LocalRegistry,
	remote *RemoteRegistry,
	controller *pkg.CoreController,
	spaceStore rapidshipstore.SpaceStore,
	authorizer authz.Authorizer,
	dBStore *DBStore,
) *Controller {
	return NewController(local, remote, controller, spaceStore, authorizer, dBStore)
}

func DBStoreProvider(
	blobRepo store.BlobRepository,
	imageDao store.ImageRepository,
	artifactDao store.ArtifactRepository,
	bandwidthStatDao store.BandwidthStatRepository,
	downloadStatDao store.DownloadStatRepository,
) *DBStore {
	return NewDBStore(blobRepo, imageDao, artifactDao, bandwidthStatDao, downloadStatDao)
}

func StorageServiceProvider(cfg *types.Config, driver storagedriver.StorageDriver) *storage.Service {
	return GetStorageService(cfg, driver)
}

func ProvideReporter() event.Reporter {
	return &event.Noop{}
}

func ProvideProxyController(
	registry *LocalRegistry, ms ManifestService, secretService secret.Service,
	spacePathStore rapidshipstore.SpacePathStore,
) proxy2.Controller {
	manifestCacheHandler := getManifestCacheHandler(registry, ms)
	return proxy2.NewProxyController(registry, ms, secretService, spacePathStore, manifestCacheHandler)
}

func getManifestCacheHandler(
	registry *LocalRegistry, ms ManifestService,
) map[string]proxy2.ManifestCacheHandler {
	cache := proxy2.GetManifestCache(registry, ms)
	listCache := proxy2.GetManifestListCache(registry)

	return map[string]proxy2.ManifestCacheHandler{
		manifestlist.MediaTypeManifestList: listCache,
		v1.MediaTypeImageIndex:             listCache,
		schema2.MediaTypeManifest:          cache,
		proxy2.DefaultHandler:              cache,
	}
}

var ControllerSet = wire.NewSet(ControllerProvider)
var DBStoreSet = wire.NewSet(DBStoreProvider)
var RegistrySet = wire.NewSet(LocalRegistryProvider, ManifestServiceProvider, RemoteRegistryProvider)
var ProxySet = wire.NewSet(ProvideProxyController)
var StorageServiceSet = wire.NewSet(StorageServiceProvider)
var AppSet = wire.NewSet(NewApp)
var WireSet = wire.NewSet(ControllerSet, DBStoreSet, RegistrySet, StorageServiceSet, AppSet, ProxySet)
