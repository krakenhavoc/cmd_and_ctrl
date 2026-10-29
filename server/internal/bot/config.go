// Package bot implements the slash-command half of the S12.5
// Discord integration. The OAuth / identity-fetch half lives in
// package discord; the two run as separate binaries and only share
// env-var names.
//
// The bot connects to the Discord gateway with bwmarrin/discordgo
// and is gated to an allow-list of guild IDs. It calls back into
// the game server over loopback HTTP to create games and list
// them; it never imports the lobby package directly, which keeps
// the lobby's single-writer invariant intact.
package bot

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// Default endpoints. The bot runs on the same VPS as the server
// (per the S12.5 deploy plan), so loopback is the expected transport.
// The client base URL is only used to compose invite links for
// users to click; it is NOT where the bot makes HTTP calls.
const (
	defaultServerBaseURL = "http://127.0.0.1:8080"
	defaultClientBaseURL = "https://cmd.labxp.io"
)

// Config carries everything the bot needs to run. Built from env
// by ConfigFromEnv; validated by Validate. Keep fields unexported-
// from-outside-the-struct-literal style — every caller should go
// through ConfigFromEnv so the defaulting is consistent.
type Config struct {
	// BotToken is the Discord bot token (from the Developer Portal,
	// Bot tab → Reset Token). Required. Treat as a high-value
	// secret — anyone with this token can impersonate the bot.
	BotToken string

	// AppID is the Discord application ID (aka client ID). Required
	// because slash-command registration is scoped by app ID, not
	// by bot token.
	AppID string

	// GuildIDs is the allow-list. Commands are registered only on
	// these guilds, and interaction handlers reject guild IDs not
	// in this set. Required (empty rejects at startup). A Discord
	// snowflake is a ~18-digit numeric string; we do not validate
	// the format beyond non-emptiness.
	GuildIDs []string

	// ServerBaseURL is the HTTP origin the bot uses to call the
	// game server (POST /admin/login, POST /games, GET /games).
	// Defaults to loopback.
	ServerBaseURL string

	// AdminToken is the game server's admin shared secret. Same
	// value as the server's CMDCTRL_ADMIN_TOKEN. Required.
	AdminToken string

	// ClientBaseURL is the origin used to compose invite links
	// posted back to Discord. Shape:
	//   {ClientBaseURL}/#/games/{uuid}/join?t={invite_token}
	// Defaults to the prod VPS origin.
	ClientBaseURL string

	// AdminUserIDs and AdminRoleIDs together define who may run
	// /c2-end: a Discord user on AdminUserIDs, or a guild member
	// holding a role on AdminRoleIDs. Both optional and both
	// comma-separated snowflakes, same shape as GuildIDs.
	//
	// This is the "admin" half of issue #614's "host or admin"
	// check. If both lists are empty AND the game being ended has no
	// creator, /c2-end refuses every caller rather than failing open;
	// see Config.IsAdmin. The "host" half — the game's own creator,
	// now that games.created_by exists (#1044, #1098) — is not a
	// Config field: it is per-game, so it is answered by the server
	// (GET /games/{id}/creator) and combined with IsAdmin in
	// Handler.mayEnd (end.go), not by this Config alone.
	AdminUserIDs []string
	AdminRoleIDs []string

	// CardChannelIDs lists the channels where the bot watches for
	// [[Card Name]] and replies with the card (#3324). Empty turns the
	// watcher off, and then the bot does not ask Discord for the
	// privileged Message Content intent at all. /c2-card works either way.
	CardChannelIDs []string

	// AccessRequestChannelID is the community server's #account-requests
	// channel (ADR 0156's 2026-10-16 amendment, #3325). A member without
	// the player role who posts there gets an access request with
	// Approve and Decline buttons for the mods. Empty turns it off, and
	// then the bot does not ask Discord for message events on its
	// account. Needs PlayerRoleID.
	AccessRequestChannelID string

	// PlayerRoleID is the role Approve grants: Elder Dragon Highlander on
	// the C2 server, the same CMDCTRL_DISCORD_PLAYER_ROLE_ID the server's
	// sign-up gate checks (ADR 0156 §1). The bot needs Manage Roles and a
	// role above it.
	PlayerRoleID string

	// ModLogChannelID is the private #mod-log channel, the same
	// CMDCTRL_DISCORD_MOD_LOG_CHANNEL_ID the server posts its entries to.
	// The bot adds a line there when a mod approves or declines an access
	// request. Empty: it adds nothing.
	ModLogChannelID string

	// BanSync is CMDCTRL_DISCORD_BAN_SYNC=1 (ADR 0156's 2026-10-29
	// amendment, #3325): when the community server bans a member, takes
	// the player role away or sees them leave, the bot asks the server to
	// end that person's sessions at once. Off by default, because it needs
	// the privileged Server Members intent, and asking for that before it
	// is switched on in the Developer Portal makes Discord close the
	// connection (4014). Needs CommunityGuildID and PlayerRoleID.
	BanSync bool

	// CommunityGuildID is the C2 server, the same
	// CMDCTRL_DISCORD_COMMUNITY_GUILD_ID the server's sign-up gate reads.
	// Ban sync acts only on its events.
	CommunityGuildID string
}

// ConfigFromEnv reads the CMDCTRL_* env vars. Missing optional
// vars get defaults; missing required vars are left blank so
// Validate can surface them all at once.
func ConfigFromEnv() Config {
	return Config{
		BotToken:       strings.TrimSpace(os.Getenv("CMDCTRL_DISCORD_BOT_TOKEN")),
		AppID:          strings.TrimSpace(os.Getenv("CMDCTRL_DISCORD_APP_ID")),
		GuildIDs:       splitGuildIDs(os.Getenv("CMDCTRL_DISCORD_GUILD_IDS")),
		ServerBaseURL:  orDefault(os.Getenv("CMDCTRL_SERVER_BASE_URL"), defaultServerBaseURL),
		AdminToken:     strings.TrimSpace(os.Getenv("CMDCTRL_ADMIN_TOKEN")),
		ClientBaseURL:  orDefault(os.Getenv("CMDCTRL_CLIENT_BASE_URL"), defaultClientBaseURL),
		AdminUserIDs:   splitGuildIDs(os.Getenv("CMDCTRL_DISCORD_ADMIN_USER_IDS")),
		AdminRoleIDs:   splitGuildIDs(os.Getenv("CMDCTRL_DISCORD_ADMIN_ROLE_IDS")),
		CardChannelIDs: splitGuildIDs(os.Getenv("CMDCTRL_DISCORD_CARD_CHANNEL_IDS")),

		AccessRequestChannelID: strings.TrimSpace(os.Getenv(AccessRequestChannelEnv)),
		PlayerRoleID:           strings.TrimSpace(os.Getenv(PlayerRoleEnv)),
		ModLogChannelID:        strings.TrimSpace(os.Getenv(ModLogChannelEnv)),

		BanSync:          strings.TrimSpace(os.Getenv(BanSyncEnv)) == "1",
		CommunityGuildID: strings.TrimSpace(os.Getenv(CommunityGuildEnv)),
	}
}

// BanSyncEnv turns ban sync on with "1". CommunityGuildEnv is the
// server's variable of the same name.
const (
	BanSyncEnv        = "CMDCTRL_DISCORD_BAN_SYNC"
	CommunityGuildEnv = "CMDCTRL_DISCORD_COMMUNITY_GUILD_ID"
)

// BanSyncEnabled reports whether the bot listens for bans, role
// removals and leaves in the community server.
func (c Config) BanSyncEnabled() bool {
	return c.BanSync && c.CommunityGuildID != "" && c.PlayerRoleID != ""
}

// Intents is what the bot asks Discord for when it connects. Each
// feature adds only what it needs, so a feature that is off never asks
// for a privileged intent the Developer Portal has not switched on: an
// intent asked for and not enabled makes Discord refuse the connection
// (close code 4014), and the whole bot with it.
func (c Config) Intents() discordgo.Intent {
	// IntentsGuilds lets us observe GuildCreate so the bot can
	// register (or explicitly refuse to register) commands on a
	// guild that was added while the process was running. Not
	// privileged; no portal approval required.
	intents := discordgo.IntentsGuilds
	// The [[Card Name]] watcher (#3324) reads message text, which needs the
	// privileged Message Content intent, switched on in the Developer
	// Portal. Ask for it only when card channels are configured.
	if len(c.CardChannelIDs) > 0 {
		intents |= discordgo.IntentsGuildMessages | discordgo.IntentsMessageContent
	}
	// Access requests (ADR 0156's 2026-10-16 amendment, #3325) need to
	// see who posts in #account-requests, which is the unprivileged
	// Guild Messages intent; they never read the text.
	if c.AccessRequestsEnabled() {
		intents |= discordgo.IntentsGuildMessages
	}
	// Ban sync (ADR 0156's 2026-10-29 amendment): GUILD_BAN_ADD comes
	// with the unprivileged Guild Moderation intent; GUILD_MEMBER_UPDATE
	// and GUILD_MEMBER_REMOVE need the privileged Server Members intent.
	// OpenGateway drops the latter, and keeps bans, if Discord refuses it.
	if c.BanSyncEnabled() {
		intents |= discordgo.IntentGuildModeration | discordgo.IntentGuildMembers
	}
	return intents
}

// The variables behind the access requests and the mod log (ADR 0156's
// 2026-10-16 amendment). The player role and the mod log channel are
// the server's variables of the same names.
const (
	AccessRequestChannelEnv = "CMDCTRL_DISCORD_ACCESS_REQUEST_CHANNEL_ID"
	PlayerRoleEnv           = "CMDCTRL_DISCORD_PLAYER_ROLE_ID"
	ModLogChannelEnv        = "CMDCTRL_DISCORD_MOD_LOG_CHANNEL_ID"
)

// AccessRequestsEnabled reports whether the bot posts access requests.
func (c Config) AccessRequestsEnabled() bool {
	return c.AccessRequestChannelID != "" && c.PlayerRoleID != ""
}

// Disabled reports whether the bot should no-op at startup.
// An empty bot token is the canonical "feature off" signal,
// matching how the OAuth half degrades in package discord.
func (c Config) Disabled() bool {
	return c.BotToken == ""
}

// Validate returns the first problem found with the config, or
// nil if the config is good to boot with. Callers that observe
// Disabled() == true should skip Validate and exit cleanly.
func (c Config) Validate() error {
	if c.BotToken == "" {
		return errors.New("CMDCTRL_DISCORD_BOT_TOKEN is required")
	}
	if c.AppID == "" {
		return errors.New("CMDCTRL_DISCORD_APP_ID is required")
	}
	if len(c.GuildIDs) == 0 {
		return errors.New("CMDCTRL_DISCORD_GUILD_IDS must list at least one guild snowflake (comma-separated)")
	}
	if c.AdminToken == "" {
		return errors.New("CMDCTRL_ADMIN_TOKEN is required (same value as the game server)")
	}
	if c.ServerBaseURL == "" {
		return errors.New("CMDCTRL_SERVER_BASE_URL is required")
	}
	if c.ClientBaseURL == "" {
		return errors.New("CMDCTRL_CLIENT_BASE_URL is required")
	}
	for _, v := range []struct{ name, value string }{
		{AccessRequestChannelEnv, c.AccessRequestChannelID},
		{PlayerRoleEnv, c.PlayerRoleID},
		{ModLogChannelEnv, c.ModLogChannelID},
		{CommunityGuildEnv, c.CommunityGuildID},
	} {
		if v.value != "" && !isSnowflake(v.value) {
			return fmt.Errorf("%s is not a Discord id (digits only, at most 20)", v.name)
		}
	}
	if c.AccessRequestChannelID != "" && c.PlayerRoleID == "" {
		return fmt.Errorf("%s is set but %s is not: Approve grants that role, so name it", AccessRequestChannelEnv, PlayerRoleEnv)
	}
	if c.BanSync && (c.CommunityGuildID == "" || c.PlayerRoleID == "") {
		return fmt.Errorf("%s=1 needs %s and %s: ban sync acts on that server's events and that role", BanSyncEnv, CommunityGuildEnv, PlayerRoleEnv)
	}
	return nil
}

// isSnowflake reports whether s looks like a Discord id: 1 to 20 digits.
func isSnowflake(s string) bool {
	if s == "" || len(s) > 20 {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// Redacted returns a loggable copy with the two secrets replaced
// by a length indicator. Safe to pass to slog.
func (c Config) Redacted() map[string]any {
	return map[string]any{
		"app_id":           c.AppID,
		"guild_ids":        c.GuildIDs,
		"server_base_url":  c.ServerBaseURL,
		"client_base_url":  c.ClientBaseURL,
		"bot_token":        fmt.Sprintf("<%d chars>", len(c.BotToken)),
		"admin_token":      fmt.Sprintf("<%d chars>", len(c.AdminToken)),
		"admin_user_ids":   c.AdminUserIDs,
		"admin_role_ids":   c.AdminRoleIDs,
		"card_channel_ids": c.CardChannelIDs,

		"access_request_channel_id": c.AccessRequestChannelID,
		"player_role_id":            c.PlayerRoleID,
		"mod_log_channel_id":        c.ModLogChannelID,
		"ban_sync":                  c.BanSync,
		"community_guild_id":        c.CommunityGuildID,
	}
}

// IsAdmin reports whether the interaction's invoker is authorized to
// run /c2-end: a Discord user on AdminUserIDs, or a guild member
// holding a role on AdminRoleIDs. An empty configuration (both lists
// empty) never authorizes anyone — see the AdminUserIDs/AdminRoleIDs
// doc comment.
func (c Config) IsAdmin(i *discordgo.Interaction) bool {
	userID, roles := interactionInvoker(i)
	return c.isAdminID(userID, roles)
}

// isAdminID is IsAdmin for a user id and roles read from somewhere other
// than an interaction (a message's author, for the access requests).
func (c Config) isAdminID(userID string, roles []string) bool {
	if userID == "" {
		return false
	}
	for _, id := range c.AdminUserIDs {
		if id == userID {
			return true
		}
	}
	for _, role := range roles {
		for _, allowed := range c.AdminRoleIDs {
			if role == allowed {
				return true
			}
		}
	}
	return false
}

// HasAdmins reports whether any admin (user or role) is configured
// at all. Used to distinguish "you specifically are not an admin"
// from "nobody is configured as an admin yet" in the refusal
// message /c2-end gives an unauthorized caller.
func (c Config) HasAdmins() bool {
	return len(c.AdminUserIDs) > 0 || len(c.AdminRoleIDs) > 0
}

// CardChannelAllowed reports whether the [[Card Name]] watcher is on in
// the channel.
func (c Config) CardChannelAllowed(channelID string) bool {
	for _, id := range c.CardChannelIDs {
		if id == channelID {
			return true
		}
	}
	return false
}

// GuildAllowed reports whether the given guild ID is in the
// allow-list. Used both at registration time (skip non-allowed
// guilds) and inside the interaction handler (defense-in-depth
// reject).
func (c Config) GuildAllowed(guildID string) bool {
	for _, g := range c.GuildIDs {
		if g == guildID {
			return true
		}
	}
	return false
}

// splitGuildIDs parses the comma-separated env var. Trims
// whitespace around each entry and drops empties, so
// "123 , ,456" yields ["123", "456"].
func splitGuildIDs(raw string) []string {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

func orDefault(v, def string) string {
	if t := strings.TrimSpace(v); t != "" {
		return t
	}
	return def
}
