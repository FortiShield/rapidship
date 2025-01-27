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

package api

import (
	usercontroller "github.com/nxenv/rapidship/app/api/controller/user"
	"github.com/nxenv/rapidship/app/auth/authn"
	"github.com/nxenv/rapidship/app/auth/authz"
	corestore "github.com/nxenv/rapidship/app/store"
	urlprovider "github.com/nxenv/rapidship/app/url"
	"github.com/nxenv/rapidship/registry/app/api/handler/generic"
	mavenhandler "github.com/nxenv/rapidship/registry/app/api/handler/maven"
	ocihandler "github.com/nxenv/rapidship/registry/app/api/handler/oci"
	"github.com/nxenv/rapidship/registry/app/api/router"
	storagedriver "github.com/nxenv/rapidship/registry/app/driver"
	"github.com/nxenv/rapidship/registry/app/driver/factory"
	"github.com/nxenv/rapidship/registry/app/driver/filesystem"
	"github.com/nxenv/rapidship/registry/app/driver/s3-aws"
	"github.com/nxenv/rapidship/registry/app/pkg"
	"github.com/nxenv/rapidship/registry/app/pkg/docker"
	"github.com/nxenv/rapidship/registry/app/pkg/filemanager"
	generic2 "github.com/nxenv/rapidship/registry/app/pkg/generic"
	"github.com/nxenv/rapidship/registry/app/pkg/maven"
	"github.com/nxenv/rapidship/registry/app/store/database"
	"github.com/nxenv/rapidship/registry/config"
	"github.com/nxenv/rapidship/registry/gc"
	"github.com/nxenv/rapidship/types"

	"github.com/google/wire"
	"github.com/rs/zerolog/log"
)

type RegistryApp struct {
	Config *types.Config

	AppRouter router.AppRouter
}

func BlobStorageProvider(c *types.Config) (storagedriver.StorageDriver, error) {
	var d storagedriver.StorageDriver
	var err error

	if c.Registry.Storage.StorageType == "filesystem" {
		filesystem.Register()
		d, err = factory.Create("filesystem", config.GetFilesystemParams(c))
		if err != nil {
			log.Fatal().Stack().Err(err).Msgf("")
			panic(err)
		}
	} else {
		s3.Register()
		d, err = factory.Create("s3aws", config.GetS3StorageParameters(c))
		if err != nil {
			log.Error().Stack().Err(err).Msg("failed to init s3 Blob storage ")
			panic(err)
		}
	}
	return d, err
}

func NewHandlerProvider(
	controller *docker.Controller, spaceStore corestore.SpaceStore,
	tokenStore corestore.TokenStore, userCtrl *usercontroller.Controller, authenticator authn.Authenticator,
	urlProvider urlprovider.Provider, authorizer authz.Authorizer, config *types.Config,
) *ocihandler.Handler {
	return ocihandler.NewHandler(
		controller,
		spaceStore,
		tokenStore,
		userCtrl,
		authenticator,
		urlProvider,
		authorizer,
		config.Registry.HTTP.RelativeURL,
	)
}

func NewMavenHandlerProvider(
	controller *maven.Controller, spaceStore corestore.SpaceStore,
	tokenStore corestore.TokenStore, userCtrl *usercontroller.Controller, authenticator authn.Authenticator,
	authorizer authz.Authorizer,
) *mavenhandler.Handler {
	return mavenhandler.NewHandler(
		controller,
		spaceStore,
		tokenStore,
		userCtrl,
		authenticator,
		authorizer,
	)
}

func NewGenericHandlerProvider(
	spaceStore corestore.SpaceStore, controller *generic2.Controller, tokenStore corestore.TokenStore,
	userCtrl *usercontroller.Controller, authenticator authn.Authenticator, urlProvider urlprovider.Provider,
	authorizer authz.Authorizer,
) *generic.Handler {
	return generic.NewGenericArtifactHandler(
		spaceStore,
		controller,
		tokenStore,
		userCtrl,
		authenticator,
		urlProvider,
		authorizer,
	)
}

var WireSet = wire.NewSet(
	BlobStorageProvider,
	NewHandlerProvider,
	NewMavenHandlerProvider,
	NewGenericHandlerProvider,
	database.WireSet,
	pkg.WireSet,
	docker.WireSet,
	filemanager.WireSet,
	maven.WireSet,
	router.WireSet,
	gc.WireSet,
	generic2.WireSet,
)

func Wire(_ *types.Config) (RegistryApp, error) {
	wire.Build(WireSet, wire.Struct(new(RegistryApp), "*"))
	return RegistryApp{}, nil
}
