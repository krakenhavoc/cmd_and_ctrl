# deploy/

Unit files for the systemd services that run cmd_and_ctrl on the VPS.
Host provisioning and the full environment matrix live in
[`docs/environments.md`](../docs/environments.md).

| File | Service | Deployed by |
|---|---|---|
| `cmd-and-ctrl-dev.service` | develop preview game server | push to `develop` |
| `cmd-and-ctrl-bot.service` | Discord bot (production only) | push to `main` |

## Missing: the production game-server unit

`cmd-and-ctrl.service` is **not** in this directory. It was installed
by hand (via the HomeLab cloud-init template) and has only ever
existed on the host, which means the running production configuration
is not under version control.

Capture it rather than reconstructing it — a guessed unit that gets
copied over the working one is how a preview environment takes down
production:

```sh
ssh krkn@$CMDCTRL_HOST 'systemctl cat cmd-and-ctrl' \
  | sed '1d' > deploy/cmd-and-ctrl.service
```

Review the result (drop any host-specific absolute paths that belong
in the env file instead), commit it, and only then treat this
directory as the source of truth for production. Until that happens,
do **not** `cp` anything from here over `/etc/systemd/system/cmd-and-ctrl.service`.
