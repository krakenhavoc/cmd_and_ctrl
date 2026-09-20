# ADR 0023 — A `develop` branch with its own preview deployment

**Status:** Implemented · 2026-09-09 · Branch `chore/develop-environment`
**Revised:** 2026-09-09 · §2 replaced · Branch `chore/dev-env-second-vm`
**Revised:** 2026-09-16 · §4 and Consequences corrected (nightly e2e
runs locally; both follow-ups dropped) · Branch `docs/adr-0023-nightly-e2e`
**Amended by:** [ADR 0075](0075-table-settings-and-host-controls.md) (2026-09-19,
S35). Spawning is no longer dev-only. A PRODUCTION table may opt in to it
through the visible `allow_spawn` table setting, and then the host or the admin
may spawn through `POST /games/{id}/spawn` — announced in the game log and
undoable. Nothing below changes: the dev route `POST /games/{id}/dev/spawn`
keeps its `requireDevFeature` gate and its anyone-at-the-table semantics, and
§4's rule that a dev-only feature is gated server-side is untouched. What the
amendment revises is the *claim* in §7 and §8 that spawning is inherently a
cheat on a live table — it is, unless the table has said otherwise out loud and
can watch every use of it.

## Context

Every change currently lands on `main` and deploys straight to
`cmd.labxp.io`, which is the table people actually play on. There is
no place to look at a feature running against real data before it
becomes the thing everyone is playing. Two consequences show up
repeatedly:

- The Sept 2026 UX redesign (`design/foundations`) is a large visual
  change with no way to live with it for a week before committing.
- Card-implementation batches are verified by unit tests and a local
  server, then meet real players for the first time in production.

We also want a class of tooling that is *actively dangerous* in
production — spawning arbitrary cards, mutating life totals, driving
another player's seat — but which makes testing a four-player game
tractable for one person.

## Decisions

### 1. `develop` is the integration branch; `main` stays the release branch

Feature branches PR into `develop`. `develop` deploys to
`cmd-dev.labxp.io` on every push. Promotion to production is a PR
from `develop` to `main`, which deploys to `cmd.labxp.io`.

`main` remains exactly what it is today, so nothing about the
production path changes shape. The alternative — trunk-based with
flags only — was rejected because the redesign is a coordinated
change across ~20 components; flagging it at that granularity costs
more than a branch does.

### 2. Two VMs, one definition *(revised)*

The develop preview is its own Proxmox VM, `cmd-and-ctrl-dev`, behind
`cmd-dev.labxp.io`. It is declared in the HomeLab repo as a second
entry in a `for_each` over an environments map, sharing the cloud-init
template, the systemd unit, the Caddy config and the data-disk layout
with production.

**The two hosts are identical below the fqdn.** Same paths, same
service name, same service user, same listen port. The only
differences are hostname, fqdn, the two tokens, and `CMDCTRL_ENV`.
That is the property worth having: the CD recipe that deploys to the
preview is byte-for-byte the recipe that deploys to production, so a
preview deploy actually rehearses the thing it is meant to rehearse.
An environment reached by a *different* deploy path is testing the
deploy path as much as the code.

> **This section replaces the original.** The first version put a
> second service unit on production's box — own user, own env file,
> `/opt/cmd_and_ctrl-dev`, port `:8081`, `MemoryMax=768M`, and a
> read-only symlink to production's Scryfall dump — and justified the
> sharing on the cost of a second rented VPS. The host is not a rented
> VPS. It is a Proxmox node driven by Terraform, where another VM is a
> map entry and its Cloudflare tunnel can be declared beside it. Once
> that was clear the comparison inverted, and the namespacing, the
> memory cap and the provisioning script all became machinery serving
> a constraint that did not exist.
>
> The isolation argument stands on its own regardless of cost: sharing
> a box meant a runaway preview process could push production into
> swap, and `MemoryMax` bounded that rather than removing it. Separate
> VMs remove it.

The cost is ~4 GB of RAM and a 20 GB data disk on the Proxmox node,
plus a second Scryfall dump on its own refresh timer. Per-PR ephemeral
environments were still rejected — dynamic hostnames, tunnel
provisioning and teardown are a lot of machinery for a solo project.

### 3. `CMDCTRL_ENV` is the single gate, and it fails closed

`appenv.Parse` maps `CMDCTRL_ENV` onto `prod` (default, including
when unset) or `dev`. An unrecognised value is fatal at boot rather
than a silent fallback: `CMDCTRL_ENV=development` quietly behaving as
prod would strip every dev feature off the dev box with no signal
except a missing button.

The value is written three times over, deliberately. Terraform stamps
it into the env file at build; CD rewrites it on **every** deploy of
both branches, so a hand edit cannot outlive a restart; and a
post-deploy step curls `/config` over loopback and fails the run if
the service does not report the environment it was just told to be.

### 4. Dev features cannot be enabled in production, even deliberately

`appenv.LoadFeatures` returns the zero value whenever the environment
is not `dev`, before consulting any per-feature variable. Inside a
dev deployment every feature defaults **on**, and each
`CMDCTRL_DEV_*` variable can only turn one **off** — which is how you
rehearse what production will look like without a surface.

This asymmetry is the point. A `CMDCTRL_DEV_CARD_SPAWN=1` line
copy-pasted into the production env file is inert; `main.go` logs it
at WARN on boot so the misconception gets corrected. The existing
`CMDCTRL_DEV_SKIP_DECK_VALIDATION` and
`CMDCTRL_DEV_RELAX_RATE_LIMITS` knobs predate this and remain
independent — they are ops escape hatches, not player-visible
features. Folding them under `CMDCTRL_ENV` was considered and dropped
(2026-09-16): their users — the Playwright web server,
`make dev-skip-validation`, the lobby tests — run with `CMDCTRL_ENV`
unset, so a hard gate would mean setting `dev` there and switching on
all four dev features just to keep an ops knob working.

### 5. Every dev feature is gated server-side; the client flag only decides what to draw

`GET /config` reports `{env, features}` and the client uses it to
decide what to render. That is a UI convenience, **not** the security
boundary. Every dev-only HTTP route is wrapped in `requireDev`, and
dev-only actions are refused by the action registry on the same
check. Flipping a flag in devtools reveals a button that 404s.

`requireDev` answers **404, not 403**: production should not confirm
that a dev route exists. It also wraps **outside** `auth.Middleware`,
because wrapped the other way an unauthenticated probe gets 401 —
which confirms the route exists just as surely.

The rule for anyone adding a dev feature: *if the only thing stopping
a curl against production is a client-side `if`, it is not gated.*
Spawning a card, mutating life, and acting as another seat are
outright cheats on a live table, so this is the one invariant in this
ADR with a test asserting it directly
(`TestLoadFeaturesProdIgnoresOverrides`, `TestRequireDev`,
`TestDevRoutesAreAbsentInProduction`).

### 6. Discord: a second redirect URI, and no bot on dev

The existing Discord application gains
`https://cmd-dev.labxp.io/auth/discord/callback` as an additional
redirect URI, so sign-in works on dev with no second app to keep in
sync.

The develop deployment **never ships or restarts the bot**. Two bot
processes logged into the same application both answer every slash
command; this is enforced in CD with `if: env.IS_DEV != 'true'` on
both the deploy and the restart step, not left to discipline. Slash
commands are tested against production or a private test guild.

### 7. The environment is visible in the UI, permanently

`EnvBadge` renders a fixed corner badge whenever `env === "dev"`. It
is not dismissible, because a badge you can hide is hidden exactly
when it matters. The failure it guards against is human: filing a bug
against dev thinking it was prod, or assuming prod is dev and
spawning a card mid-game.

### 8. The two admin tokens are different, and Terraform enforces it

The preview exposes card spawning and seat swapping to any admin
session. Sharing production's `CMDCTRL_ADMIN_TOKEN` would make a leak
from the lower-trust box a compromise of the live table, so the
HomeLab variable carries a validation rejecting equal tokens.

## Consequences

- Both Caddy sites must route `/config`, and the dev site also `/dev`.
  Until they do, the client's fetch gets `index.html` back, fails to
  parse, and falls back to production defaults — the correct answer
  for production, and a silently featureless preview on dev.
- Reading the cloud-init template to add the dev VM turned up a
  production gap: the Caddyfile proxied only `/ws`, `/healthz`,
  `/games*`, `/cards/*`, `/admin/*` and `/me`. Discord OAuth, avatars,
  bug reports and logout were never routed, so those features depended
  on a hand-edit that a VM rebuild would have reverted. Fixed in the
  same change, along with `CMDCTRL_TRUST_FORWARDED` — without which
  every client shares one rate-limit bucket, since cloudflared
  forwards from `127.0.0.1`.
- `deploy/cmd-and-ctrl.service` does not exist in this repo; the unit
  is written by the HomeLab cloud-init template, which is now the
  single source of truth for both hosts. `deploy/README.md` points
  there.
- The nightly e2e workflow (`e2e-nightly.yml`) targets neither host.
  Both its jobs — the bot whole-game tests and Playwright — run on the
  self-hosted runner, and Playwright starts its own `go run` server and
  Vite dev server there (`tests-e2e/playwright.config.ts`). An earlier
  revision of this ADR said it targeted production; it does not.
  Running e2e against the preview was considered and dropped
  (2026-09-16): it would need the preview's admin token in CI and
  relaxed rate limits on a public host. The nightly stays local.
- Dev features land behind their flags one at a time: the frame
  inspector shipped with this ADR, then the card spawner, then seat
  swap, then the replay scrubber. Three of the four turned out to be
  client-only; only the card spawner needed a server route. The
  action-registry gate this ADR anticipated was never built, because
  nothing has needed it.
- Seeded shuffles were scoped with the scrubber and dropped. The card
  spawner reaches a chosen board state directly, which is most of
  what a deterministic shuffle was wanted for; the rest did not
  justify threading a seed through the game-start RNG.
- Seat swap needed no server change at all, contrary to the estimate
  when this ADR was written. `WSAuthorizer` already accepts
  `?player=` for an admin session and the hub already stamps the
  bound seat onto every action, so an admin could always drive any
  seat by hand-editing the WebSocket URL. The dev feature is the
  control, not the capability — which is why its flag gates a client
  component and nothing else.
- The card spawner is HTTP (`GET /dev/cards`,
  `POST /games/{id}/dev/spawn`) rather than a new WS action type. The
  card index already lives in `lobby.Config`, the lobby already owns
  the mutate-and-broadcast path that deck upload uses, and
  `requireDevFeature` already gates HTTP. Routing it through
  `actions.Dispatch` would have meant threading the index into a
  package that deliberately depends on nothing but `game`, plus a
  second environment gate on the WS side — more surface for the same
  result. No dev feature has needed that gate yet — seat swap, the
  obvious candidate, turned out to need no server change — so it
  stays unbuilt until one does.
