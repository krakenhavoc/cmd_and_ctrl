package lobby

// invite_dm.go is POST /games/{id}/invites/dm — ADR 0051 decision 5's
// direct-message invite (S34 sub-PR 6, tracking #607).
//
// The server opens a DM with Discord's REST API using a bot token
// (CMDCTRL_DISCORD_BOT_TOKEN) and posts the game's ORDINARY player
// invite link into it. Nothing new is minted: an invite DM is a
// different delivery of the same link the creator would otherwise
// paste into a channel, so it does not invalidate a link already
// shared. The gateway bot binary is not involved; the
// `/c2-invite-dm` slash command (#613) will be a thin client of this
// route, so there stays exactly one place that builds and sends an
// invite DM.
//
// The user's own OAuth tokens are irrelevant here, as decision 5
// spells out: a bot can DM any user who shares a guild with it using
// only the bot token and the user's snowflake.

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/discord"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/users"
)

// maxDMNameLen caps how much of a display name or table name is
// interpolated into the DM. Both are already capped upstream (a table
// name at 80, a Discord display name by Discord); this is the belt to
// that braces.
const maxDMNameLen = 80

// dmInviteRequest names the person to DM. Exactly one of the two
// fields.
//
//   - user_id is OUR user id (the id GET /me/tablemates returns), and
//     is what an ordinary signed-in caller uses. The server resolves it
//     to a Discord snowflake through identities; the snowflake is never
//     on the wire in either direction.
//   - discord_id is a raw Discord snowflake, and is ADMIN-ONLY. It
//     exists for #613's `/c2-invite-dm @user`, which holds a mention
//     and nothing else — the target may have never signed in here, so
//     there is no user id to send. Allowing it for everyone would turn
//     the route into "DM any Discord user who shares a guild with the
//     bot", which is a spam primitive; restricted to the admin
//     credential the bot already holds, it is the bot's own path and
//     nothing more.
type dmInviteRequest struct {
	UserID    string `json:"user_id,omitempty"`
	DiscordID string `json:"discord_id,omitempty"`
}

// dmInviteResponse is the 200 body. It deliberately carries no
// snowflake and no invite token: the caller learns that the DM was
// sent, not what was in it.
type dmInviteResponse struct {
	Sent bool `json:"sent"`
	// UserID echoes the target when the caller named one, so a client
	// that fired several can match the answer to the row.
	UserID string `json:"user_id,omitempty"`
	// DisplayName is the target's current display name when we know
	// it, for "Invite sent to Alice".
	DisplayName string `json:"display_name,omitempty"`
}

// ErrInviteTokenUnavailable is returned when the game's player invite
// plaintext is not in this process's memory — see inviteDM.
var ErrInviteTokenUnavailable = errors.New("lobby: this process does not hold the game's invite link")

// inviteDM is POST /games/{id}/invites/dm.
//
// Order of refusals: game first (404), then authorisation (403), then
// the body (400), then configuration (503), then the invite (409),
// then Discord.
func inviteDM(c Config, w http.ResponseWriter, r *http.Request) error {
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		return httpError(http.StatusInternalServerError, "missing principal")
	}
	id, err := gameIDFromPath(r)
	if err != nil {
		return err
	}
	meta, err := c.Lobby.Get(id)
	if err != nil {
		return err
	}
	createdBy, err := c.Lobby.CreatedBy(id)
	if err != nil {
		return err
	}
	if !canInviteDM(p, meta, createdBy) {
		return httpError(http.StatusForbidden,
			"only someone seated at this table, the person who created it, or the admin may send an invite DM")
	}

	var body dmInviteRequest
	if err := decodeJSON(w, r, &body); err != nil {
		return err
	}
	body.UserID = strings.TrimSpace(body.UserID)
	body.DiscordID = strings.TrimSpace(body.DiscordID)
	switch {
	case body.UserID == "" && body.DiscordID == "":
		return httpError(http.StatusBadRequest, "name the person to invite: user_id")
	case body.UserID != "" && body.DiscordID != "":
		return httpError(http.StatusBadRequest, "name the person to invite with user_id or discord_id, not both")
	case body.DiscordID != "" && p.Role != auth.RoleAdmin:
		return httpError(http.StatusForbidden, "discord_id is admin-only; name the person with user_id")
	}

	// Not configured must never fall open, and must be told apart from
	// a Discord failure: 503, naming the variable.
	if !c.DiscordBot.Enabled() {
		return httpError(http.StatusServiceUnavailable,
			"direct-message invites are not configured on this server ("+discord.BotTokenEnv+" is not set)")
	}
	if c.InviteBaseURL == "" {
		return httpError(http.StatusServiceUnavailable,
			"direct-message invites are not configured on this server (CMDCTRL_PUBLIC_BASE_URL is not set)")
	}

	recipient, displayName, err := c.dmRecipient(r, body)
	if err != nil {
		return err
	}

	// The link is the game's CURRENT player invite, read from this
	// process's memory. Only the process that minted it holds the
	// plaintext (GET /games/{id} documents the same gap), so a game
	// recovered from the database after a restart has a hash and no
	// link. We refuse rather than rotate: rotating would silently
	// revoke the link the table has already shared in a channel, which
	// is a startling side effect of "DM this to Alice". The caller is
	// pointed at POST /games/{id}/invites/rotate, which says plainly
	// what it costs.
	if meta.InviteToken == "" {
		return httpError(http.StatusConflict,
			"this server no longer holds this table's invite link (it was lost to a restart); "+
				"mint a replacement with POST /games/{id}/invites/rotate first — note that doing so "+
				"stops the old link working for everyone who already has it")
	}

	content := dmInviteMessage(inviterName(p), meta.Name, inviteURL(c.InviteBaseURL, meta.ID, meta.InviteToken))
	if err := c.DiscordBot.SendDM(r.Context(), c.DiscordHTTPClient, recipient, content); err != nil {
		return dmSendError(c, err)
	}
	return writeJSON(w, http.StatusOK, dmInviteResponse{
		Sent:        true,
		UserID:      body.UserID,
		DisplayName: displayName,
	})
}

// dmRecipient turns a request body into the Discord snowflake to DM,
// plus the target's display name when we know it.
func (c Config) dmRecipient(r *http.Request, body dmInviteRequest) (recipient, displayName string, err error) {
	if body.DiscordID != "" {
		// Admin-only path (#613): a snowflake, straight through. We
		// know nothing else about the target, by design.
		return body.DiscordID, "", nil
	}
	target, err := uuid.Parse(body.UserID)
	if err != nil || target == uuid.Nil {
		return "", "", httpError(http.StatusBadRequest, "invalid user_id")
	}
	store := c.userStore()
	u, err := store.Get(r.Context(), target)
	if err != nil {
		if errors.Is(err, users.ErrNotFound) {
			return "", "", httpError(http.StatusNotFound, "no such user")
		}
		c.logger().Error("reading a DM-invite target failed", "err", err)
		return "", "", httpError(http.StatusInternalServerError, "could not look that person up; try again")
	}
	subject, err := store.DiscordSubject(r.Context(), target)
	if err != nil {
		if errors.Is(err, users.ErrNotFound) {
			return "", "", httpError(http.StatusUnprocessableEntity,
				"that person has no Discord account linked here, so there is nowhere to send a DM")
		}
		c.logger().Error("resolving a DM-invite target's Discord id failed", "err", err)
		return "", "", httpError(http.StatusInternalServerError, "could not look that person up; try again")
	}
	return subject, u.DisplayName, nil
}

// dmSendError maps a discord.SendDM failure onto an HTTP status. The
// bot token appears in none of these, and neither does Discord's own
// error text.
func dmSendError(c Config, err error) error {
	switch {
	case errors.Is(err, discord.ErrBotNotConfigured):
		return httpError(http.StatusServiceUnavailable,
			"direct-message invites are not configured on this server ("+discord.BotTokenEnv+" is not set)")
	case errors.Is(err, discord.ErrDMForbidden):
		// 422: the request was well-formed and allowed; the target
		// simply cannot be reached this way.
		return httpError(http.StatusUnprocessableEntity, err.Error())
	case errors.Is(err, discord.ErrUnknownRecipient):
		return httpError(http.StatusNotFound, err.Error())
	case errors.Is(err, discord.ErrDMRateLimited):
		return httpError(http.StatusTooManyRequests, err.Error())
	case errors.Is(err, discord.ErrBotUnauthorized):
		// An operator problem, not a caller problem. Logged once here
		// (without the token, which never leaves the header) so it is
		// findable; the caller gets a 502 they can do nothing about.
		c.logger().Error("Discord rejected the bot token", "env", discord.BotTokenEnv)
		return httpError(http.StatusBadGateway, "Discord rejected this server's bot credentials")
	default:
		c.logger().Error("sending an invite DM failed", "err", err)
		return httpError(http.StatusBadGateway, "could not send the DM; try again")
	}
}

// canInviteDM is decision 5's caller rule: seated in, or the creator
// of, the game — or the admin.
//
// "Seated in" is satisfied two ways, because a person and a seat are
// not the same credential: a player session bound to a seat at THIS
// table, or any session carrying a UserID that holds a seat here
// (the signed-in person who has an identity session open in another
// tab). A spectator, a player at a different table, and a signed-in
// stranger are all refused.
func canInviteDM(p auth.Principal, meta GameMeta, createdBy string) bool {
	if p.Role == auth.RoleAdmin {
		return true
	}
	if p.Role == auth.RolePlayer && p.GameID == meta.ID && p.PlayerID != uuid.Nil {
		for _, s := range meta.Players {
			if s.PlayerID == p.PlayerID {
				return true
			}
		}
	}
	if p.UserID == uuid.Nil {
		return false
	}
	if _, ok := seatOfUser(meta.Players, p.UserID); ok {
		return true
	}
	return createdBy != "" && createdBy == p.UserID.String()
}

// inviterName is what the DM says invited you. A principal always has
// one of these; the fallback keeps the sentence grammatical for an
// admin session, which is a server credential and not a person.
func inviterName(p auth.Principal) string {
	for _, s := range []string{p.DiscordGlobalName, p.Name, p.DiscordUsername} {
		if s = strings.TrimSpace(s); s != "" {
			return s
		}
	}
	return "Someone"
}

// dmInviteMessage is the DM's text. Two lines: who and what, then the
// link on its own line so Discord renders it as one.
//
// Names are flattened first: a newline would let a display name forge
// a second line of the message, and a backtick or asterisk would let
// it style one. Pinging is impossible regardless — SendDM sends
// allowed_mentions with an empty parse list.
func dmInviteMessage(inviter, table, url string) string {
	return fmt.Sprintf("%s invited you to a game of cmd_and_ctrl: %s\n%s",
		flattenForDM(inviter), flattenForDM(table), url)
}

// flattenForDM strips the characters a name could use to restructure
// the message, and caps its length.
func flattenForDM(s string) string {
	s = strings.Map(func(r rune) rune {
		switch r {
		case '\n', '\r', '`', '*', '_', '~', '|', '>', '@':
			return ' '
		}
		if r < 0x20 {
			return ' '
		}
		return r
	}, s)
	s = strings.TrimSpace(strings.Join(strings.Fields(s), " "))
	return trimToLimit(s, maxDMNameLen)
}

// inviteURL builds the link the DM carries. Same shape the bot's
// /c2-invite posts (server/internal/bot/commands.go's buildInviteURL),
// so a player who has clicked one has clicked both. An invite token is
// base64url, so it needs no escaping.
func inviteURL(base string, gameID uuid.UUID, token string) string {
	return fmt.Sprintf("%s/#/games/%s/join?t=%s", strings.TrimRight(base, "/"), gameID, token)
}
