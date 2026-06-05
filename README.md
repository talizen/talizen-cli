# Talizen CLI

Talizen CLI is a thin local bridge for syncing site code between a local directory and Talizen.

The CLI can also run a local Vite preview for pulled Talizen projects. Talizen remains responsible for cloud rendering, CMS, assets, and the realtime preview environment.

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
TALIZEN_API_HOST=http://localhost:8433 talizen login --web=http://localhost:5173
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

This clears the saved token and any saved API host. The next command will use the production default unless you set `TALIZEN_API_HOST`.

## List Projects

```bash
talizen project list
```

For local development:

```bash
TALIZEN_API_HOST=http://localhost:8433 talizen project list
```

Example output:

```text
project_id    Project Name
  project_id/site_id    Site Name
```

Use the `project_id/site_id` value with `pull`, `push`, and `sync`.

## Create Project

Create a new project:

```bash
talizen project create --name="My Project"
```

For local development:

```bash
TALIZEN_API_HOST=http://localhost:8433 talizen project create --name="My Project"
```

You can also create from an existing project or template when the backend allows it:

```bash
talizen project create --name="My Project" --from_id=<project_id>
talizen project create --name="My Project" --tpl_id=<template_id>
```

## Pull Site Code

Download the current remote site files into a local directory:

```bash
talizen pull --site_id=<project_id>/<site_id> --dir=./mysite
```

For local development:

```bash
TALIZEN_API_HOST=http://localhost:8433 talizen pull --site_id=<project_id>/<site_id> --dir=./mysite
```

The command writes remote files such as `/page/...`, `/component/...`, and `talizen.config.ts` into the target directory.

## Local Vite Preview

Talizen projects pulled by the CLI usually do not have their own `package.json`
or `node_modules`. The local preview plugin therefore uses Vite only for local
file serving and TSX transpilation; third-party packages continue to resolve
through the Talizen import map, matching the Web editor preview model. In
`talizen dev`, the CLI loads the platform import map from server system info and
passes it to the Vite plugin; the plugin's local map is only a fallback.

Install Vite in the local project folder:

```bash
cd ./mysite
npm init -y
npm install -D vite esbuild talizen-cli
```

Create `vite.config.mjs`:

```js
import { defineConfig } from 'vite'
import talizen from 'talizen-cli/vite'

export default defineConfig({
  plugins: [
    talizen({
      apiHost: 'https://talizen.com',
      projectId: '<project_id>',
      // token: process.env.TALIZEN_TOKEN,
    }),
  ],
})
```

Run it:

```bash
npx vite --host 0.0.0.0
```

The plugin maps `/page/Index.tsx` to `/`, `/page/About.tsx` to `/about`, starts
from the platform import map, merges `talizen.config.ts` import-map entries,
loads `/index.css` through the Tailwind browser runtime, proxies local `/api/*`
requests to `apiHost`, calls page `getServerSideProps()` in the browser for a
preview-only first render, and uses Vite HMR to re-import the current page
module after local file changes without a full page reload.

## Push Local Changes

Push the current local directory snapshot to Talizen and exit:

```bash
talizen push --site_id=<project_id>/<site_id> --dir=./mysite
```

For local development:

```bash
TALIZEN_API_HOST=http://localhost:8433 talizen push --site_id=<project_id>/<site_id> --dir=./mysite
```

The CLI scans the local directory and calls the existing Talizen `site_action`
API to create or update remote files.

## Sync Local Changes

Run watch mode for a local directory:

```bash
talizen sync --site_id=<project_id>/<site_id> --dir=./mysite
```

For local development:

```bash
TALIZEN_API_HOST=http://localhost:8433 talizen sync --site_id=<project_id>/<site_id> --dir=./mysite
```

`sync` first pushes the current local snapshot, then keeps running and
automatically listens for local file changes. When a file is changed locally,
the CLI calls the existing Talizen `site_action` API and updates the remote site
in realtime. The command also prints the remote preview URL when available.

## Local Web Editor Bidirectional Sync

Run local files and the online Talizen editor against the same cloud realtime
files:

```bash
talizen dev --site_id=<project_id>/<site_id> --dir=./mysite
```

For local backend or web development:

```bash
TALIZEN_API_HOST=http://localhost:8433 talizen dev --web=http://localhost:5173 --site_id=<project_id>/<site_id> --dir=./mysite
```

The command prints the online Web editor URL, pushes local file changes to
Talizen, and listens to the existing WebSocket collaboration channel so editor
changes are written back to the local directory. MVP conflict handling is last
write wins.

`dev` also starts a local Vite preview by default:

```text
  VITE v8.0.14  ready in 529 ms
  ➜  Local:   http://localhost:5173/
Local Vite:  started (preferred http://localhost:5173; use the Vite Local URL above)
```

Use `--preview-port` or `--preview-host` to change the preferred local preview
address. If that port is occupied, Vite uses its normal auto-port behavior and
prints the actual URL in the terminal:

```bash
talizen dev --site_id=<project_id>/<site_id> --dir=./mysite --preview-port=5174
```

Disable the local preview when you only want file sync:

```bash
talizen dev --site_id=<project_id>/<site_id> --dir=./mysite --no-preview
```

The preview uses the bundled `talizen-cli/vite` plugin. If the site directory
has `node_modules/.bin/vite`, that local Vite is used; otherwise the CLI starts
a hidden temporary Vite runtime under `.talizen/` and installs `vite` plus
`esbuild` there.

Local file changes are pushed through Vite HMR as a React root re-render. This
avoids a browser-level refresh, but it is not yet full React Fast Refresh and
does not guarantee component state preservation.

## Open Preview

Open the remote preview URL for a site in the browser:

```bash
talizen preview --site_id=<project_id>/<site_id>
```

For local development:

```bash
TALIZEN_API_HOST=http://localhost:8433 talizen preview --site_id=<project_id>/<site_id>
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
TALIZEN_API_HOST=http://localhost:8433 talizen publish --site_id=<project_id>/<site_id>
```

## Manage CMS Collections

List CMS collections:

```bash
talizen cms collections --site_id=<project_id>/<site_id>
```

Create a collection from a JSON Schema file:

```bash
talizen cms collection create --site_id=<project_id>/<site_id> --key=blogs --name="Blogs" --schema=./blogs.schema.json
```

Update or delete by collection key or id:

```bash
talizen cms collection get --site_id=<project_id>/<site_id> --key=blogs
talizen cms collection update --site_id=<project_id>/<site_id> --key=blogs --schema=./blogs.schema.json
talizen cms collection delete --site_id=<project_id>/<site_id> --key=blogs
```

`--schema` can point to either a raw JSON Schema object or a full collection JSON object containing fields such as `key`, `name`, `desc`, and `json_schema`.

## Manage CMS Content

List, get, create, update, and delete content entries:

```bash
talizen content list --site_id=<project_id>/<site_id> --collection=blogs
talizen content get --site_id=<project_id>/<site_id> --collection=blogs --slug=hello-world
talizen content create --site_id=<project_id>/<site_id> --collection=blogs --data=./content.json --slug=hello-world
talizen content update --site_id=<project_id>/<site_id> --collection=blogs --id=<content_id> --data=./content.json
talizen content delete --site_id=<project_id>/<site_id> --collection=blogs --id=<content_id>
```

`--data` can point to either a plain CMS content body or a full content object. A plain content body may include a business field named `body`. The CLI treats JSON as a full content object only when it includes wrapper fields such as `id`, `slug`, `content_app_id`, `json_schema`, `status`, `sort`, or `tags`.

If your business JSON has a top-level `slug`, do not pass it as plain body JSON because `slug` is a content wrapper field. Either pass the slug as a flag and omit it from `--data`:

```bash
talizen content create --site_id=<project_id>/<site_id> --collection=prompts --data=./content-body.json --slug=typography-v02
```

Or use a full content object and put business fields under `body`:

```json
{
  "slug": "typography-v02",
  "body": {
    "title": "Typography V.02",
    "description": "100vh",
    "tags": ["skill"]
  }
}
```

## Manage Forms

List, create, update, and delete forms:

```bash
talizen form list --site_id=<project_id>/<site_id>
talizen form create --site_id=<project_id>/<site_id> --key=contact-form --name="Contact form" --schema=./contact.schema.json
talizen form get --site_id=<project_id>/<site_id> --key=contact-form
talizen form update --site_id=<project_id>/<site_id> --key=contact-form --schema=./contact.schema.json
talizen form delete --site_id=<project_id>/<site_id> --key=contact-form
```

Inspect and delete form submissions:

```bash
talizen form logs --site_id=<project_id>/<site_id> --key=contact-form
talizen form log get --site_id=<project_id>/<site_id> --key=contact-form --log_id=<log_id>
talizen form log delete --site_id=<project_id>/<site_id> --key=contact-form --log_id=<log_id>
```

Submit a form payload through the platform API:

```bash
talizen form submit --site_id=<project_id>/<site_id> --key=contact-form --data=./payload.json
```

After creating or changing CMS collections or forms, run `talizen pull` again to refresh generated files such as `/types/cms.d.ts` and `/types/form.d.ts` before writing code that imports those types.

## Upload Assets

Upload a local file through the Talizen site asset flow:

```bash
talizen upload --site_id=<project_id>/<site_id> --file=./image.png
```

The command prints the public file URL by default. Use `--json` to inspect the
full upload metadata, including `site_path`, a stable `/_assets/...` path that
can be used from Talizen site code:

```bash
talizen upload --site_id=<project_id>/<site_id> --file=./image.png --json
```

Optional flags:

```bash
talizen upload --site_id=<project_id>/<site_id> --file=./image.png --name=hero.png --mimetype=image/png
```

## Push And Sync Boundary

The current MVP push/sync mode is one-way:

```text
local directory -> Talizen remote site
```

`push` fetches the remote file list to build the local path to remote file id
mapping, scans the local directory, uploads the current local snapshot, and then
exits.

`sync` is watch mode. It performs the same initial local snapshot push, then
keeps running and automatically listens for later local changes.

Neither command pulls Web editor changes back to the local directory while
running. If you edit the same site in the Web editor, run `pull` manually or
restart from a clean local copy before continuing.

Use a test project/site while validating the CLI. Do not run `push` or `sync`
against production content unless the local directory is intended to be the
source of truth.

## Commands

Talizen CLI is a local bridge for Talizen site code. It can authenticate with
Talizen, list projects and sites, pull remote site files into a local directory,
push local files back to Talizen, watch local files for realtime sync, open the
remote preview, and publish a site.

The CLI commands still use the Talizen backend and web app for the canonical
preview. The Vite plugin is a local development helper and intentionally does
not implement full production SSR.

```bash
talizen login [--web=https://talizen.com]
talizen logout
talizen project list
talizen pull --site_id=<project_id>/<site_id> --dir=./mysite
talizen push --site_id=<project_id>/<site_id> --dir=./mysite
talizen sync --site_id=<project_id>/<site_id> --dir=./mysite
talizen dev --site_id=<project_id>/<site_id> --dir=./mysite [--web=https://talizen.com]
talizen preview --site_id=<project_id>/<site_id>
talizen publish --site_id=<project_id>/<site_id> [--note=<note>]
talizen cms collections --site_id=<project_id>/<site_id>
talizen cms collection create --site_id=<project_id>/<site_id> --key=<key> --name=<name> --schema=./schema.json
talizen content list --site_id=<project_id>/<site_id> --collection=<key>
talizen content create --site_id=<project_id>/<site_id> --collection=<key> --data=./content.json
talizen form list --site_id=<project_id>/<site_id>
talizen form create --site_id=<project_id>/<site_id> --key=<key> --name=<name> --schema=./schema.json
talizen upload --site_id=<project_id>/<site_id> --file=./image.png
talizen version
```

Command meanings:

- `login`: Authenticate this machine with Talizen and save a CLI token.
- `logout`: Remove the saved CLI token and API host configuration.
- `project`: List available projects and sites. Use `project_id/site_id` with site commands. Also supports `project create`.
- `pull`: Download the current remote site files into a local directory.
- `push`: Push the current local directory snapshot to the remote site.
- `sync`: Watch mode; push the current snapshot, then keep listening for local changes.
- `dev`: Bidirectionally sync local files with cloud realtime files and the online Web editor.
- `preview`: Open the remote preview URL for a site in the browser.
- `publish`: Publish a site to make the current remote site version live.
- `cms`: Manage CMS collections.
- `content`: Manage CMS content entries.
- `form`: Manage forms and form submissions.
- `upload`: Upload a local file as a Talizen site asset and print its URL.
- `version`: Print the installed CLI version.

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
