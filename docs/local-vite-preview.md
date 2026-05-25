# Talizen Local Vite Preview

This document describes how to run a pulled Talizen project locally with Vite.

The local Vite preview is a development helper. It is not the canonical Talizen
renderer and does not implement full production SSR. The canonical preview is
still the remote Talizen preview URL printed by `talizen pull`, `talizen sync`,
or `talizen preview`.

## Why Import Map Instead Of Vite Bundling

Talizen site code pulled by the CLI is not a normal Vite app. It usually has:

- `/page/*.tsx` route files
- `/component/*` reusable components
- `/index.css`
- `talizen.config.ts`
- generated `/types/*.d.ts`
- no `package.json`
- no local `node_modules`

Because of that, Vite should not try to bundle every dependency from local
`node_modules`. The plugin uses Vite only for local serving and TSX transform.
Bare imports such as `react`, `framer-motion`, `lucide-react`, `talizen/cms`,
and `talizen/form` are resolved by the Talizen import map, which matches the Web
editor preview model.

## Install

Pull a site first:

```bash
talizen pull --site_id=<project_id>/<site_id> --dir=./mysite
cd ./mysite
```

Install Vite and the CLI package in that local folder:

```bash
npm init -y
npm install -D vite esbuild talizen-cli
```

For local CLI development, use the checkout path instead:

```bash
npm install -D vite esbuild
npm install -D /Users/bysir/dev/bysir/talizen-cli
```

## Configure Vite

Create `vite.config.mjs` in the pulled site directory:

```js
import { defineConfig } from 'vite'
import talizen from 'talizen-cli/vite'

export default defineConfig({
  plugins: [
    talizen({
      apiHost: 'https://talizen.com',
      projectId: '<project_id>',
      token: process.env.TALIZEN_TOKEN,
    }),
  ],
})
```

For local backend development:

```js
import { defineConfig } from 'vite'
import talizen from 'talizen-cli/vite'

export default defineConfig({
  plugins: [
    talizen({
      apiHost: 'http://localhost:8433',
      projectId: '<project_id>',
      token: process.env.TALIZEN_TOKEN,
    }),
  ],
})
```

`token` is optional. When set, the plugin proxies local `/api/*` requests with
`Authorization: Bearer <token>`. This is useful for CMS/form requests that need
the same auth as the CLI.

## Run

```bash
npx vite --host 0.0.0.0
```

Local file changes are delivered through Vite HMR as a Talizen runtime update.
The browser keeps the same page and the runtime re-imports the current page
module with a fresh timestamp, then re-renders the React root. This avoids a
full page reload. It is not yet equivalent to React Fast Refresh, so component
state is not guaranteed to be preserved.

Open the printed local URL, usually:

```text
http://localhost:5173/
```

The CLI passes `--preview-port` to Vite as the preferred port. It does not force
strict port binding, so if `5173` is already occupied, Vite automatically tries
the next available port and prints the real local URL.

## Routing

The plugin maps Talizen page files to local URLs:

```text
/page/Index.tsx      -> /
/page/About.tsx      -> /about
/page/docs/Index.tsx -> /docs
/page/blog/[slug].tsx -> /blog/:slug
```

Canvas files are ignored:

```text
/page/About.canvas.tsx
```

If a URL does not match a page exactly, the plugin falls back to `/` or the
first available page.

## Runtime Behavior

The plugin:

- injects the Talizen system import map provided by server system info
- merges `talizen.config.ts` `importMap.imports`
- injects `customCode.head`, `customCode.bodyStart`, and `customCode.bodyEnd`
- loads `/index.css` through the Tailwind browser runtime
- transforms local `.tsx`, `.ts`, `.jsx`, and `.js` files with Vite esbuild
- rewrites relative imports to local Talizen module URLs
- serves `?raw` and `?url` imports
- proxies local `/api/*` to `apiHost`
- calls page `getServerSideProps(context)` in the browser before first render

The browser-side `getServerSideProps` context contains:

```js
{
  query: Object.fromEntries(new URLSearchParams(window.location.search)),
  params: routeParams,
}
```

## Limitations

This preview intentionally does not implement the full render service:

- no production SSR
- no server-only execution sandbox
- no exact production HTML output
- no production Tailwind build pipeline
- no remote file sync by itself

Use it for fast local visual checks. Use `talizen sync` or `talizen preview` for
the canonical remote preview that matches Talizen production behavior.

## Minimal Workflow

```bash
talizen pull --site_id=<project_id>/<site_id> --dir=./mysite
cd ./mysite
npm init -y
npm install -D vite esbuild talizen-cli
cat > vite.config.mjs <<'EOF'
import { defineConfig } from 'vite'
import talizen from 'talizen-cli/vite'

export default defineConfig({
  plugins: [
    talizen({
      apiHost: 'https://talizen.com',
      projectId: '<project_id>',
      token: process.env.TALIZEN_TOKEN,
    }),
  ],
})
EOF
npx vite --host 0.0.0.0
```
