# ADR 0004 — Discord identity and slash-command bot (S12.5)

**Status:** Accepted · 2026-04-22 · Sprint S12.5

## Context

The playgroup already coordinates on Discord — in-game chat was
removed in S08.5 on the grounds that Discord is where
conversation happens anyway. S12.5 leans into that: Discord
supplies display names, avatars, and an entry point into the
game without the web client having to own identity.

Two surfaces ship under S12.5 and are covered here together:

1. **OAuth2 sign-in on the invite-link landing page.** A friend
   who clicks an invite shared in Discord signs in once with
   Discord and shows up in the game with their Discord name and
   avatar already on the seat. (Implemented in PR #127–129.)
2. **Slash-command bot for posting invites from inside
   Discord.** Admins type `/cc-invite` in an approved server and
   get a game + an invite URL pasted back to the channel. This
   is the PR 1–4 slice that closes the S12.5 bot tasks.

Rich Presence, DM invites (`/cc-invite-dm`), and
re-link-after-the-fact are **deferred**. They're out-of-scope
for the MVP; the user's ask was slash commands + approved-server
gating, and those two ship here.

## Decisions

### 1. OAuth via stdlib `net/http`; bot via `bwmarrin/discordgo`

The OAuth half ([server/internal/discord/oauth.go](../../server/internal/discord/oauth.go))
is ~3 HTTP endpoints (authorize redirect, token exchange,
`/users/@me` fetch) and stays on the Go standard library.
That matches the project's stdlib-first convention and keeps
the surface we audit small.

The bot half ([server/cmd/bot/](../../server/cmd/bot/)) connects
over Discord's gateway (WebSocket), which means heartbeats,
resume, ratelimits, and slash-command registration. Rolling
that stack against stdlib would be ~500 LoC we'd own; we pick
[bwmarrin/discordgo](https://github.com/bwmarrin/discordgo) v0.29
as the first substantial third-party dep in the server module.
Deviation from stdlib-first is intentional and documented here
so future readers of `server/internal/discord/config.go:7-12`
understand why.

**Not chosen:** interaction webhooks. Webhooks require a public
HTTPS endpoint, Ed25519 signature verification on every
request, and a separate public-route surface. The gateway dials
out from the VPS; nothing new has to be reachable from the
internet.

### 2. Bot is a separate binary under the same Go module

[server/cmd/bot/main.go](../../server/cmd/bot/main.go) builds
as `cmd_and_ctrl-bot` alongside the existing
`cmd_and_ctrl-server`. Shared module = shared types
(`lobby.GameMeta` reused directly in the bot's HTTP client);
separate binary = crash isolation + independent restart cadence
+ a Linux-user boundary between the bot (holds admin token) and
the server (holds everything else).

**Not chosen:** in-process goroutine inside `cmd_and_ctrl-server`.
A panic in discordgo's reconnect loop would evict live WS
clients. Not worth it for the deploy-simplicity savings.

### 3. Guild allow-list via `CMDCTRL_DISCORD_GUILD_IDS`, enforced twice

Comma-separated snowflakes in a single env var. Enforced at:

1. **Registration time.** `RegisterCommands` iterates the
   allow-list and registers the two commands per allowed guild.
   A `GuildCreate` event on a non-listed guild logs a warning and
   skips registration; we do NOT auto-leave the guild because
   an operator may be mid-setup and have not yet added the new
   guild ID to the env file.
2. **Per-interaction.** The dispatcher re-checks `GuildAllowed`
   before touching the HTTP client and rejects ephemerally if
   the guild ID isn't on the list. Defense-in-depth for stale
   registrations and for the race where a bot gets added to a
   new guild before the operator restarts.

**Not chosen:** a runtime `/cc-approve <guild_id>` admin command
(more surface area, requires persistence), or a JSON config
file (no upside over env vars on a systemd deploy).

### 4. Guild-scoped slash commands, not global

Guild-scoped commands propagate **instantly**; global commands
take up to an hour. For a private bot with a small,
known-in-advance guild list, guild-scope wins on the dev-loop
and also means the commands don't exist at all in non-allowed
servers. A future switch to global would be a deliberate choice
— flag noted here so anyone making that swap understands the
propagation trade-off.

### 5. Bot → server goes over loopback HTTP, not an imported lobby

The bot calls `POST /admin/login` + `POST /games` + `GET /games`
through [server/internal/bot/client.go](../../server/internal/bot/client.go),
a thin `net/http` wrapper. Preserves the lobby's single-writer
property (the bot is just another admin client) and avoids an
import of `lobby` that would drag the ws + game packages into
the bot binary.

The 5 s per-request timeout is safely below Discord's 3 s
interaction-response deadline when accounting for one login +
one work call. A future cold-start regression that blows past
3 s would force us to switch to
`InteractionResponseDeferredChannelMessageWithSource` and edit
the response after. Not needed today.

### 6. Bot token storage: separate `EnvironmentFile`, not co-located with admin token

The bot holds both its Discord bot token and the server's
`CMDCTRL_ADMIN_TOKEN` — co-locating them in one env file
collapses two independent blast radii into one. The deployed
shape is:

- `/etc/cmd_and_ctrl/server.env` owned `krkn:krkn` mode `0640` —
  game-server secrets only.
- `/etc/cmd_and_ctrl/bot.env` owned `cmdctrl-bot:cmdctrl-bot`
  mode `0640` — bot token + the admin-token copy.

A `systemd-creds`-based upgrade is cleanly possible later and
should be considered on the next incident rotation.

## Consequences

- **Admin-token reuse** — the bot has the same authority as a
  human admin. Bot compromise ≈ admin compromise. On any
  incident, rotate BOTH the bot token AND `CMDCTRL_ADMIN_TOKEN`.
- **Allow-list changes need a bot restart** — there's no SIGHUP
  reload today. Acceptable for a hobby deploy; flagged for
  follow-up if the guild list starts churning.
- **First substantial third-party Go dep in the server module.**
  Pins to `bwmarrin/discordgo v0.29`; minor-version bumps are
  safe, major-version bumps need a deliberate upgrade PR.

## Operator runbook

> **Production provisioning has moved to CD.** "First deploy" and
> "Rotating tokens" below describe the original hand-installed setup
> and are out of date, except step 2's guild scope, which still
> stands. CD now installs the unit and writes
> `/etc/cmd_and_ctrl/bot.env` (`root:cmdctrl-bot 0640`) from Actions
> secrets and variables, so do not hand-edit it, and rotate the bot
> token by updating the Actions secret. The current host steps are in
> [deploy/README.md](../../deploy/README.md#discord-bot-production-only).
> Correcting §6 and this runbook is tracked in #615 and #251.

### Env vars

The game server reads the existing OAuth trio (see S12.5 OAuth
PRs for those). The bot binary adds:

| Var | Required | Default | Notes |
| --- | --- | --- | --- |
| `CMDCTRL_DISCORD_BOT_TOKEN` | yes | — | Developer Portal → Bot tab → Reset Token |
| `CMDCTRL_DISCORD_APP_ID` | yes | — | Developer Portal → General Info → Application ID |
| `CMDCTRL_DISCORD_GUILD_IDS` | yes | — | Comma-separated snowflakes, e.g. `111222333,444555666` |
| `CMDCTRL_ADMIN_TOKEN` | yes | — | Same shared secret the server uses |
| `CMDCTRL_SERVER_BASE_URL` | no | `http://127.0.0.1:8080` | Where the bot calls the admin API |
| `CMDCTRL_CLIENT_BASE_URL` | no | `https://cmd.labxp.io` | Used to compose the invite URL posted to Discord |

Empty `CMDCTRL_DISCORD_BOT_TOKEN` disables the bot — the
binary logs `bot disabled` and exits 0. Convenient for a dev
stack that doesn't have a registered Discord app.

### First deploy

1. Create the Discord application in the Developer Portal; enable
   the Bot user; copy the bot token.
2. Invite the bot to each approved guild with the
   `applications.commands` scope (and nothing else — the bot
   does not need message-content, member-list, or voice
   permissions).
3. Place `/etc/cmd_and_ctrl/bot.env` on the VPS:
   ```
   CMDCTRL_DISCORD_BOT_TOKEN=...
   CMDCTRL_DISCORD_APP_ID=...
   CMDCTRL_DISCORD_GUILD_IDS=111222333,444555666
   CMDCTRL_ADMIN_TOKEN=... # same value as /etc/cmd_and_ctrl/server.env
   ```
4. Install the systemd unit at [deploy/cmd-and-ctrl-bot.service](../../deploy/cmd-and-ctrl-bot.service):
   ```
   sudo cp deploy/cmd-and-ctrl-bot.service /etc/systemd/system/
   sudo systemctl daemon-reload
   sudo systemctl enable --now cmd-and-ctrl-bot
   ```
5. CI-driven deploys restart the bot after every successful
   server deploy (see CI changes in this sub-PR); the restart
   is conditional on the unit being enabled so pre-rollout
   CI runs don't fail.

### Rotating tokens

Both the bot token and the admin token are high-value. On
suspected compromise:

1. Reset both in the Discord Developer Portal / server env.
2. Update `/etc/cmd_and_ctrl/bot.env` and
   `/etc/cmd_and_ctrl/server.env`.
3. `sudo systemctl restart cmd-and-ctrl cmd-and-ctrl-bot`.
4. Invalidate any admin sessions minted before the rotation if
   the threat model requires it (currently no revocation UI —
   the in-memory authenticator's entries TTL out in 12 h).
