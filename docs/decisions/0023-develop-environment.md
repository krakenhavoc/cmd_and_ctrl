# ADR 0023 — A `develop` branch with its own preview deployment

**Status:** Implemented · 2026-09-09 · Branch `chore/develop-environment`

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
`dev.cmd.labxp.io` on every push. Promotion to production is a PR
from `develop` to `main`, which deploys to `cmd.labxp.io`.

`main` remains exactly what it is today, so nothing about the
production path changes shape. The alternative — trunk-based with
flags only — was rejected because the redesign is a coordinated
change across ~20 components; flagging it at that granularity costs
more than a branch does.

### 2. One VPS, two fully namespaced deployments

The dev deployment gets its own service user (`cmdctrl-dev`), env
file (`/etc/cmd_and_ctrl/dev.env`), deploy root
(`/opt/cmd_and_ctrl-dev`), data directory
(`/var/lib/cmd_and_ctrl-dev`), web root
(`/var/www/cmdctrl-client-dev`), port (`:8081`), and vhost. It shares
only the box and a read-only symlink to production's Scryfall bulk
dump — a 2 GB file that is identical in both environments and has its
own refresh cron.

`cmdctrl-dev` is not in the `cmdctrl` group, so the dev process
cannot read production replays or bug-report artifacts. The unit
carries `MemoryMax=768M` and `CPUWeight=50` so a runaway dev process
is OOM-killed rather than pushing production into swap. That cap is
the price of sharing a box; a separate VPS was rejected on cost for a
solo project, and per-PR ephemeral environments on machinery.

### 3. `CMDCTRL_ENV` is the single gate, and it fails closed

`appenv.Parse` maps `CMDCTRL_ENV` onto `prod` (default, including
when unset) or `dev`. An unrecognised value is fatal at boot rather
than a silent fallback: `CMDCTRL_ENV=development` quietly behaving as
prod would strip every dev feature off the dev box with no signal
except a missing button.

CD rewrites `CMDCTRL_ENV` into the env file on **every** deploy of
both branches, so neither environment can drift into the other's
identity through a hand edit. A post-deploy step then curls
`/config` over loopback and fails the run if the service does not
report the environment it was just told to be.

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
features, and folding them in is left for a follow-up.

### 5. Every dev feature is gated server-side; the client flag only decides what to draw

`GET /config` reports `{env, features}` and the client uses it to
decide what to render. That is a UI convenience, **not** the security
boundary. Every dev-only HTTP route is wrapped in `requireDev`, and
dev-only actions are refused by the action registry on the same
check. Flipping a flag in devtools reveals a button that 404s.

`requireDev` answers **404, not 403**: production should not confirm
that a dev route exists.

The rule for anyone adding a dev feature: *if the only thing stopping
a curl against production is a client-side `if`, it is not gated.*
Spawning a card, mutating life, and acting as another seat are
outright cheats on a live table, so this is the one invariant in this
ADR with a test asserting it directly
(`TestLoadFeaturesProdIgnoresOverrides`, `TestRequireDev`).

### 6. Discord: a second redirect URI, and no bot on dev

The existing Discord application gains
`https://dev.cmd.labxp.io/auth/discord/callback` as an additional
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

## Consequences

- Production's reverse proxy must learn to route `/config`. Until it
  does, the client's fetch 404s and `env.ts` falls back to production
  defaults — the correct answer for production — so client and proxy
  can be rolled out in either order.
- `deploy/cmd-and-ctrl.service` does not exist in the repo; the
  production unit has only ever lived on the host. `deploy/README.md`
  documents capturing it with `systemctl cat` rather than
  reconstructing it, because a guessed unit copied over the working
  one is how a preview environment takes down production.
- The nightly e2e workflow still targets production only. Pointing it
  at dev is a follow-up.
- Dev features land behind their flags one at a time: the frame
  inspector ships with this ADR, then the card spawner, then seat
  swap; the replay scrubber comes next.
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
