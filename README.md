# Talizen CLI

Talizen CLI is a thin local bridge for syncing site code between a local directory and Talizen.

The CLI does not render sites locally. Talizen remains responsible for rendering, CMS, assets, and the realtime preview environment.

## Install

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
talizen projects [--api=https://talizen.com]
talizen pull --site_id=<project_id>/<site_id> --dir=./mysite [--api=https://talizen.com]
talizen sync --site_id=<project_id>/<site_id> --dir=./mysite [--api=https://talizen.com]
```
