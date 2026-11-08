package main

import (
	"context"
	"log/slog"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/db"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/discord"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/linkedroles"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/lobby"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/users"
)

// startLinkedRoles turns on ADR 0163's Discord Linked Roles when
// everything it needs is here, and says once at boot which state it is
// in. It needs the switch (CMDCTRL_DISCORD_LINKED_ROLES=1), the
// application id and the bot token (lr.Enabled), the sign-in config, the
// user database and CMDCTRL_IDENTITY_KEY, without which the grant could
// not be kept. On, it registers the metadata fields in the background
// (idempotent, so every boot does) and starts the push worker. Off, it
// returns nil and nothing else changes.
func startLinkedRoles(ctx context.Context, log *slog.Logger, lr discord.LinkedRoles, oauth discord.Config,
	database *db.DB, userStore users.Store) *linkedroles.Service {
	if !lr.Requested {
		log.Info("discord linked roles off; set "+discord.LinkedRolesEnv+"=1 to enable", "var", discord.LinkedRolesEnv)
		return nil
	}
	if missing := lr.Missing(); missing != "" {
		log.Info("discord linked roles off: "+discord.LinkedRolesEnv+" is 1 but "+missing+
			" is not set (the bot token is provisioned on production only)", "var", discord.LinkedRolesEnv)
		return nil
	}
	store, ok := userStore.(*users.SQLStore)
	switch {
	case !oauth.Enabled():
		log.Warn("discord linked roles off: it needs Discord sign-in (CMDCTRL_DISCORD_CLIENT_ID/SECRET/REDIRECT_URI)")
		return nil
	case database == nil || !ok:
		log.Warn("discord linked roles off: it needs the user database (CMDCTRL_DATA_DIR)")
		return nil
	case !store.CanSeal():
		log.Warn("discord linked roles off: the grant cannot be kept without "+users.IdentityKeyEnv, "var", users.IdentityKeyEnv)
		return nil
	}
	svc := linkedroles.New(linkedroles.NewDiscord(oauth, lr, nil), store, linkedroles.NewSQLStats(database), log)
	go func() {
		if err := svc.Register(ctx); err != nil {
			log.Warn("discord linked roles: registering the metadata fields failed; the next boot tries again", "err", err)
			return
		}
		log.Info("discord linked roles: metadata fields registered")
	}()
	go svc.Run(ctx)
	log.Info("discord linked roles on: the verification URL is /auth/discord/linked-role", "var", discord.LinkedRolesEnv)
	return svc
}

// lobbyLinkedRoles is svc as the lobby's interface, nil when it is off
// (a nil *Service in an interface would not be nil).
func lobbyLinkedRoles(svc *linkedroles.Service) lobby.LinkedRoles {
	if svc == nil {
		return nil
	}
	return svc
}
