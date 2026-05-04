# Talizen CLI

Talizen CLI is a thin local bridge for syncing site code between a local directory and Talizen.

The CLI does not render sites locally. Talizen remains responsible for rendering, CMS, assets, and the realtime preview environment.

## Install

Using npm:

```bash
npm install -g talizen-cli
```

Build from source:

```bash
cd /Users/bysir/dev/bysir/talizen-cli
go build -o talizen ./cmd/talizen
```

Optional:

```bash
mv ./talizen /usr/local/bin/talizen
```

## Login

For production:

```bash
talizen login
```

For local development:

```bash
talizen login --api=http://localhost:8433 --web=http://localhost:5173
```

The command opens a browser authorization page. After authorization succeeds, the CLI stores the token in:

```text
~/Library/Application Support/talizen/config.json
```

The config file contains the API host and CLI token.

When `--web` is omitted, the CLI uses `TALIZEN_WEB_HOST` if set. For local API hosts such as `localhost` or `127.0.0.1`, it defaults to `http://localhost:5173`.
For production, the default API host and default web host are both `https://talizen.com`.

## Logout

Remove the saved CLI config:

```bash
talizen logout
```

This clears the saved token and any saved API host. The next command will use the production default unless you pass `--api` or set `TALIZEN_API_HOST`.

## List Projects

```bash
talizen projects
```

For local development:

```bash
talizen projects --api=http://localhost:8433
```

Example output:

```text
project_id    Project Name
  project_id/site_id    Site Name
```

Use the `project_id/site_id` value with `pull` and `sync`.

## Pull Site Code

Download the current remote site files into a local directory:

```bash
talizen pull --site_id=<project_id>/<site_id> --dir=./mysite
```

For local development:

```bash
talizen pull --api=http://localhost:8433 --site_id=<project_id>/<site_id> --dir=./mysite
```

The command writes remote files such as `/page/...`, `/component/...`, and `talizen.config.ts` into the target directory.

## Sync Local Changes

Watch a local directory and sync local file changes to Talizen:

```bash
talizen sync --site_id=<project_id>/<site_id> --dir=./mysite
```

For local development:

```bash
talizen sync --api=http://localhost:8433 --site_id=<project_id>/<site_id> --dir=./mysite
```

When a file is changed locally, the CLI calls the existing Talizen `site_action` API and updates the remote site in realtime. The command also prints the remote preview URL when available.

## Open Preview

Open the remote preview URL for a site in the browser:

```bash
talizen preview --site_id=<project_id>/<site_id>
```

For local development:

```bash
talizen preview --api=http://localhost:8433 --site_id=<project_id>/<site_id>
```

## Publish Site

Publish a site:

```bash
talizen publish --site_id=<project_id>/<site_id>
```

With a publish note:

```bash
talizen publish --site_id=<project_id>/<site_id> --note="Update homepage copy"
```

For local development:

```bash
talizen publish --api=http://localhost:8433 --site_id=<project_id>/<site_id>
```

## Sync Boundary

The current MVP sync mode is one-way:

```text
local directory -> Talizen remote site
```

When `sync` starts, it fetches the remote file list once to build the local path to remote file id mapping. After that, it watches local files and pushes local changes.

It does not yet pull Web editor changes back to the local directory while running. If you edit the same site in the Web editor, run `pull` manually or restart from a clean local copy before continuing.

Use a test project/site while validating the CLI. Do not run `sync` against production content unless the local directory is intended to be the source of truth.

## Commands

```bash
talizen login [--api=https://talizen.com] [--web=https://talizen.com]
talizen logout
talizen projects [--api=https://talizen.com]
talizen pull --site_id=<project_id>/<site_id> --dir=./mysite [--api=https://talizen.com]
talizen sync --site_id=<project_id>/<site_id> --dir=./mysite [--api=https://talizen.com]
talizen preview --site_id=<project_id>/<site_id> [--api=https://talizen.com]
talizen publish --site_id=<project_id>/<site_id> [--api=https://talizen.com] [--note=<note>]
talizen version
```

## Release

GitHub Releases are created by GitHub Actions when a tag matching `v*` is pushed.
The same workflow publishes the npm package `talizen-cli`.

The release workflow builds binaries for:

- macOS: `darwin/amd64`, `darwin/arm64`
- Linux: `linux/amd64`, `linux/arm64`
- Windows: `windows/amd64`, `windows/arm64`

Create and push a release tag:

```bash
git tag v0.1.0
git push origin v0.1.0
```

Before pushing a release tag, make sure `package.json` has the same version as the
tag without the leading `v`, and configure npm Trusted Publishing for `talizen-cli`
with GitHub repository `talizen/talizen-cli` and workflow filename `release.yml`.

If this repository is mirrored to GitHub with a different remote name, push the tag to that remote:

```bash
git remote add github git@github.com:talizen/talizen-cli.git
git push github main
git push github v0.1.0
```

Local dry run:

```bash
goreleaser release --snapshot --clean
```
