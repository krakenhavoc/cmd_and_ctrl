# deploy/

| File | Service | Deployed by |
|---|---|---|
| `cmd-and-ctrl-bot.service` | Discord bot (production only) | push to `main` |

## The game server's unit is not here

Both hosts get `cmd-and-ctrl.service`, the Caddyfile and
`/etc/cmd_and_ctrl/env` from the cloud-init template in
[krakenhavoc/HomeLab](https://github.com/krakenhavoc/HomeLab):
`terraform/deployments/lab/templates/setup-cmd_and_ctrl.yaml.tftpl`.

That template is the single source of truth for both production and the
develop preview. To change host configuration, change it there and
apply — do not edit files on the box. A rebuild reverts them, which is
how production ended up serving Discord OAuth from a Caddyfile the
template did not contain.

See [docs/environments.md](../docs/environments.md).

## The bot unit is still here

Because the bot is production-only and is installed by hand:

```sh
sudo cp deploy/cmd-and-ctrl-bot.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now cmd-and-ctrl-bot
```

Never install it on the preview host. Two bot processes logged into the
same Discord application both answer every slash command.
