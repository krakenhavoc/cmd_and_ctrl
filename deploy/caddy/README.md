# deploy/caddy/

Caddy site configuration. See [`docs/environments.md`](../../docs/environments.md).

| File | Site |
|---|---|
| `dev.cmd.labxp.io.caddyfile` | the develop preview environment |

## Production needs a one-line change too

`GET /config` is a new route (ADR 0023). Production's existing site
block predates it, so add `/config` to whatever matcher already routes
`/games`, `/cards`, `/me` and friends to `127.0.0.1:8080`.

Until that lands, production serves `index.html` for `/config`, the
client's fetch fails to parse, and `lib/env.ts` falls back to the
production defaults — which is the *correct* answer for production,
so nothing breaks. It just means the environment probe is answering by
accident rather than on purpose, and `curl -s https://cmd.labxp.io/config`
returns HTML instead of JSON.

## Production's site block is not in this repo

Same situation as `deploy/cmd-and-ctrl.service`: it lives only on the
host, written by the HomeLab cloud-init template. Capture it before
treating this directory as the source of truth for production:

```sh
ssh krkn@$CMDCTRL_HOST 'cat /etc/caddy/Caddyfile' > deploy/caddy/Caddyfile.prod
```

Review, strip anything host-specific or secret, and commit. Until
then, do not overwrite the running config with anything from here.
