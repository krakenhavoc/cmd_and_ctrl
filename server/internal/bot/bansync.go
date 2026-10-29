package bot

// bansync.go is ban sync (ADR 0156's 2026-10-29 amendment, #3325). With
// CMDCTRL_DISCORD_BAN_SYNC=1, three gateway events from the community
// server (CMDCTRL_DISCORD_COMMUNITY_GUILD_ID) end a person's CMD & CTRL
// sessions at once, instead of at their next renewal:
//
//   - GUILD_BAN_ADD: banned.
//   - GUILD_MEMBER_UPDATE where the member no longer holds the player
//     role (CMDCTRL_DISCORD_PLAYER_ROLE_ID): the role was taken away.
//   - GUILD_MEMBER_REMOVE: they left, or were kicked.
//
// The bot decides only whether an event is one of those. It then calls
// the server's POST /admin/discord-members/{discord_id}/revoke-sessions
// with its admin session, as every other command calls back in, and the
// server does the rest: it finds the account by Discord id, skips admins,
// the gate's exemption list and anyone already put out, revokes through
// the same path as POST /admin/users/{id}/revoke-sessions, and posts the
// mod log line. A Discord user with no account costs one loopback call
// and nothing else: no log line, no mod log.
//
// The member events need the privileged Server Members intent; bans need
// only Guild Moderation. Config.Intents asks for both when ban sync is
// on, and OpenGateway drops Server Members, keeping bans, if Discord
// refuses it with close code 4014.

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gorilla/websocket"
)

// The reasons the server's ban sync route takes (lobby.MemberRemoval).
const (
	banSyncBanned      = "banned"
	banSyncRoleRemoved = "role_removed"
	banSyncLeft        = "left"
)

const (
	// banSyncCallTimeout bounds one call to the server.
	banSyncCallTimeout = 10 * time.Second
	// banSyncAttempts is how many times a call the server did not answer
	// (a restart during a deploy) is tried. Past it, the person is put
	// out at their next renewal instead, as before ban sync.
	banSyncAttempts = 3
	// memberRemoveDelay holds a GUILD_MEMBER_REMOVE back before it is
	// sent. A ban removes the member too, and Discord does not promise
	// which of the two events comes first; the wait lets the ban's call
	// land first, so the mod log says "banned" and the remove finds the
	// person already out.
	memberRemoveDelay = 5 * time.Second
)

// closeDisallowedIntents is Discord's gateway close code for an intent
// asked for and not enabled in the Developer Portal.
const closeDisallowedIntents = 4014

// IsDisallowedIntents reports whether err is Discord closing the gateway
// with 4014. discordgo's Session.Open returns the websocket's close
// error as it is when Discord answers the identify that way.
func IsDisallowedIntents(err error) bool {
	var ce *websocket.CloseError
	return errors.As(err, &ce) && ce.Code == closeDisallowedIntents
}

// OpenGateway opens the session. If Discord refuses the connection with
// 4014 while the bot asked for the Server Members intent (ban sync), it
// says so in an ERROR line naming the fix, drops that one intent and
// opens again, so the slash commands, the card watcher and the access
// requests keep working and bans still sync. Any other failure, and a
// second 4014 (Message Content, for the card watcher), is returned as it
// is.
func OpenGateway(s *discordgo.Session, log *slog.Logger) error {
	return openWithIntentFallback(s.Open, &s.Identify.Intents, log)
}

func openWithIntentFallback(open func() error, intents *discordgo.Intent, log *slog.Logger) error {
	err := open()
	if err == nil || !IsDisallowedIntents(err) || *intents&discordgo.IntentGuildMembers == 0 {
		return err
	}
	log.Error("Discord refused the Server Members intent (gateway close 4014), so ban sync cannot see role removals or leaves. "+
		"Turn on Server Members Intent in the Discord Developer Portal (the application, Bot, Privileged Gateway Intents) and restart the bot. "+
		"Connecting without it: bans still end sessions, and everything else works.",
		"setting", BanSyncEnv, "error", err.Error())
	*intents &^= discordgo.IntentGuildMembers
	return open()
}

// banSyncReasonForBan is the reason a GUILD_BAN_ADD asks for, or "" when
// the event is not ban sync's to act on.
func (c Config) banSyncReasonForBan(e *discordgo.GuildBanAdd) (userID, reason string) {
	if e == nil || e.User == nil || !c.banSyncApplies(e.GuildID, e.User) {
		return "", ""
	}
	return e.User.ID, banSyncBanned
}

// banSyncReasonForUpdate is the reason a GUILD_MEMBER_UPDATE asks for:
// the member does not hold the player role now, and did not lose it
// before this update as far as the bot knows. Discord sends an update for
// every nickname, avatar and role change; when the state cache had the
// member and they already lacked the role, nothing was taken away now
// (they never had it, or they lost it earlier and that update was acted
// on). Without a cached copy the server's own check keeps it idempotent.
func (c Config) banSyncReasonForUpdate(e *discordgo.GuildMemberUpdate) (userID, reason string) {
	if e == nil || e.Member == nil || e.User == nil || !c.banSyncApplies(e.GuildID, e.User) {
		return "", ""
	}
	if hasRole(e.Roles, c.PlayerRoleID) {
		return "", ""
	}
	if e.BeforeUpdate != nil && !hasRole(e.BeforeUpdate.Roles, c.PlayerRoleID) {
		return "", ""
	}
	return e.User.ID, banSyncRoleRemoved
}

// banSyncReasonForRemove is the reason a GUILD_MEMBER_REMOVE asks for.
func (c Config) banSyncReasonForRemove(e *discordgo.GuildMemberRemove) (userID, reason string) {
	if e == nil || e.Member == nil || e.User == nil || !c.banSyncApplies(e.GuildID, e.User) {
		return "", ""
	}
	return e.User.ID, banSyncLeft
}

// banSyncApplies is the part every event shares: ban sync is on, the
// event is the community server's, and the person is neither a bot nor
// on the bot's admin list. The server skips admins and the gate's
// exemption list too; this saves it the call.
func (c Config) banSyncApplies(guildID string, u *discordgo.User) bool {
	if !c.BanSyncEnabled() || guildID != c.CommunityGuildID {
		return false
	}
	if u.ID == "" || u.Bot || u.System {
		return false
	}
	return !c.isAdminID(u.ID, nil)
}

// HandleBan is the GUILD_BAN_ADD callback.
func (h *Handler) HandleBan(e *discordgo.GuildBanAdd) {
	if id, reason := h.cfg.banSyncReasonForBan(e); id != "" {
		h.syncRemoval(id, reason)
	}
}

// HandleMemberUpdate is the GUILD_MEMBER_UPDATE callback.
func (h *Handler) HandleMemberUpdate(e *discordgo.GuildMemberUpdate) {
	if id, reason := h.cfg.banSyncReasonForUpdate(e); id != "" {
		h.syncRemoval(id, reason)
	}
}

// HandleMemberRemove is the GUILD_MEMBER_REMOVE callback. It waits
// memberRemoveDelay first, so a ban's own event is acted on before it.
func (h *Handler) HandleMemberRemove(e *discordgo.GuildMemberRemove) {
	if id, reason := h.cfg.banSyncReasonForRemove(e); id != "" {
		h.pause(memberRemoveDelay)
		h.syncRemoval(id, reason)
	}
}

// syncRemoval asks the server to end the person's sessions, trying again
// while the server does not answer. discordgo runs each event handler on
// its own goroutine, so the waits block nothing else.
func (h *Handler) syncRemoval(discordID, reason string) {
	for attempt := 1; ; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), banSyncCallTimeout)
		res, err := h.client.RevokeDiscordMember(ctx, discordID, reason)
		cancel()
		if err == nil {
			switch res.Outcome {
			case BanSyncRevoked:
				h.log.Info("ban sync: sessions revoked", "member", discordID, "reason", reason, "sockets_closed", res.SocketsClosed)
			case BanSyncGateOff:
				h.log.Warn("ban sync: the server's sign-up gate is off, so nothing was revoked", "reason", reason)
			}
			// no_account, exempt and already_out: nothing to say.
			return
		}
		if !errors.Is(err, ErrServerUnreachable) || attempt >= banSyncAttempts {
			h.log.Warn("ban sync: the server did not revoke; the person is put out at their next renewal instead",
				"member", discordID, "reason", reason, "attempt", attempt, "error", err.Error())
			return
		}
		h.pause(time.Duration(attempt) * 2 * time.Second)
	}
}

// pause waits d, or what a test put in Handler.sleep.
func (h *Handler) pause(d time.Duration) {
	if h.sleep != nil {
		h.sleep(d)
		return
	}
	time.Sleep(d)
}
