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
| Replay scrubber + seeded shuffles | `CMDCTRL_DEV_REPLAY_SCRUBBER` | planned |

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

**Adding a dev feature.** Add the field to `appenv.Features`, register
its variable in `featureVars`, mirror the JSON key in
`client/src/lib/env.ts`, and wrap every route it needs in
`requireDev` / `requireDevFeature`. If the only thing stopping a curl
against production is a client-side `if`, it is not gated.

## One-time host provisioning

Run once, as a user with sudo on the VPS. Nothing here is repeated by
CD; CD assumes it has been done.

```sh
# 1. Service user and directories.
sudo useradd --system --no-create-home --shell /usr/sbin/nologin cmdctrl-dev
sudo mkdir -p /opt/cmd_and_ctrl-dev/server/bin \
              /opt/cmd_and_ctrl-dev/scripts \
              /var/www/cmdctrl-client-dev
sudo chown -R krkn:krkn /opt/cmd_and_ctrl-dev /var/www/cmdctrl-client-dev

# 2. Env file. Mode 0640, group cmdctrl-dev — CD rewrites
#    CMDCTRL_ENV, CMDCTRL_PUBLIC_BASE_URL and the bug-report token
#    into it on every deploy; the rest is set once here.
sudo install -o root -g cmdctrl-dev -m 0640 /dev/null /etc/cmd_and_ctrl/dev.env
sudo tee -a /etc/cmd_and_ctrl/dev.env >/dev/null <<'ENV'
CMDCTRL_ENV=dev
CMDCTRL_ADDR=:8081
CMDCTRL_DATA_DIR=/var/lib/cmd_and_ctrl-dev
CMDCTRL_ADMIN_TOKEN=<a DIFFERENT long random string from production>
CMDCTRL_SECURE_COOKIES=1
CMDCTRL_TRUST_FORWARDED=1
CMDCTRL_DISCORD_CLIENT_ID=<same app as prod>
CMDCTRL_DISCORD_CLIENT_SECRET=<same app as prod>
CMDCTRL_DISCORD_REDIRECT_URI=https://dev.cmd.labxp.io/auth/discord/callback
ENV

# 3. systemd unit.
sudo cp deploy/cmd-and-ctrl-dev.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now cmd-and-ctrl-dev

# 4. Share the Scryfall dump read-only rather than downloading a
#    second 2 GB copy. Replace PROD_DATA with production's real
#    CMDCTRL_DATA_DIR.
PROD_DATA=/var/lib/cmd_and_ctrl
sudo ln -s "$PROD_DATA/scryfall" /var/lib/cmd_and_ctrl-dev/scryfall
sudo chgrp -R cmdctrl "$PROD_DATA/scryfall"
sudo chmod -R g+rX "$PROD_DATA/scryfall"
sudo usermod -aG cmdctrl cmdctrl-dev   # read-only: the dir is g+rX, not g+w
```

> The last line is the one place the two environments touch. If you
> would rather they touch nowhere at all, drop it and give dev its own
> `scryfall-refresh.sh` cron and its own 2 GB copy.

### Reverse proxy

Add a vhost for `dev.cmd.labxp.io` mirroring the production one,
proxying to `127.0.0.1:8081` with the client dist at
`/var/www/cmdctrl-client-dev`, plus a TLS certificate.

**Both vhosts need a new location for `/config`.** Production's needs
it too — add it alongside `/games`, `/cards`, `/me`, `/auth`,
`/avatars`, `/admin`, `/bugreport`, `/healthz`. Order does not
matter: until the proxy routes it the client's fetch 404s and falls
back to production defaults, which is the correct answer for
production and merely means dev features stay hidden on dev.

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
