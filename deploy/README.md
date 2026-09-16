# deploy/

| File | Installed on | Installed by |
|---|---|---|
| `Caddyfile` | both hosts | CD, every deploy of `main` and `develop` ("Deploy Caddy config") |
| `cmd-and-ctrl-bot.service` | production only | CD, every deploy of `main` ("Install bot systemd unit"); **enabling** it is an operator step, repeated after every rebuild, below |

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
   their defaults, which are production's values. The step skips with a
   notice while the secret is unset, or while the `cmdctrl-bot` group
   does not exist.
4. Restarts the bot, **only if the unit is enabled**.

Rotating the bot token: update the secret, then start a new `main`
deploy (see step 3 below for why a new run rather than a rerun).

### Operator setup (running VM and every rebuild)

The HomeLab cloud-init template does not create the `cmdctrl-bot` user
yet (the HomeLab half of #249). Until it does, these steps are needed on
the running production VM **and again after every rebuild of
production**. A rebuilt VM comes up with no `cmdctrl-bot` user and no
enabled unit, which is how the bot was lost in the August reprovision.

CD will not tell you. On such a host "Sync bot env" skips with only a
notice, "Install bot systemd unit" installs the unit without enabling
it, the restart step skips, and every deploy stays green with no bot.
After any rebuild, check `systemctl is-enabled cmd-and-ctrl-bot` and
`id cmdctrl-bot`.

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
   no notice about a missing group or secret.
4. Enable and start the bot:

   ```sh
   sudo systemctl enable --now cmd-and-ctrl-bot
   ```

   From then on every `main` deploy restarts it.

### Verify

`systemctl is-active` is not proof. The unit is `Type=simple`, so it
reports active the moment the process forks, even when the bot exits a
second later and systemd restarts it every 5 seconds. Check:

```sh
journalctl -u cmd-and-ctrl-bot -n 50 --no-pager   # "bot config loaded", then "discord session ready"
systemctl show -p NRestarts cmd-and-ctrl-bot      # stays 0 over a minute
sudo stat -c '%U:%G %a' /etc/cmd_and_ctrl/bot.env # root:cmdctrl-bot 640
```

`bot config invalid` names the missing key; `bot disabled` means the
token is empty. Then in an allowed guild, `/cc-games` should answer with
an ephemeral list. A reply saying the bot is not authorized against the
game server means the two admin tokens differ.

### Discord scopes

ADR 0004 (First deploy, step 2) authorizes the app in each allowed
guild with the `applications.commands` scope and nothing else:
`https://discord.com/oauth2/authorize?client_id=<APP_ID>&scope=applications.commands`.
That is enough for `/cc-invite` and `/cc-games`, and it is the recorded
decision, so follow it.

**Open question for the owner, not settled here:** direct-message
invites (#613, ADR 0051 Decision 5) send DMs as the bot, which only
works for users who share a guild with the bot user, and the bot user
only joins a guild when the app is authorized with the `bot` scope too.
If #613 ships as designed, each guild will need re-authorizing with
`scope=bot+applications.commands`, and ADR 0004 needs amending to say
so.
