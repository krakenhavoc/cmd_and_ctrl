# deploy/

| File | Installed on | Installed by |
|---|---|---|
| `Caddyfile` | both hosts | CD, every deploy of `main` and `develop` ("Deploy Caddy config") |
| `cmd-and-ctrl-bot.service` | production only | CD, every deploy of `main` ("Install bot systemd unit"); **enabling** it is an operator step, repeated after every rebuild, below |
| `cmd-and-ctrl-backup.service`, `cmd-and-ctrl-backup.timer` | both hosts | CD, every deploy of `main` and `develop` ("Ensure off-site backup"), which also enables the timer, or disables it and warns when a backup secret or variable is unset. Nightly restic backup to R2; runbook in [docs/environments.md](../docs/environments.md#backups) |

## The game server's unit is not here

Both hosts get `cmd-and-ctrl.service`, a first-boot placeholder
Caddyfile and `/etc/cmd_and_ctrl/env` from the cloud-init template in
[krakenhavoc/HomeLab](https://github.com/krakenhavoc/HomeLab):
`terraform/deployments/lab/templates/setup-cmd_and_ctrl.yaml.tftpl`.
CD then replaces the Caddyfile with the one in this directory and keeps
the env file's CD-managed keys in sync.

That template is the single source of truth for both production and the
develop preview. To change host configuration, change it there and
apply — do not edit files on the box. A rebuild reverts them, which is
how production ended up serving Discord OAuth from a Caddyfile the
template did not contain.

See [docs/environments.md](../docs/environments.md).

## Discord bot (production only)

**Never on the preview host.** Two bot processes logged into the same
Discord application both answer every slash command. Every bot step in
CD is gated `if: env.IS_DEV != 'true'`; do not work around that by hand.

### What CD does on every `main` deploy

1. Creates `/opt/cmd_and_ctrl/bot/bin` (krkn owns `/opt/cmd_and_ctrl`)
   and rsyncs `cmd_and_ctrl-bot` into it. A failure fails the deploy.
2. Installs `cmd-and-ctrl-bot.service` to `/etc/systemd/system`,
   `root:root 0644`, and runs `daemon-reload`, only when it changed.
   It never enables the unit.
3. **Sync bot env:** writes `/etc/cmd_and_ctrl/bot.env`
   (`root:cmdctrl-bot 0640`, created and re-asserted by CD) through
   `scripts/set-server-env.sh`, values on stdin:

   | Key | Source |
   |---|---|
   | `CMDCTRL_DISCORD_BOT_TOKEN` | Actions **secret** of the same name |
   | `CMDCTRL_DISCORD_APP_ID` | Actions **variable** of the same name |
   | `CMDCTRL_DISCORD_GUILD_IDS` | Actions **variable** of the same name, comma-separated guild IDs |
   | `CMDCTRL_ADMIN_TOKEN` | copied on the host from `/etc/cmd_and_ctrl/env` every deploy, so it cannot drift from the server's |

   `CMDCTRL_SERVER_BASE_URL` and `CMDCTRL_CLIENT_BASE_URL` are left to
   their defaults, which are production's values. The step skips while
   the secret is unset, or while the `cmdctrl-bot` group does not exist.
4. Restarts the bot, **only if the unit is enabled**, then checks it is
   still active 4 seconds later.

Rotating the bot token: update the secret, then start a new `main`
deploy (see step 3 below for why a new run rather than a rerun).

### What CD reports

The bot's state never fails a deploy; only a failed copy or ssh does.
How loud CD is depends on whether the `CMDCTRL_DISCORD_BOT_TOKEN` secret
is set. CD computes that as a `true`/`false` flag (`BOT_TOKEN_SET`) and
never prints the token.

| Host state | Secret unset | Secret set |
|---|---|---|
| no `cmdctrl-bot` group ("Sync bot env"; `bot.env` not written) | `::notice::` (the step skips before looking) | `::warning::` |
| group but no `cmdctrl-bot` user ("Sync bot env"; `bot.env` still written) | `::notice::` (the step skips before looking) | `::warning::` |
| unit not enabled (restart step; bot not restarted) | `::notice::` | `::warning::` |
| unit not active 4 s after its restart (restart step) | `::notice::` | `::warning::` |

Every warning is titled "Discord bot not provisioned" or "Discord bot
not running" and points at the one-time host setup below. So once the
secret is set, a green deploy with no warning annotations means the bot
was installed, configured, restarted and survived its first 4 seconds.
It does not prove the bot logged in; see [Verify](#verify).

### One-time host setup (repeat after every rebuild)

The HomeLab cloud-init template does not create the `cmdctrl-bot` user
yet (the HomeLab half of #249). Until it does, these steps are needed on
the running production VM **and again after every rebuild of
production**. A rebuilt VM comes up with no `cmdctrl-bot` user and no
enabled unit, which is how the bot was lost in the August reprovision.

CD warns but stays green. With the secret set, such a host gets a
`::warning::` from "Sync bot env" (no `cmdctrl-bot` group) and another
from the restart step (unit not enabled), and the deploy still passes
with no bot. Look for them on the first `main` deploy after any rebuild,
or check `systemctl is-enabled cmd-and-ctrl-bot` and `id cmdctrl-bot`
by hand.

This is not a permission limit: CD runs as `krkn` with sudo. The user
belongs in the template, which is where a rebuild should get it from.
Enabling the unit is kept as a deliberate operator switch, and the CD
restart step keys on it.

1. Set the secret and the two variables (above) in the repo's Actions
   settings. After a rebuild they are normally already set.
2. On the production host, create the service user. CD needs its group
   before it will write `bot.env`:

   ```sh
   sudo useradd --system --user-group --no-create-home \
     --home-dir /nonexistent --shell /usr/sbin/nologin cmdctrl-bot
   ```

3. Get a `main` deploy that includes these CD steps to run to
   completion. CD runs only on a push to `main` (or `develop`, which
   never touches the bot). A manual "Run workflow" dispatch never reaches
   the deploy job. Re-running an existing run reuses that run's `vars`
   snapshot (#598 §1), so if you set or changed a variable in step 1, a
   rerun deploys without it. In that case the deploy has to come from a
   new push to `main`, normally the next `develop` → `main` promotion.

   In the log, "Install bot systemd unit" prints
   `cmd-and-ctrl-bot.service installed; daemon reloaded.` only on the
   deploy that first installs or changes the unit, and
   `cmd-and-ctrl-bot.service unchanged.` after that. Either is fine.
   "Sync bot env" should print `set CMDCTRL_...` for all four keys, with
   no annotation about a missing group, user or secret. The restart step
   still warns that the unit is not enabled; step 4 fixes that.
4. Enable and start the bot:

   ```sh
   sudo systemctl enable --now cmd-and-ctrl-bot
   ```

   From then on every `main` deploy restarts it.
5. Authorize the app in each allowed guild with the install URL under
   [Discord scopes](#discord-scopes). This is per guild, not per host, so
   a rebuild does not need it again. A guild that was authorized before
   2026-09-16 with `applications.commands` alone needs it **once more**,
   with the new URL.

### Verify

`systemctl is-active` is not proof. The unit is `Type=simple`, so it
reports active the moment the process forks, even when the bot exits a
second later and systemd restarts it every 5 seconds. CD's 4-second
check catches a bot that dies at once, but not one that fails later.
Check:

```sh
journalctl -u cmd-and-ctrl-bot -n 50 --no-pager   # "bot config loaded", then "discord session ready"
systemctl show -p NRestarts cmd-and-ctrl-bot      # stays 0 over a minute
sudo stat -c '%U:%G %a' /etc/cmd_and_ctrl/bot.env # root:cmdctrl-bot 640
```

`bot config invalid` names the missing key; `bot disabled` means the
token is empty. Then in an allowed guild, `/cc-games` should answer with
an ephemeral list. A reply saying the bot is not authorized against the
game server means the two admin tokens differ. The bot user should also
appear in each allowed guild's member list; if it does not, that guild
was authorized without the `bot` scope.

### Discord scopes

**Decided 2026-09-16:** each allowed guild authorizes the app with
**`bot applications.commands`**, and no guild permissions
(ADR 0004, First deploy step 2, revised the same day):

```
https://discord.com/oauth2/authorize?client_id=<APP_ID>&scope=bot+applications.commands&permissions=0
```

`<APP_ID>` is the `CMDCTRL_DISCORD_APP_ID` value. Authorizing needs
Manage Server in the guild. It is done once per guild in Discord, not on
the host, so a rebuild of production does not repeat it.

- `applications.commands` is what `/cc-invite` and `/cc-games` need.
  Their replies are interaction responses, which need no channel
  permissions.
- `bot` makes the bot user a member of the guild. Direct-message invites
  (#613, ADR 0051 Decision 5) need that: Discord lets a bot open a DM
  only with a user who shares a guild with it.
- `permissions=0`, because DMs need no guild permission bits. Sending
  one takes only the bot token, the shared guild, and a recipient whose
  privacy settings allow DMs from members of that server and who has not
  blocked the bot. It needs no privileged gateway intent either. The
  server sends DMs over REST, and the gateway bot keeps only the
  unprivileged `Guilds` intent.

**Re-authorize every already-authorized guild once (owner action).** A
guild authorized under the old `scope=applications.commands` URL has the
commands but not the bot user. Open the URL above, pick that guild, and
authorize. Its slash commands keep working throughout. When the bot
user joins, the running bot sees the guild and re-registers the same two
commands, which is harmless.
