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

Use the `project_id/site_id` value with `pull`, `push`, and `sync`.

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

## Push Local Changes

Push the current local directory snapshot to Talizen and exit:

```bash
talizen push --site_id=<project_id>/<site_id> --dir=./mysite
```

For local development:

```bash
talizen push --api=http://localhost:8433 --site_id=<project_id>/<site_id> --dir=./mysite
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
talizen sync --api=http://localhost:8433 --site_id=<project_id>/<site_id> --dir=./mysite
```

`sync` first pushes the current local snapshot, then keeps running and
automatically listens for local file changes. When a file is changed locally,
the CLI calls the existing Talizen `site_action` API and updates the remote site
in realtime. The command also prints the remote preview URL when available.

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

`--data` can point to either a plain CMS content body or a full content object. A plain content body may include a business field named `body`. The CLI treats JSON as a full content object only when it includes wrapper fields such as `id`, `slug`, `content_app_id`, `json_schema`, `draft_body`, `status`, `sort`, or `tags`.

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

It does not render sites locally. Rendering, CMS, assets, and realtime preview
are handled by the Talizen backend and web app.

```bash
talizen login [--api=https://talizen.com] [--web=https://talizen.com]
talizen logout
talizen projects [--api=https://talizen.com]
talizen pull --site_id=<project_id>/<site_id> --dir=./mysite [--api=https://talizen.com]
talizen push --site_id=<project_id>/<site_id> --dir=./mysite [--api=https://talizen.com]
talizen sync --site_id=<project_id>/<site_id> --dir=./mysite [--api=https://talizen.com]
talizen preview --site_id=<project_id>/<site_id> [--api=https://talizen.com]
talizen publish --site_id=<project_id>/<site_id> [--api=https://talizen.com] [--note=<note>]
talizen cms collections --site_id=<project_id>/<site_id> [--api=https://talizen.com]
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
- `projects`: List available projects and sites. Use `project_id/site_id` with site commands.
- `pull`: Download the current remote site files into a local directory.
- `push`: Push the current local directory snapshot to the remote site.
- `sync`: Watch mode; push the current snapshot, then keep listening for local changes.
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
