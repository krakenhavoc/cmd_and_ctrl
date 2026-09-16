# Environments

Two VMs on the same Proxmox node, identical below the fqdn. Rationale in
[ADR 0023](decisions/0023-develop-environment.md).

| | production | develop preview |
|---|---|---|
| Branch | `main` | `develop` |
| URL | `https://cmd.labxp.io` | `https://cmd-dev.labxp.io` |
| VM | `cmd-and-ctrl` | `cmd-and-ctrl-dev` |
| `CMDCTRL_ENV` | `prod` | `dev` |
| Actions variable | `CMDCTRL_HOST` | `CMDCTRL_DEV_HOST` |
| Discord bot | yes — binary, unit and `bot.env` from CD; [deploy/README.md](../deploy/README.md#discord-bot-production-only) | **never** (see below) |
| Dev features | none, unconditionally | all, by default |

Everything else is the same on both: `/opt/cmd_and_ctrl` for artifacts,
`/var/www/cmdctrl-client` for the client, `/etc/cmd_and_ctrl/env`,
`/var/lib/cmd_and_ctrl/data` on a dedicated disk, the `cmdctrl` user,
`cmd-and-ctrl.service`, and `127.0.0.1:8080` behind Caddy.

That sameness is the point. The CD recipe that deploys the preview is
the recipe that deploys production, with only the target host swapped,
so a preview deploy rehearses the real thing rather than a lookalike.

## Branch flow

```
feat/* ──PR──► develop ──deploy──► cmd-dev.labxp.io
                  │
                 PR
                  ▼
                main ──deploy──► cmd.labxp.io
```

Feature branches PR into `develop`, with `cmdctrl-ci` required. `main`
is still the repo's default branch, so pass `--base develop` to
`gh pr create`.

There is no merge queue. GitHub offers one only on organization-owned
repositories and this one is user-owned, so the `merge_group` trigger
in `ci-cd.yml` (#430) never fires. Two PRs that are each green against
a stale `develop` can still break it together; `develop` breaking is
what the preview is for, and it is caught before promotion rather
than on the live table.

Promotion is a `develop` → `main` PR, merged with a **merge commit** —
the only method `main` allows. A squash would give `main` a commit
`develop` does not have, and the two drift further apart with every
promotion.

Hotfixes may PR straight into `main` from a `hotfix/*` branch. Then
open a `main` → `develop` PR so `develop` never falls behind.

### Merge methods

| PR | Method | Enforced |
|---|---|---|
| `feat/*`, `fix/*`, … → `develop` | squash | convention |
| `develop` → `main` (promotion) | merge commit | ruleset |
| `hotfix/*` → `main` | merge commit | ruleset |
| `main` → `develop` (hotfix back-merge) | merge commit | convention |

**Squash into `develop`** so each PR is one commit carrying its PR
number, revertable in one `git revert`, without a branch's worth of
"address review" commits.

**Merge commits everywhere else** so no commit is ever copied between
the two branches. A merge commit keeps `develop` an ancestor of `main`,
and `git log --first-parent main` reads as the list of releases.

**The back-merge is the one to watch.** `develop` allows both methods,
so GitHub will offer squash there too. A squashed back-merge gives
`develop` a copy of the hotfix rather than `main`'s commit. That merges
cleanly until `develop` touches the same lines, and then the next
promotion conflicts over a change both branches already have. `develop`
is not squash-only for exactly this reason: that would force the
back-merge to squash.

The rulesets enforce this. `main` requires `promotion-guard`, which
fails any PR whose head is not `develop` or `hotfix/*`. If it fires on
yours, retarget it: `gh pr edit <n> --base develop`. Neither branch
accepts direct pushes.

## The host is Terraform

Both VMs, the develop Cloudflare tunnel, its ingress config and its DNS
record live in
[krakenhavoc/HomeLab](https://github.com/krakenhavoc/HomeLab) under
`terraform/deployments/lab/`. There is no provisioning script and
nothing to run by hand: a `terraform apply` builds the VM, cloud-init
writes the env file, the systemd unit and a first-boot placeholder
Caddyfile, and cloudflared registers itself.

To change the host — the memory size, the disk layout, a boot-time env
var — edit `templates/setup-cmd_and_ctrl.yaml.tftpl` or the tfvars and
apply. Do not edit files on the box; a rebuild reverts them, which is
exactly how production ended up serving Discord OAuth from a Caddyfile
that the template did not contain.

**Exceptions owned by this repo and installed by the CD job on every
deploy.** Caddy's routing table is
[`deploy/Caddyfile`](../deploy/Caddyfile) — cloud-init writes only the
first-boot placeholder and cannot update a running host, since the VM
carries `ignore_changes` on its cloud-init. Server env vars are written
by `scripts/set-server-env.sh` from the "Sync server env" steps. On
production only, the Discord bot's binary directory, its systemd unit
and `/etc/cmd_and_ctrl/bot.env` come from the bot steps. Change any of
them here, not in HomeLab.

The bot leaves one exception to "nothing to run by hand": creating the
`cmdctrl-bot` user and enabling the unit on production. The HomeLab
template does not create the user yet (the HomeLab half of #249), so
these steps are needed on the running production VM and **again after
every rebuild of production**, until the template does. A rebuilt host
without them still deploys green, with no bot. See
[deploy/README.md](../deploy/README.md#operator-setup-running-vm-and-every-rebuild).

### Secrets

| Where | Name | Notes |
|---|---|---|
| HomeLab | `CMD_AND_CTRL_ADMIN_TOKEN` | production |
| HomeLab | `CMD_AND_CTRL_DEV_ADMIN_TOKEN` | preview; Terraform rejects it if equal to production's |
| HomeLab | `CMD_AND_CTRL_TUNNEL_TOKEN` | production tunnel, made by hand |
| HomeLab | `CLOUDFLARE_API_TOKEN` | Tunnel:Edit + DNS:Edit, nothing broader |
| HomeLab | `CLOUDFLARE_ACCOUNT_ID`, `CLOUDFLARE_ZONE_ID` | |
| this repo | `CMDCTRL_HOST`, `CMDCTRL_DEV_HOST` | Actions **variables**, not secrets; CD fails the run if the one for the branch is unset, or if the two are equal |
| this repo | `CMDCTRL_PUBLIC_BASE_URL`, `CMDCTRL_DEV_PUBLIC_BASE_URL` | variables; default to the two fqdns |
| this repo | `CMDCTRL_DISCORD_REDIRECT_URI`, `CMDCTRL_DEV_DISCORD_REDIRECT_URI` | variables; no default — unset means sign-in stays off on that host |
| this repo | `CMDCTRL_DISCORD_CLIENT_ID`, `CMDCTRL_DISCORD_CLIENT_SECRET`, `CMDCTRL_GITHUB_TOKEN` | secrets; shared by both hosts |
| this repo | `CMDCTRL_DISCORD_BOT_TOKEN` | secret; **production only**, written to `bot.env`. Unset skips the bot env sync |
| this repo | `CMDCTRL_DISCORD_APP_ID`, `CMDCTRL_DISCORD_GUILD_IDS` | variables; **production only**, written to `bot.env`. The app ID equals the sign-in client ID; guild IDs are comma-separated |

The preview's admin token is generated by Terraform and lands in
`/etc/cmd_and_ctrl/env` on the dev VM — the same path production uses
on its own VM. Read it with
`sudo grep CMDCTRL_ADMIN_TOKEN /etc/cmd_and_ctrl/env` on that host.

## Dev-only features

Enabled only when `CMDCTRL_ENV=dev`. Each may be individually turned
**off** on the preview to rehearse production behaviour; none can be
turned **on** in production.

| Feature | Variable |
|---|---|
| Raw protocol frame inspector | `CMDCTRL_DEV_FRAME_INSPECTOR` |
| Card spawn | `CMDCTRL_DEV_CARD_SPAWN` |
| Drive all seats from one browser | `CMDCTRL_DEV_SEAT_SWAP` |
| Replay scrubber | `CMDCTRL_DEV_REPLAY_SCRUBBER` |

`GET /config` reports the live answer:

```console
$ curl -s https://cmd-dev.labxp.io/config
{"env":"dev","features":{"card_spawn":true,"seat_swap":true,"frame_inspector":true,"replay_scrubber":true}}
$ curl -s https://cmd.labxp.io/config
{"env":"prod","features":{"card_spawn":false,"seat_swap":false,"frame_inspector":false,"replay_scrubber":false}}
```

### Card spawner

The **SPAWN** tab in the dev dock. Search by name, pick a seat and a
zone, spawn.

- Spawned cards go through `deck.ToGameCard`, the same conversion the
  deck importer uses, so they carry a real OracleID and the effect
  catalog matches them. A spawned Blood Artist actually watches for
  deaths.
- **Battlefield spawns fire ETB triggers**, so a spawn can put things
  on the stack. That is the point, but the resulting snapshot may
  differ by more than the cards you asked for.
- The game must be **active**; the stack is refused (spawn to hand and
  cast it); count is capped at 20.
- Spawns route through the room like every other mutation, so they land
  in the replay.

Life, counters, poison, energy and phase are **not** here — an admin
session can already set those on any seat through the ordinary action
protocol.

### Seat swap

The **SEAT** dropdown. Admin sessions only. It adds no server
capability: an admin could always bind to any seat with `?player=` on
the WebSocket URL. Two surprises:

- **Binding to a seat gives up the admin bypass.** Player-scoped gates
  now apply to you as that seat. Switch to *— spectate —* for the
  moderator override.
- **Switching reconnects.** A brief blank frame is expected.

### Replay scrubber

The **REPLAY** tab loads this game's recorded snapshots and steps the
board through them. No re-simulation — every frame is exactly what the
server recorded.

- **Frames are unfiltered**: a replay shows every hand.
- **Scrubbing hides the quick-action toolbar**, which would otherwise
  mutate the live game while you look at the past.
- "reload" picks up frames added since you loaded. A download taken
  mid-write can end in a half-written line; the parser drops it and
  keeps everything before.

Seeded shuffles were scoped alongside this and dropped — the spawner
reaches a chosen board state directly. Say so if you want them.

**Adding a dev feature.** Add the field to `appenv.Features`, register
its variable in `featureVars`, mirror the JSON key in
`client/src/lib/env.ts`, wrap every route in `requireDev` /
`requireDevFeature` **outside** `auth.Middleware`, and add the path to
**both** `client/vite.config.ts` and the `@api` matcher in
[`deploy/Caddyfile`](../deploy/Caddyfile). A route missing from either
fails silently: the client gets `index.html`, fails to parse it, and
hides the feature.

`deploy/Caddyfile` lives in this repo, not in the HomeLab template: the
CD job validates it and installs it on both hosts on every deploy, so
the route and its proxy entry belong in the same pull request. That is
the whole reason it moved — while it lived in cloud-init, nobody
editing routes ever opened it.

## Discord

Add `https://cmd-dev.labxp.io/auth/discord/callback` to the existing
application's redirect URIs. Same client ID and secret.

**Do not run the bot on dev.** Two processes logged into the same
application both answer every slash command. CD enforces this
(`if: env.IS_DEV != 'true'` on every bot step: binary, unit, `bot.env`
and restart) — do not work around it by installing the bot unit on the
preview by hand.

## Operating notes

**Which environment am I looking at?** The client shows a fixed `DEV`
badge in the bottom-right on the preview and nothing on production.
`curl -s <host>/config` is the authoritative answer.

**Logs.** `journalctl -u cmd-and-ctrl -f` on either host — same unit
name. The server logs a boot line naming its environment and, on dev, a
WARN listing the enabled features. If the journal and the browser badge
disagree, the browser is talking to a different backend than you think.

**Resetting preview state.** Preview games are disposable:

```sh
sudo systemctl stop cmd-and-ctrl
sudo find /var/lib/cmd_and_ctrl/data/games -mindepth 1 -delete
sudo systemctl start cmd-and-ctrl
```

Check `hostname` first. The command is identical on both boxes, which
is convenient right up until it isn't.

**A deploy failed at "Verify reported environment".** The service is
running but reports the wrong `CMDCTRL_ENV`. Check the env file on that
host. This failing on `main` is the serious direction — production came
up with dev features reachable — so treat it as an incident, not a
flake.
