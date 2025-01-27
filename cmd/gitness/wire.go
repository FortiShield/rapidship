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

//go:build wireinject
// +build wireinject

package main

import (
	"context"

	"github.com/nxenv/rapidship/app/api/controller/aiagent"
	"github.com/nxenv/rapidship/app/api/controller/capabilities"
	checkcontroller "github.com/nxenv/rapidship/app/api/controller/check"
	"github.com/nxenv/rapidship/app/api/controller/connector"
	"github.com/nxenv/rapidship/app/api/controller/execution"
	githookCtrl "github.com/nxenv/rapidship/app/api/controller/githook"
	gitspaceCtrl "github.com/nxenv/rapidship/app/api/controller/gitspace"
	infraproviderCtrl "github.com/nxenv/rapidship/app/api/controller/infraprovider"
	controllerkeywordsearch "github.com/nxenv/rapidship/app/api/controller/keywordsearch"
	"github.com/nxenv/rapidship/app/api/controller/limiter"
	controllerlogs "github.com/nxenv/rapidship/app/api/controller/logs"
	"github.com/nxenv/rapidship/app/api/controller/migrate"
	"github.com/nxenv/rapidship/app/api/controller/pipeline"
	"github.com/nxenv/rapidship/app/api/controller/plugin"
	"github.com/nxenv/rapidship/app/api/controller/principal"
	"github.com/nxenv/rapidship/app/api/controller/pullreq"
	"github.com/nxenv/rapidship/app/api/controller/repo"
	"github.com/nxenv/rapidship/app/api/controller/reposettings"
	"github.com/nxenv/rapidship/app/api/controller/secret"
	"github.com/nxenv/rapidship/app/api/controller/service"
	"github.com/nxenv/rapidship/app/api/controller/serviceaccount"
	"github.com/nxenv/rapidship/app/api/controller/space"
	"github.com/nxenv/rapidship/app/api/controller/system"
	"github.com/nxenv/rapidship/app/api/controller/template"
	controllertrigger "github.com/nxenv/rapidship/app/api/controller/trigger"
	"github.com/nxenv/rapidship/app/api/controller/upload"
	"github.com/nxenv/rapidship/app/api/controller/user"
	"github.com/nxenv/rapidship/app/api/controller/usergroup"
	controllerwebhook "github.com/nxenv/rapidship/app/api/controller/webhook"
	"github.com/nxenv/rapidship/app/api/openapi"
	"github.com/nxenv/rapidship/app/auth/authn"
	"github.com/nxenv/rapidship/app/auth/authz"
	"github.com/nxenv/rapidship/app/bootstrap"
	connectorservice "github.com/nxenv/rapidship/app/connector"
	gitevents "github.com/nxenv/rapidship/app/events/git"
	gitspaceevents "github.com/nxenv/rapidship/app/events/gitspace"
	gitspaceinfraevents "github.com/nxenv/rapidship/app/events/gitspaceinfra"
	pipelineevents "github.com/nxenv/rapidship/app/events/pipeline"
	pullreqevents "github.com/nxenv/rapidship/app/events/pullreq"
	repoevents "github.com/nxenv/rapidship/app/events/repo"
	infrastructure "github.com/nxenv/rapidship/app/gitspace/infrastructure"
	"github.com/nxenv/rapidship/app/gitspace/logutil"
	"github.com/nxenv/rapidship/app/gitspace/orchestrator"
	containerorchestrator "github.com/nxenv/rapidship/app/gitspace/orchestrator/container"
	"github.com/nxenv/rapidship/app/gitspace/orchestrator/ide"
	"github.com/nxenv/rapidship/app/gitspace/orchestrator/runarg"
	"github.com/nxenv/rapidship/app/gitspace/platformconnector"
	"github.com/nxenv/rapidship/app/gitspace/scm"
	gitspacesecret "github.com/nxenv/rapidship/app/gitspace/secret"
	"github.com/nxenv/rapidship/app/pipeline/canceler"
	"github.com/nxenv/rapidship/app/pipeline/commit"
	"github.com/nxenv/rapidship/app/pipeline/converter"
	"github.com/nxenv/rapidship/app/pipeline/file"
	"github.com/nxenv/rapidship/app/pipeline/manager"
	"github.com/nxenv/rapidship/app/pipeline/resolver"
	"github.com/nxenv/rapidship/app/pipeline/runner"
	"github.com/nxenv/rapidship/app/pipeline/scheduler"
	"github.com/nxenv/rapidship/app/pipeline/triggerer"
	"github.com/nxenv/rapidship/app/router"
	"github.com/nxenv/rapidship/app/server"
	"github.com/nxenv/rapidship/app/services"
	aiagentservice "github.com/nxenv/rapidship/app/services/aiagent"
	capabilitiesservice "github.com/nxenv/rapidship/app/services/capabilities"
	"github.com/nxenv/rapidship/app/services/cleanup"
	"github.com/nxenv/rapidship/app/services/codecomments"
	"github.com/nxenv/rapidship/app/services/codeowners"
	"github.com/nxenv/rapidship/app/services/exporter"
	"github.com/nxenv/rapidship/app/services/gitspaceevent"
	"github.com/nxenv/rapidship/app/services/gitspaceservice"
	"github.com/nxenv/rapidship/app/services/importer"
	"github.com/nxenv/rapidship/app/services/instrument"
	"github.com/nxenv/rapidship/app/services/keywordsearch"
	svclabel "github.com/nxenv/rapidship/app/services/label"
	locker "github.com/nxenv/rapidship/app/services/locker"
	messagingservice "github.com/nxenv/rapidship/app/services/messaging"
	"github.com/nxenv/rapidship/app/services/metric"
	migrateservice "github.com/nxenv/rapidship/app/services/migrate"
	"github.com/nxenv/rapidship/app/services/notification"
	"github.com/nxenv/rapidship/app/services/notification/mailer"
	"github.com/nxenv/rapidship/app/services/protection"
	"github.com/nxenv/rapidship/app/services/publicaccess"
	"github.com/nxenv/rapidship/app/services/publickey"
	pullreqservice "github.com/nxenv/rapidship/app/services/pullreq"
	"github.com/nxenv/rapidship/app/services/refcache"
	reposervice "github.com/nxenv/rapidship/app/services/repo"
	"github.com/nxenv/rapidship/app/services/rules"
	secretservice "github.com/nxenv/rapidship/app/services/secret"
	"github.com/nxenv/rapidship/app/services/settings"
	systemsvc "github.com/nxenv/rapidship/app/services/system"
	"github.com/nxenv/rapidship/app/services/trigger"
	"github.com/nxenv/rapidship/app/services/usage"
	usergroupservice "github.com/nxenv/rapidship/app/services/usergroup"
	"github.com/nxenv/rapidship/app/services/webhook"
	"github.com/nxenv/rapidship/app/sse"
	"github.com/nxenv/rapidship/app/store"
	"github.com/nxenv/rapidship/app/store/cache"
	"github.com/nxenv/rapidship/app/store/database"
	"github.com/nxenv/rapidship/app/store/logs"
	"github.com/nxenv/rapidship/app/url"
	"github.com/nxenv/rapidship/audit"
	"github.com/nxenv/rapidship/blob"
	cliserver "github.com/nxenv/rapidship/cli/operations/server"
	"github.com/nxenv/rapidship/encrypt"
	"github.com/nxenv/rapidship/events"
	"github.com/nxenv/rapidship/git"
	"github.com/nxenv/rapidship/git/api"
	"github.com/nxenv/rapidship/git/storage"
	infraproviderpkg "github.com/nxenv/rapidship/infraprovider"
	"github.com/nxenv/rapidship/job"
	"github.com/nxenv/rapidship/livelog"
	"github.com/nxenv/rapidship/lock"
	"github.com/nxenv/rapidship/pubsub"
	"github.com/nxenv/rapidship/registry/app/pkg/docker"
	"github.com/nxenv/rapidship/ssh"
	"github.com/nxenv/rapidship/store/database/dbtx"
	"github.com/nxenv/rapidship/types"
	"github.com/nxenv/rapidship/types/check"

	"github.com/google/wire"
)

func initSystem(ctx context.Context, config *types.Config) (*cliserver.System, error) {
	wire.Build(
		cliserver.NewSystem,
		cliserver.ProvideRedis,
		bootstrap.WireSet,
		cliserver.ProvideDatabaseConfig,
		database.WireSet,
		cliserver.ProvideBlobStoreConfig,
		mailer.WireSet,
		notification.WireSet,
		blob.WireSet,
		dbtx.WireSet,
		cache.WireSet,
		refcache.WireSet,
		router.WireSet,
		pullreqservice.WireSet,
		services.WireSet,
		services.ProvideGitspaceServices,
		server.WireSet,
		url.WireSet,
		space.WireSet,
		limiter.WireSet,
		publicaccess.WireSet,
		repo.WireSet,
		reposettings.WireSet,
		pullreq.WireSet,
		controllerwebhook.WireSet,
		controllerwebhook.ProvidePreprocessor,
		svclabel.WireSet,
		serviceaccount.WireSet,
		user.WireSet,
		upload.WireSet,
		service.WireSet,
		principal.WireSet,
		usergroupservice.WireSet,
		system.WireSet,
		authn.WireSet,
		authz.WireSet,
		infrastructure.WireSet,
		infraproviderpkg.WireSet,
		gitspaceevents.WireSet,
		pipelineevents.WireSet,
		infraproviderCtrl.WireSet,
		gitspaceCtrl.WireSet,
		gitevents.WireSet,
		pullreqevents.WireSet,
		repoevents.WireSet,
		storage.WireSet,
		api.WireSet,
		cliserver.ProvideGitConfig,
		git.WireSet,
		store.WireSet,
		check.WireSet,
		encrypt.WireSet,
		cliserver.ProvideEventsConfig,
		events.WireSet,
		cliserver.ProvideWebhookConfig,
		cliserver.ProvideNotificationConfig,
		webhook.WireSet,
		cliserver.ProvideTriggerConfig,
		trigger.WireSet,
		githookCtrl.ExtenderWireSet,
		githookCtrl.WireSet,
		cliserver.ProvideLockConfig,
		lock.WireSet,
		locker.WireSet,
		cliserver.ProvidePubsubConfig,
		pubsub.WireSet,
		cliserver.ProvideJobsConfig,
		job.WireSet,
		cliserver.ProvideCleanupConfig,
		cleanup.WireSet,
		codecomments.WireSet,
		protection.WireSet,
		checkcontroller.WireSet,
		execution.WireSet,
		pipeline.WireSet,
		logs.WireSet,
		livelog.WireSet,
		controllerlogs.WireSet,
		secret.WireSet,
		connector.WireSet,
		connectorservice.WireSet,
		template.WireSet,
		manager.WireSet,
		triggerer.WireSet,
		file.WireSet,
		converter.WireSet,
		runner.WireSet,
		sse.WireSet,
		scheduler.WireSet,
		commit.WireSet,
		controllertrigger.WireSet,
		plugin.WireSet,
		resolver.WireSet,
		importer.WireSet,
		migrateservice.WireSet,
		canceler.WireSet,
		exporter.WireSet,
		metric.WireSet,
		reposervice.WireSet,
		cliserver.ProvideCodeOwnerConfig,
		codeowners.WireSet,
		gitspaceevent.WireSet,
		cliserver.ProvideKeywordSearchConfig,
		keywordsearch.WireSet,
		rules.WireSet,
		controllerkeywordsearch.WireSet,
		settings.WireSet,
		systemsvc.WireSet,
		usergroup.WireSet,
		openapi.WireSet,
		repo.ProvideRepoCheck,
		audit.WireSet,
		ssh.WireSet,
		publickey.WireSet,
		migrate.WireSet,
		scm.WireSet,
		platformconnector.WireSet,
		gitspacesecret.WireSet,
		orchestrator.WireSet,
		containerorchestrator.WireSet,
		cliserver.ProvideIDEVSCodeWebConfig,
		cliserver.ProvideDockerConfig,
		cliserver.ProvideGitspaceEventConfig,
		logutil.WireSet,
		cliserver.ProvideGitspaceOrchestratorConfig,
		ide.WireSet,
		gitspaceinfraevents.WireSet,
		gitspaceservice.WireSet,
		cliserver.ProvideGitspaceInfraProvisionerConfig,
		cliserver.ProvideIDEVSCodeConfig,
		cliserver.ProvideIDEJetBrainsConfig,
		instrument.WireSet,
		aiagentservice.WireSet,
		aiagent.WireSet,
		capabilities.WireSet,
		capabilitiesservice.WireSet,
		docker.ProvideReporter,
		secretservice.WireSet,
		messagingservice.WireSet,
		runarg.WireSet,
		usage.WireSet,
	)
	return &cliserver.System{}, nil
}
