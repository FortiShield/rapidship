# Nxenv Migrator

**nxenv-migrate** is a command line utility to help convert and migrate
continuous integration pipelines from other providers to Nxenv CI. You can use this tool to migrate repositories following guidlines [below](#migrate-repositories).

**Please review [known conversion and migration issues](KNOWN_ISSUES_CONVERT.md
) before using this tool.**

## Migrate Repositories

Migrating repositoires is a two-step process. 

1. Export: Use `git-export` to export your repositories from your current provider. Guidlines for [Bitbucket On-perm](cmd/stash/README.md), [GitHub](cmd/github/README.md), [Gitlab](cmd/gitlab/README.md), and [Bitbucket](cmd/bitbucket/README.md). The exported data will be saved in a zip file.

2. Import: Import the exported zip file into Nxenv using `git-import` to create new repositories as explained [here](cmd/gitimporter/README.md).

## Migrate Continuous Integration Pipelines 

## Install on Mac

Intel CPU

```sh
curl -L https://github.com/nxenv/rapidship/nxenv-migrate/releases/latest/download/nxenv-migrate-darwin-amd64.tar.gz | tar zx
```

Apple silicon (M1 or M2) CPU

```sh
curl -L https://github.com/nxenv/rapidship/nxenv-migrate/releases/latest/download/nxenv-migrate-darwin-arm64.tar.gz | tar zx
```

Copy the binary into place

```sh
sudo cp nxenv-migrate /usr/local/bin
```

Verify the install

```sh
nxenv-migrate --help
```

## Build

```term
$ git clone https://github.com/nxenv/rapidship/nxenv-migrate.git
$ cd nxenv-migrate
$ go build
```

## Usage

### Drone

Convert a drone pipeline:

```term
nxenv-migrate drone convert /path/to/.drone.yml
```

Export a github namespace from drone:

```term
nxenv-migrate drone export \
  --namespace example \
  --github-token $GITHUB_TOKEN \
  export.json
```

❗ To avoid pipelines triggering in both your Drone instance and in Nxenv CI, you must first deactivate the pipelines in your Drone instance.

This script uses [jq](https://jqlang.github.io/jq/) and the [Drone CLI](https://docs.drone.io/cli/install/) to disable all pipelines defined in your `export.json`:

```bash
#!/bin/bash

DRONE_NAMESPACE=$(jq -r .name export.json)
for REPO_NAME in $(jq -r .project[].name export.json); do
  drone repo disable $DRONE_NAMESPACE/$REPO_NAME
done
```

Import a drone namespace:

```term
nxenv-migrate drone import \
  --nxenv-account $NXENV_ACCOUNT \
  --nxenv-org example \
  --docker-connector account.nxenvImage \
  --github-token $GITHUB_TOKEN \
  export.json
```

### BitBucket

Convert a bitbucket pipeline:

```term
nxenv-migrate bitbucket convert /path/to/bitbucket-pipelines.yml
```

### CircleCI

Convert a circle pipeline:

```term
nxenv-migrate circle convert /path/to/.circle/config.yml
```

### GitHub Actions

Convert a github pipeline:

```term
nxenv-migrate github convert /path/to/.github/workflows/main.yml
```

### GitLab

Convert a gitlab pipeline:

```term
nxenv-migrate gitlab convert /path/to/.gitlab-ci.yml
```

### Terraform

Generate terraform configuration from an export, and apply it to your Nxenv account:

```term
$ nxenv-migrate terraform \
  --account $NXENV_ACCOUNT \
  --docker-connector account.nxenvImage \
  --org $NXENV_ORG \
  --repo-connector $NXENV_REPO_CONNECTOR \
  export.json \
  output.tf
$ terraform init
$ terraform apply
```
