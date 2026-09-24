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

That is why a push to `develop` runs CI again even though its PR
passed: it is the first run that tests the merged tree. It does not
hold up the preview, though. `ci-cd.yml` runs `build`, `server-check`,
`server-test` and `client` as parallel jobs, `cmdctrl-cd-dev` deploys
as soon as `build` is done, and the tests finish behind it. A push to
`main` waits for all of CI (`cmdctrl-cd-prod` needs `cmdctrl-ci`).
On a PR, `changes` skips the Go jobs when only `client/` or
`tests-e2e/` changed, and the client job when neither did.
`cmdctrl-ci` is a small job that passes when the others passed or
were skipped, so it stays the one required check.

The CI jobs (`changes` through `census-publish`) run on GitHub-hosted
`ubuntu-latest` runners (#1433), which are free for a public
repository. The self-hosted runners do only what needs the LAN: the
two CD jobs, the Scryfall refresh and the nightly workflows. Two
safeguards keep the hosted jobs free:

- If the repository is ever made private, the CI jobs go back to
  `self-hosted` automatically.
- Setting the repository variable `CI_RUNNER` to `self-hosted` moves
  them back on purpose, without a PR.

Keep `CI_RUNNER` to a standard label. Larger-runner labels are billed
even on public repositories.

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
and `/etc/cmd_and_ctrl/bot.env` come from the bot steps. On both hosts,
the off-site backup's units, `restic` itself (on a host built before
HomeLab#58) and `/etc/cmd_and_ctrl/backup.env` come from "Ensure
off-site backup" (see [Backups](#backups)). Change any of them here,
not in HomeLab.

The bot leaves one exception to "nothing to run by hand": creating the
`cmdctrl-bot` user and enabling the unit on production. The HomeLab
template does not create the user yet (the HomeLab half of #249), so
these steps are needed on the running production VM and **again after
every rebuild of production**, until the template does. A rebuilt host
without them still deploys green, with no bot; once the
`CMDCTRL_DISCORD_BOT_TOKEN` secret is set, CD flags it with a
`::warning::` pointing back at the steps. See
[deploy/README.md](../deploy/README.md#one-time-host-setup-repeat-after-every-rebuild).

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
| this repo | `CMDCTRL_DISCORD_BOT_TOKEN` | secret; **production only**, written to **both** `bot.env` (the gateway bot) and `/etc/cmd_and_ctrl/env` (the server's DM-invite route — see below). Unset skips the bot env sync with a notice; set turns a host that cannot run the bot into a CD warning (never a failure) |
| this repo | `CMDCTRL_DISCORD_APP_ID`, `CMDCTRL_DISCORD_GUILD_IDS` | variables; **production only**, written to `bot.env`. The app ID equals the sign-in client ID; guild IDs are comma-separated |
| this repo | `CMDCTRL_DISCORD_ADMIN_USER_IDS`, `CMDCTRL_DISCORD_ADMIN_ROLE_IDS` | variables; **production only**, written to `bot.env` only when set. Comma-separated snowflakes — who may run `/cc-end` (#614). Unset is supported: `/cc-end` refuses every caller rather than failing open |
| this repo, **environments `prod` and `dev`** | `CMDCTRL_R2_REPOSITORY` | environment variable, set in each; the restic repository URL for that host's bucket, `s3:https://<account_id>.r2.cloudflarestorage.com/<bucket>`. See [Backups](#backups) |
| this repo, **environments `prod` and `dev`** | `CMDCTRL_R2_ACCESS_KEY_ID`, `CMDCTRL_R2_SECRET_ACCESS_KEY` | environment secrets, set in each; an R2 API token scoped to that host's bucket. Written to `backup.env` |
| this repo, **environments `prod` and `dev`** | `CMDCTRL_RESTIC_PASSWORD` | environment secret, set in each; that host's restic repository key. **Never change it once the repository exists**, and keep a copy in the password manager: it is the only way to read a backup if GitHub's copy is lost, since secrets are write-only |
| this repo | `CENSUS_DEPLOY_KEY` | secret; the `census-publish` job's own push credential, not `GITHUB_TOKEN`. See [Census publish deploy key](#census-publish-deploy-key) below for setup |

The preview's admin token is generated by Terraform and lands in
`/etc/cmd_and_ctrl/env` on the dev VM — the same path production uses
on its own VM. Read it with
`sudo grep CMDCTRL_ADMIN_TOKEN /etc/cmd_and_ctrl/env` on that host.

`CMDCTRL_SESSION_KEY`, the session signing key (#517,
[ADR 0044](decisions/0044-surviving-a-deploy.md) decision 3), is in
neither table because it lives nowhere but the host. CD's "Ensure server
env (session signing and identity keys)" step generates it on the VM the first time it
finds it missing, and leaves it alone after that. Each host has its own,
so a preview session does not validate on production. Rotating it logs
every player out: delete the line from `/etc/cmd_and_ctrl/env` and
redeploy.

`CMDCTRL_IDENTITY_KEY`, the key that encrypts Discord refresh tokens in
the database (S34 sub-PR 2, [ADR 0051](decisions/0051-user-database.md)
decision 5), is provisioned the same way by the same CD step, "Ensure
server env (session signing and identity keys)": generated on the host
the first time it is missing, never rewritten, one per host. It must be
at least 32 bytes and differ from both the admin token and the session
key, or the server refuses to boot. Without it, Discord sign-in still
works and the user row is still written, but the refresh token is
dropped and the boot log warns. Rotating it (delete the line, redeploy)
makes the stored refresh tokens unreadable; nothing reads them yet, and
no one is logged out.

`CMDCTRL_DISCORD_BOT_TOKEN` is the one Discord value **both** units
need. The gateway bot has always read it from `/etc/cmd_and_ctrl/bot.env`
(`root:cmdctrl-bot`, mode `0640`); since S34 sub-PR 6 the *server*
reads it too, from its own `/etc/cmd_and_ctrl/env`, for ADR 0051
decision 5's direct-message invites (`POST /games/{id}/invites/dm`,
[#613](https://github.com/krakenhavoc/cmd_and_ctrl/issues/613)). Two
files rather than one shared file, because the two units run as
different users off different `EnvironmentFile=`s and the server cannot
read the bot's copy. Both are written from the same Actions secret on
the same deploy — "Sync server env (Discord DM invites)" and "Sync bot
env" — so rotating the secret in the Developer Portal and re-running
main updates both.

**Production only**, like every other bot value: there is one Discord
application, and a preview box DMing real people from the same bot
identity is the double-send the bot's prod-only rule already exists to
prevent. The develop deployment therefore has no token, which is a
supported state and not a failure: the DM route answers `503` naming
the variable, every other route is unchanged, and the boot log says
which state the process is in, once. It never falls open. The route
also needs an origin to build the invite link against
(`CMDCTRL_PUBLIC_BASE_URL`, falling back to `CMDCTRL_CLIENT_BASE_URL`);
missing that is the same 503, for the same reason bug-report
attachments have no default origin — a wrong one produces a DM full of
dead links.

### Census publish deploy key

`CENSUS_DEPLOY_KEY` is what the `census-publish` job in `ci-cd.yml`
pushes the regenerated catalog census with, back to `develop` and
`main` (Discussion #1231, review points 2 and 3). It is a repository
deploy key rather than `GITHUB_TOKEN` or a ruleset bypass for the
GitHub Actions app, on purpose: a deploy key can be scoped to exactly
this repository and named individually as a bypass actor, where an
`always` bypass for the Actions app — the shape #1247 shipped — covers
every push this workflow makes, not just the one docs-only commit that
actually needs to bypass anything.

This is a one-time **admin** setup (the account that pushes day to day
has push access but not admin on this repository, so these four steps
are Luke's):

1. Generate a dedicated ed25519 key pair. Do not reuse a personal or
   host key:
   ```bash
   ssh-keygen -t ed25519 -C "cmd_and_ctrl census-publish" \
     -f census-publish-deploy-key -N ""
   ```
2. Add the **public** half (`census-publish-deploy-key.pub`) as a
   repository deploy key with write access: repo Settings → Deploy
   keys → Add deploy key → paste the contents → check "Allow write
   access".
3. Add the **private** half (the whole `census-publish-deploy-key`
   file, `-----BEGIN OPENSSH PRIVATE KEY-----` line through
   `-----END...-----`) as the repository secret `CENSUS_DEPLOY_KEY`:
   Settings → Secrets and variables → Actions → New repository secret.
   Delete the local copy of both files afterwards — like the restic
   password above, GitHub secrets are write-only, so there is no way
   to read this back later; regenerate and replace it if it is lost.
4. On **both** the `develop` and `main` rulesets (Settings → Rules →
   Rulesets), add **Deploy keys** as a bypass actor, mode "Always".
   This names deploy keys as a class of actor, not this specific key,
   so it authorizes exactly the credential `census-publish` holds and
   nothing else in this workflow — the narrower scope Luke's review
   asked for in place of the app-wide bypass #1247 shipped with.

**Rotation:** repeat steps 1-3 with a new key pair, then remove the
old public key from Settings → Deploy keys. Step 4 does not need to
change — it names the actor type, not a specific key.

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

Each allowed guild authorizes the app with the `bot applications.commands`
scopes and no permissions (decided 2026-09-16, ADR 0004): the bot user has
to be a guild member to send DM invites (#613). A guild authorized before
then needs re-authorizing once. The install URL is in
[deploy/README.md](../deploy/README.md#discord-scopes).

## Operating notes

**Which environment am I looking at?** The client shows a fixed `DEV`
badge in the bottom-right on the preview and nothing on production.
`curl -s <host>/config` is the authoritative answer.

**Logs.** `journalctl -u cmd-and-ctrl -f` on either host — same unit
name. The server logs a boot line naming its environment and, on dev, a
WARN listing the enabled features. If the journal and the browser badge
disagree, the browser is talking to a different backend than you think.

**What a restart cost.** [ADR 0044](decisions/0044-surviving-a-deploy.md)
decision 1 and its 2026-09-24 amendment (#524): the server logs what a
deploy costs each live table on the way down and the way back up, so
"did that restart rewind anybody" is a `journalctl` search rather than
a question for the players.

At SIGTERM, after the HTTP listener closes and before the hub drops any
connection, one line per live game:

```
level=INFO msg="shutdown census: game" game_id=d3b53ef1-… archived=false clean=true seq=1 has_restore_point=true restore_seq=1 seq_behind=0 restore_age=0s
level=WARN msg="shutdown census: game" game_id=761b8df2-… archived=true clean=false seq=2 has_restore_point=true restore_seq=1 seq_behind=1 restore_age=0s continuations=1 continuation_labels="[stack effect: named continuation for the log test]"
level=INFO msg="shutdown census: summary" games=2 clean=1 would_rewind=1 actions_rewound=1
```

`clean=true` means the table's CURRENT state is itself a usable restore
point — a restart right now would not rewind it at all. `clean=false`
names what's blocking it (`continuation_labels`, the same census the
catalog census tooling uses) and `seq_behind` is how many actions have
happened since the last state a restart CAN rebuild exactly. `WARN`
level is what makes `journalctl -u cmd-and-ctrl -p warning` after a
deploy answer "which tables rewound" directly. `archived=true` is a
retired table (`GET /games?archived=1`) — expected to show up here,
since archiving keeps the room registered.

At boot, `RestoreFromDisk` is timed (the number [ADR 0044](decisions/0044-surviving-a-deploy.md)'s
measurement task asks for, to confirm or refute that startup rather
than this pass dominates the restart window) and each restored game's
line grows a `restore_point_age`:

```
level=ERROR msg="restore point written by a newer server; game abandoned (file kept for roll-forward)" game_id=87516ae4-… path=/var/lib/cmd_and_ctrl/data/restore/87516ae4-….json err="game: snapshot schema is newer than this server understands: file is v7, this server reads up to v6"
level=INFO msg="game restored" game_id=b151d7c9-… seq=1 state=active turn=1 seats=2 captured_at=2026-09-24T16:14:11.919Z restore_point_age=0s
level=INFO msg="restore pass complete" restored=1 ended=0 abandoned=1 abandoned_reasons=map[schema_too_new:1]
level=INFO msg="restore from disk complete" duration=118ms resumed=1
```

`restore_point_age` is how stale that table's restore point was — how
long it had been since a clean boundary, at the moment the PREVIOUS
process wrote it. It is deliberately not paired with "how far this
table rewound": that number is the restored seq against the seq the
previous process was showing to clients right before it died, and the
only record of that is the previous process's own shutdown census line
above — this process has no reliable way to read that back. The
shutdown line already carries the true rewind delta (`seq_behind`) for
the process that is about to stop.

**Resetting preview state.** Preview games are disposable:

```sh
sudo systemctl stop cmd-and-ctrl
sudo find /var/lib/cmd_and_ctrl/data/games -mindepth 1 -delete
sudo systemctl start cmd-and-ctrl
```

Check `hostname` first. The command is identical on both boxes, which
is convenient right up until it isn't.

**Rolling back past the lobby database.** Since S34 sub-PR 3
([ADR 0051](decisions/0051-user-database.md) decision 4) the lobby's
half of a game is rows in `db/cmdctrl.sqlite`. The first boot of that
binary imported every `lobby/<id>.json` and renamed it
`<id>.json.imported`. It deleted none of them. An older binary reads
only `lobby/*.json`, so rename them back before you start it:

```sh
sudo systemctl stop cmd-and-ctrl
sudo find /var/lib/cmd_and_ctrl/data/lobby -name '*.json.imported' -exec sh -c 'mv -n "$1" "${1%.imported}"' _ {} \;
# install the older binary, then:
sudo systemctl start cmd-and-ctrl
```

`mv -n` never overwrites a `.json` that is already there. Two things
the rollback cannot bring back. A game **created** after the upgrade has
no file, so the older binary drops it. The invite tokens in each file
are the originals, and every old link works again. When you roll
forward again, the importer renames the files once more. For a game
that already has a row, the row wins, so a change the older binary made
to that game's seats in the meantime is not carried forward. A game
created during the rollback is imported as new.

**Session lifetimes.** A Discord sign-in from the login page mints an
identity session that lasts `CMDCTRL_IDENTITY_TTL` (default `720h`, 30
days; [ADR 0051](decisions/0051-user-database.md) decision 3). Seat,
spectator and admin sessions last `CMDCTRL_SESSION_TTL` (default `12h`).
Neither is set in `/etc/cmd_and_ctrl/env` by CD. Add the line by hand
to override one, then restart the service.

**Signing someone out.** A session that belongs to a user can be
withdrawn (ADR 0051 decision 6). That is every Discord sign-in, and every
seat claimed from one. The player can do it themselves with **log out
everywhere** in the lobby, or **sign out everywhere** on the login page.
An admin can do it for anyone:

```sh
curl -s -X POST -H "Authorization: Bearer $ADMIN_SESSION" \
  https://cmd.labxp.io/admin/users/<user-id>/revoke-sessions
```

Take `<user-id>` from `sqlite3 db/cmdctrl.sqlite 'SELECT id, display_name FROM users'`.
The command is revocation only. Every session that user holds stops
validating at once, their open game sockets close, and no rows are
deleted. They can sign in with Discord again straight away. It survives
a restart, because it is `users.sessions_invalid_before`. Admin, guest
and spectator sessions have no user and cannot be revoked this way.
The only way to end one early is rotating `CMDCTRL_SESSION_KEY`, which
logs everyone out. A server with no database (`CMDCTRL_DATA_DIR` empty)
has no users, so both routes answer 503 or 403 there.

**A deploy failed at "Verify reported environment".** The service is
running but reports the wrong `CMDCTRL_ENV`. Check the env file on that
host. This failing on `main` is the serious direction — production came
up with dev features reachable — so treat it as an incident, not a
flake.

## Backups

Each host backs its data dir up every night to its own Cloudflare R2
bucket with [restic](https://restic.net/) (#1031,
[ADR 0051](decisions/0051-user-database.md) decision 1). HomeLab has a
single Proxmox node and nothing off it, so the off-node copy leaves the
lab entirely. HomeLab owns only the buckets
([krakenhavoc/HomeLab#58](https://github.com/krakenhavoc/HomeLab/issues/58)).
The job, its credentials and this runbook live here, and CD ships them
to both hosts on every deploy.

| Piece | Where |
|---|---|
| Script | [`scripts/backup-offsite.sh`](../scripts/backup-offsite.sh), deployed to `/opt/cmd_and_ctrl/scripts/` |
| Units | [`deploy/cmd-and-ctrl-backup.service`](../deploy/cmd-and-ctrl-backup.service) and [`.timer`](../deploy/cmd-and-ctrl-backup.timer), installed by the CD step "Ensure off-site backup" |
| Credentials | `/etc/cmd_and_ctrl/backup.env`, `root:cmdctrl 0640`, rewritten by CD on every deploy from the Actions secrets below. It is a separate file from the server's `env`, so the game server never loads R2 credentials |
| restic cache | `/var/cache/cmd-and-ctrl-backup` (the unit's `CacheDirectory=`) |

The unit runs as `cmdctrl` with the data dir mounted read-only. It has
network access and little else: `ProtectSystem=strict`, a system-call
filter and no capabilities. It cannot read the server's `env` or
`bot.env`.

### What is backed up

Relative to `/var/lib/cmd_and_ctrl/data`, whichever of these exist:

| Path | What |
|---|---|
| `db/cmdctrl.backup.sqlite` | the user database, as the server's hourly `VACUUM INTO` copy (`CMDCTRL_DB_BACKUP_INTERVAL`). **Not** the live `cmdctrl.sqlite` or its `-wal`/`-shm`: a file copy of a live WAL database is not a consistent database |
| `restore/`, `games/` | in-flight game state and snapshots |
| `lobby/` | game metadata and invite tokens |
| `replays/` | replay logs |
| `bugreports/` | reporter screenshots and pinned replays |

**Not backed up:**

- `scryfall/`, `images/` and `avatars/`. These are caches that refill
  from Scryfall and Discord, and they are excluded explicitly.
- `/etc/cmd_and_ctrl/env`. HomeLab's Terraform writes it and CD fills
  in the rest. A rebuilt host generates a new session key, which logs
  every player out once.
- The binaries and the client, which CD redeploys.

The worst case loses about 25 hours of data: the database copy can be
up to an hour old when the nightly snapshot takes it.

### Schedule and retention

- **When:** nightly at 03:30 host time, plus up to 30 minutes of random
  delay. The timer has `Persistent=true`, so a night missed while the VM
  was down runs at the next boot.
- **Each run:** initialise the repository if there is none yet, back
  up, then run `restic forget --prune --keep-daily 7 --keep-weekly 4
  --keep-monthly 6`.
- **Retention is restic's.** Never add an R2 lifecycle rule to these
  buckets: deleting restic's pack files from under it corrupts the
  repository.
- **Weekly check:** on Sundays, `restic check --read-data-subset=5%`
  downloads and verifies a random 5% of the data.
- **Failures:** any failure fails the unit. Each successful run ends
  with one line in the journal:
  `backup-offsite: env=prod snapshot=1a2b3c4d added_bytes=… paths=6 check=no`.

Cost: $0 at this size. The data is 1–2 GB compressed. R2's free tier
covers 10 GB-month of storage and 1M Class A operations a month, and
egress is free.

### Secrets and variables

These live in the repo's **GitHub environments**, `prod` and `dev`,
under the same four names in each. The deploy job runs in `prod` for a
push to `main` and in `dev` for a push to `develop`. Each environment's
deployment branch policy admits only that branch, so no other branch can
read production's values. (The other per-host values in
[Secrets](#secrets) are still repo-level with a `CMDCTRL_DEV_` twin;
they can move into the environments later, one at a time.)

| Kind | Name (in both `prod` and `dev`) | Value |
|---|---|---|
| variable | `CMDCTRL_R2_REPOSITORY` | `s3:https://<account_id>.r2.cloudflarestorage.com/cmd-and-ctrl-backup-<prod\|dev>` |
| secret | `CMDCTRL_R2_ACCESS_KEY_ID` | the Access Key ID of an R2 API token scoped to that bucket |
| secret | `CMDCTRL_R2_SECRET_ACCESS_KEY` | that token's Secret Access Key |
| secret | `CMDCTRL_RESTIC_PASSWORD` | the restic repository key |

All eight were created with placeholder values containing `REPLACE_ME`
(2026-09-19). A value that still contains `REPLACE_ME` counts as unset.

If any of the four is unset or a placeholder for the host being
deployed, CD emits a `::warning::` naming each one, then stops and
disables the timer. The deploy stays green. With all four set, CD installs `restic`
if the host lacks it, writes `backup.env` and enables the timer. It also
warns if the unit's last run failed.

> **The restic password is the only key to the backups.**
>
> - **Store it off-host.** Generate it once, store it in the password
>   manager **first**, then set the secret. GitHub secrets are
>   write-only. If the password manager copy is lost, nobody can ever
>   read the backups again.
> - **Never change it** once the repository exists. CD writes the secret
>   into `backup.env` on every deploy, and restic cannot open a
>   repository with a different password. Every backup after a change
>   fails with `wrong password or no key found`. If that happens,
>   restore the old secret value from the password manager and
>   redeploy.
> - **Keep it to one line with no `'`.** CD refuses a password with a
>   single quote or a newline. `openssl rand -base64 36` produces a
>   safe one.

### One-time setup

1. **Apply HomeLab#58.** R2 must be enabled on the account, and the
   prod and dev buckets must exist.
2. **Create the R2 API tokens.** In the Cloudflare dashboard, go to
   R2, then Manage R2 API Tokens.
   - Create two tokens with *Object Read & Write*, each restricted to
     **one** bucket.
   - Note each token's Access Key ID and Secret Access Key. The
     dashboard shows the secret only once.
   - Optionally, create a third token with *Object Read only* on the
     prod bucket, for restore rehearsals (below).
3. **Generate the passwords.** Generate two restic passwords, one per
   host, and save both in the password manager.
4. **Set the values.** The account ID is on the R2 overview page.

   ```sh
   for e in prod dev; do
     gh variable set CMDCTRL_R2_REPOSITORY --env "$e" \
       --body "s3:https://<account_id>.r2.cloudflarestorage.com/cmd-and-ctrl-backup-$e"
     gh secret set CMDCTRL_R2_ACCESS_KEY_ID     --env "$e"   # each one prompts for the value
     gh secret set CMDCTRL_R2_SECRET_ACCESS_KEY --env "$e"
     gh secret set CMDCTRL_RESTIC_PASSWORD      --env "$e"
   done
   ```

   Or edit them under Settings → Environments → `prod` / `dev`.

5. **Deploy.** Push to `develop`, then to `main`. Use a new push, not a
   rerun: a rerun of an older run uses that run's snapshot of the
   variables (see
   [deploy/README.md](../deploy/README.md#one-time-host-setup-repeat-after-every-rebuild),
   step 3). The "Ensure off-site backup" step should print
   `wrote /etc/cmd_and_ctrl/backup.env` and the timer's next run, with
   no warning.
6. **Take the first snapshot now**, rather than waiting for 03:30. On
   each host, run `sudo systemctl start cmd-and-ctrl-backup`. It blocks
   until it finishes, and it initialises the repository. Then check it
   as below.

### Checking on it

```sh
systemctl list-timers cmd-and-ctrl-backup.timer         # next and last run
systemctl status cmd-and-ctrl-backup --no-pager          # last result
journalctl -u cmd-and-ctrl-backup -n 50 --no-pager       # ends with the backup-offsite: summary line
```

To run restic by hand, run it the way the unit does: as `cmdctrl`,
with `backup.env`, which systemd reads as root.

```sh
sudo systemd-run --quiet --pipe --wait --collect \
  -p User=cmdctrl -p EnvironmentFile=/etc/cmd_and_ctrl/backup.env \
  -p CacheDirectory=cmd-and-ctrl-backup -E RESTIC_CACHE_DIR=/var/cache/cmd-and-ctrl-backup \
  restic snapshots
```

In place of `restic snapshots` you can run:

- `restic stats latest`;
- `restic check --read-data-subset=5%`;
- `/opt/cmd_and_ctrl/scripts/backup-offsite.sh --check`, which runs a
  full backup with the check on any day of the week.

### Restore runbook: production's latest snapshot onto the dev host

Rehearse this at least once, which is the #1031 exit criterion, and
record the date on the issue. A real disaster restore onto a rebuilt
production host is the same procedure with production's own
`backup.env`, and without the warning below.

> **Warning:** afterwards the preview holds production's data: every
> account, deck and game, and **live invite tokens**. A production
> invite link works on `cmd-dev.labxp.io` once its host is swapped.
> Treat the dev host as production-sensitive until you reset it
> (step 7), and reset it soon. Sessions do not carry over, because each
> host has its own session key.

Run everything below **on the dev host**. Check `hostname` first.

1. **Stop the server and the dev backup timer.** If the timer keeps
   running, the next night backs production's data up into the dev
   bucket. The next `develop` deploy re-enables the timer, so do not
   merge to `develop` during the rehearsal.

   ```sh
   hostname                                    # cmd-and-ctrl-dev
   sudo systemctl stop cmd-and-ctrl cmd-and-ctrl-backup.timer
   ```

2. **Restore into a staging directory.** Dev's `backup.env` points at
   the dev bucket, so write production's four values into a root-only
   file. Use the read-only prod token if you made one, and take the
   prod password **from the password manager**.

   ```sh
   sudo install -m 0600 /dev/null /root/prod-restore.env
   sudo nano /root/prod-restore.env
   #   RESTIC_REPOSITORY='s3:https://<account_id>.r2.cloudflarestorage.com/<prod-bucket>'
   #   AWS_ACCESS_KEY_ID='…'
   #   AWS_SECRET_ACCESS_KEY='…'
   #   AWS_DEFAULT_REGION='auto'
   #   RESTIC_PASSWORD='…'
   sudo bash -c 'set -a; . /root/prod-restore.env; set +a
     restic --no-lock snapshots
     restic --no-lock restore latest --tag prod --target /var/tmp/cmdctrl-restore'
   ```

   `--no-lock` lets a read-only token read the repository. The staging
   tree mirrors the data dir: `/var/tmp/cmdctrl-restore/db/cmdctrl.backup.sqlite`,
   `/var/tmp/cmdctrl-restore/lobby/`, and so on.

3. **Put it in place.** The backup copy becomes the live database. The
   stale `-wal` and `-shm` files must go first: SQLite would replay a
   WAL left over from dev's old database into production's file.

   ```sh
   D=/var/lib/cmd_and_ctrl/data S=/var/tmp/cmdctrl-restore
   for d in restore replays lobby bugreports games; do
     if [ -d "$S/$d" ]; then sudo rm -rf "$D/$d"; sudo cp -a "$S/$d" "$D/$d"; fi
   done
   sudo install -d -o cmdctrl -g cmdctrl -m 0700 "$D/db"
   sudo rm -f "$D/db/cmdctrl.sqlite-wal" "$D/db/cmdctrl.sqlite-shm"
   sudo cp "$S/db/cmdctrl.backup.sqlite" "$D/db/cmdctrl.sqlite"
   ```

4. **Fix ownership and modes.** restic restores production's numeric
   owners, which need not match dev's `cmdctrl` user.

   ```sh
   sudo chown -R cmdctrl:cmdctrl "$D/db" "$D/restore" "$D/replays" "$D/lobby" "$D/bugreports" "$D/games"
   sudo chmod 0700 "$D/db"
   sudo chmod 0600 "$D/db/cmdctrl.sqlite"
   ```

5. **Start and verify.**

   ```sh
   sudo systemctl start cmd-and-ctrl
   journalctl -u cmd-and-ctrl -n 50 --no-pager        # "lobby restored from disk", and no errors
   curl -s http://127.0.0.1:8080/config               # still "env":"dev"
   ```

   Then sign in at `https://cmd-dev.labxp.io`. Production's games should
   be in the lobby, and production's accounts should exist.

6. **Clean up the credentials and the staging copy.**

   ```sh
   sudo shred -u /root/prod-restore.env
   sudo rm -rf /var/tmp/cmdctrl-restore
   ```

7. **Reset dev** when the rehearsal is over. This is the
   [resetting preview state](#operating-notes) recipe, widened to
   everything the restore brought in, including the database:

   ```sh
   sudo systemctl stop cmd-and-ctrl
   for d in games lobby restore replays bugreports; do
     sudo find "/var/lib/cmd_and_ctrl/data/$d" -mindepth 1 -delete
   done
   sudo rm -f /var/lib/cmd_and_ctrl/data/db/cmdctrl.sqlite* /var/lib/cmd_and_ctrl/data/db/cmdctrl.backup.sqlite
   sudo systemctl start cmd-and-ctrl cmd-and-ctrl-backup.timer
   ```

   The server creates an empty database when it boots.
