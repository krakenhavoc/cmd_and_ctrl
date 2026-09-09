# Environments

Two deployments, one VPS. Rationale in
[ADR 0023](decisions/0023-develop-environment.md).

| | production | develop preview |
|---|---|---|
| Branch | `main` | `develop` |
| URL | `https://cmd.labxp.io` | `https://dev.cmd.labxp.io` |
| `CMDCTRL_ENV` | `prod` | `dev` |
| systemd unit | `cmd-and-ctrl` | `cmd-and-ctrl-dev` |
| Service user | `cmdctrl` | `cmdctrl-dev` |
| Bind address | `:8080` | `:8081` |
| Env file | `/etc/cmd_and_ctrl/env` | `/etc/cmd_and_ctrl/dev.env` |
| Deploy root | `/opt/cmd_and_ctrl` | `/opt/cmd_and_ctrl-dev` |
| Data dir | `/var/lib/cmd_and_ctrl` * | `/var/lib/cmd_and_ctrl-dev` |
| Web root | `/var/www/cmdctrl-client` | `/var/www/cmdctrl-client-dev` |
| Discord bot | yes | **never** (see below) |
| Dev features | none, unconditionally | all, by default |

\* whatever production's `CMDCTRL_DATA_DIR` actually points at —
confirm with `systemctl cat cmd-and-ctrl` before relying on it.

## Branch flow

```
feat/* ──PR──► develop ──deploy──► dev.cmd.labxp.io
                  │
                 PR
                  ▼
                main ──deploy──► cmd.labxp.io
```

Feature branches PR into `develop`. Promotion is a `develop` → `main`
PR. Hotfixes may PR straight into `main`; cherry-pick or merge back
into `develop` afterwards so it never falls behind.

## Dev-only features

Enabled only when `CMDCTRL_ENV=dev`. Each may be individually turned
**off** on the dev box to rehearse production behaviour; none can be
turned **on** in production.

| Feature | Variable | Status |
|---|---|---|
| Raw protocol frame inspector | `CMDCTRL_DEV_FRAME_INSPECTOR` | shipped |
| Card spawn | `CMDCTRL_DEV_CARD_SPAWN` | shipped |
| Drive all seats from one browser | `CMDCTRL_DEV_SEAT_SWAP` | shipped |
| Replay scrubber | `CMDCTRL_DEV_REPLAY_SCRUBBER` | shipped |

`GET /config` reports the live answer:

```console
$ curl -s https://dev.cmd.labxp.io/config
{"env":"dev","features":{"card_spawn":true,"seat_swap":true,"frame_inspector":true,"replay_scrubber":true}}
$ curl -s https://cmd.labxp.io/config
{"env":"prod","features":{"card_spawn":false,"seat_swap":false,"frame_inspector":false,"replay_scrubber":false}}
```

### Card spawner

Open the **SPAWN** tab in the dev dock (bottom-left of the game
screen). Search by card name, pick a seat and a zone, spawn. The two
routes behind it:

```console
$ curl -s "$DEV/dev/cards?q=blood artist" -H "Authorization: Bearer $TOKEN"
$ curl -s -X POST "$DEV/games/$GAME/dev/spawn" -H "Authorization: Bearer $TOKEN" \
    -d '{"name":"Blood Artist","player_id":"'"$SEAT"'","zone":"battlefield","count":2}'
```

Worth knowing:

- Spawned cards go through `deck.ToGameCard`, the same conversion the
  deck importer uses, so they carry a real OracleID and the S14+
  effect catalog matches them. A spawned Blood Artist actually
  watches for deaths.
- **Battlefield spawns fire ETB triggers**, so a spawn can put things
  on the stack. That is the point — but it means the resulting
  snapshot may differ by more than the cards you asked for.
- The game must be **active**. Spawning into a lobby-state game would
  be erased by `Start` dealing opening hands.
- Zones: `battlefield`, `hand`, `graveyard`, `exile`, `library`,
  `command`. The stack is refused — spawn to hand and cast it, so the
  card gets a real cast context.
- Count is capped at 20 per request.
- Spawns route through the room like every other mutation, so they
  land in the replay: a bug found on a spawned board is still
  reproducible from the recording.

Life totals, counters, poison, energy and phase are **not** here —
an admin session can already set those on any seat through the
ordinary action protocol, and a second UI for them would just be a
duplicate to keep in sync.

### Seat swap

The **SEAT** dropdown in the dev dock switches which seat you view
and act as. Admin sessions only.

This adds no server capability — an admin session could always bind
to any seat by putting `?player=<uuid>` on the WebSocket URL
(`WSAuthorizer`, `case auth.RoleAdmin`). What the control adds is one
click instead of hand-editing a URL, which is the difference between
the capability existing and it being used.

Two things that surprise people:

- **Binding to a seat gives up the admin bypass.** The hub stamps the
  bound seat onto every action as its caller, so player-scoped gates
  (priority, active player, "you may not act for another seat") now
  apply to you as that seat. That is usually what you want when
  testing what a seat can do — but if you need the moderator
  override, switch back to *— spectate —*.
- **Switching reconnects.** `ws.ts` treats a URL change as a retarget
  and resets the seq watermark and rendered snapshot. The new seat
  starts from its own first snapshot, so a brief blank frame is
  expected, not a bug.

A player or spectator session never sees the control, and forging the
parameter would not help: `WSAuthorizer` resolves the seat from the
principal for those roles and ignores `?player=` entirely.

### Replay scrubber

The **REPLAY** tab loads this game's recorded snapshots and steps the
board through them. Transport controls, a slider, playback at four
speeds, and a per-frame readout of what changed.

No new server route: `GET /games/{id}/replay` has existed since S11
and already streams JSONL of complete snapshots, one per applied
action. Rendering frame N is picking a line and handing it to the
same board the live game uses — there is no re-simulation, so what
you see is exactly what the server recorded.

- **Frames are unfiltered.** A replay shows every hand and every
  library. That is the point when debugging, and why the route is
  admin-gated while a game is in progress.
- **Scrubbing hides the quick-action toolbar.** Those buttons act on
  the *live* game; leaving them clickable while the board shows a
  past state is the one way a read-only inspection tool could do
  damage.
- **The log grows while you watch.** "reload" picks up frames added
  since you loaded. A download taken mid-write can end in a
  half-written line; the parser drops it and keeps everything before.
- Only actions produce frames. A game nobody has acted on has an
  empty replay, and the panel says so rather than erroring.

Seeded shuffles (`CMDCTRL_DEV_SEED`) were scoped alongside this and
deliberately dropped: the card spawner reaches a specific board state
directly, which is what the seed was mostly wanted for, and the
remaining value did not justify threading a seed through the
game-start RNG. Say so if you want it.

**Adding a dev feature.** Add the field to `appenv.Features`, register
its variable in `featureVars`, mirror the JSON key in
`client/src/lib/env.ts`, and wrap every route it needs in
`requireDev` / `requireDevFeature`. If the only thing stopping a curl
against production is a client-side `if`, it is not gated.

## One-time host provisioning

Run once on the VPS. `deploy/provision-dev.sh` does everything that
can be automated; the reverse proxy, TLS, Discord and GitHub steps
below are the remainder.

```sh
sudo deploy/provision-dev.sh --check   # report only, changes nothing
sudo deploy/provision-dev.sh
```

It is idempotent — re-run it after a failed step, or to fix drift.
Every value it writes is written only when absent, so a second run
never rotates a token out from under a live session.

What it does: creates the `cmdctrl-dev` service user, the deploy and
web roots (owned by the user CI rsyncs as), and the data directory;
writes `/etc/cmd_and_ctrl/dev.env` with a **freshly generated** admin
token distinct from production's; installs and enables the systemd
unit; and symlinks production's Scryfall dump read-only rather than
downloading a second 2 GB copy.

Three things it deliberately does *not* do:

- **Copy production's admin token.** It generates one. Read it back
  with `sudo grep CMDCTRL_ADMIN_TOKEN /etc/cmd_and_ctrl/dev.env`.
- **Copy `CMDCTRL_GITHUB_TOKEN`.** CD provisions that, and a preview
  deployment filing bug reports into the real tracker is noise.
- **Write to `/etc/cmd_and_ctrl/env`.** It reads production's env file
  — that is how it learns the real `CMDCTRL_DATA_DIR` instead of
  asking you to substitute it by hand — and never writes there.

It refuses to start if any dev path resolves to its production
counterpart, so a mistyped override cannot end up chowning or
symlinking inside production.

The Discord client id and secret *are* copied from production, on
purpose: same application, one extra redirect URI (ADR 0023 §6).

### Reverse proxy (Caddy)

```sh
sudo cp deploy/caddy/dev.cmd.labxp.io.caddyfile /etc/caddy/conf.d/
sudo caddy validate --config /etc/caddy/Caddyfile
sudo systemctl reload caddy
```

If that host has no `conf.d` (and no `import conf.d/*` in its
Caddyfile), paste the site block into `/etc/caddy/Caddyfile` instead.
**TLS is automatic** — Caddy provisions and renews the certificate
itself, so there is no certbot step; the DNS record just has to
resolve to this host first.

**Production needs a one-line change too.** `GET /config` is a new
route: add `/config` to whatever matcher already routes `/games`,
`/cards` and `/me` to `127.0.0.1:8080`.

Rollout order does not matter, but the failure is silent in both
directions, so it is worth knowing what it looks like:

| Missing | Symptom |
|---|---|
| `/config` on either site | the client falls back to production defaults — on dev, every dev feature stays hidden with no error |
| `/dev` on the dev site | the card search renders "No matches" forever (`searchDevCards` swallows failures by design) |

The same trap applies to `client/vite.config.ts` for local
development: a new top-level route needs an entry there *and* in the
`@api` matcher.

One thing not to "fix": the site block deliberately does **not** set
`header_up Host`. Caddy forwards the original Host by default, and
the game server's WebSocket `CheckOrigin` compares the Origin
header's host against the request Host. Rewriting Host to the
upstream makes every `/ws` upgrade 403 while ordinary HTTP keeps
working.

### Discord

Add `https://dev.cmd.labxp.io/auth/discord/callback` to the existing
Discord application's redirect URIs. Same client ID and secret.

**Do not run the bot on dev.** Two processes logged into the same
Discord application both answer every slash command. CD enforces this
(`if: env.IS_DEV != 'true'` on the bot deploy and restart steps) —
don't work around it by installing the bot unit on the dev deployment
by hand.

### GitHub repo variables

| Variable | Purpose |
|---|---|
| `CMDCTRL_HOST` | deploy target, shared by both environments |
| `CMDCTRL_PUBLIC_BASE_URL` | production origin (default `https://cmd.labxp.io`) |
| `CMDCTRL_DEV_PUBLIC_BASE_URL` | dev origin (default `https://dev.cmd.labxp.io`) |

Set branch protection on both `main` and `develop`: require the
`cmdctrl-ci` check, and require a PR for `main`.

## Operating notes

**Which environment am I looking at?** The client shows a fixed `DEV`
badge in the bottom-right corner on dev, and nothing on production.
`curl -s <host>/config` is the authoritative answer.

**Reading logs.**

```sh
journalctl -u cmd-and-ctrl-dev -f      # dev
journalctl -u cmd-and-ctrl -f          # production
```

The dev server logs a boot WARN naming its environment and feature
set. If the journal and the browser badge disagree, the browser is
talking to a different backend than you think — check the vhost.

**Resetting dev state.** Dev games are disposable:

```sh
sudo systemctl stop cmd-and-ctrl-dev
sudo find /var/lib/cmd_and_ctrl-dev/games -mindepth 1 -delete
sudo systemctl start cmd-and-ctrl-dev
```

Never point this at `/var/lib/cmd_and_ctrl`.

**A deploy failed at "Verify reported environment".** The service is
running but reports the wrong `CMDCTRL_ENV`. Check the env file named
in the error; the most likely cause is a hand edit that CD then
rewrote, followed by a restart that did not happen. This check
failing on `main` is the serious direction — production came up with
dev features reachable — so treat it as an incident, not a flake.
