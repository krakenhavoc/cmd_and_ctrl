package lobby

// ban_sync.go is ban sync (ADR 0156, amendment 2026-10-29, #3325): the
// server half of ending a person's sessions the moment the C2 Discord
// server bans them, takes away the player role, or sees them leave.
//
// The gateway bot (a separate process, server/cmd/bot) hears those
// events and calls POST /admin/discord-members/{discord_id}/revoke-sessions
// with its admin session, the same way its other commands call back in.
// Everything that needs the user database or the server's own lists is
// decided here, not in the bot:
//
//   - The sign-up gate is off: nothing to sync, since the role admits no
//     one. Answered "gate_off".
//   - An allowlisted admin (CMDCTRL_DISCORD_ADMIN_USER_IDS) or an account
//     on CMDCTRL_DISCORD_GATE_EXEMPT_USER_IDS: never revoked, the same
//     exemption the gate gives them at sign-in and renewal. "exempt".
//   - No account for that Discord id: nothing to do, nothing logged.
//     "no_account".
//   - The last recorded gate answer is already a refusal: the person was
//     already put out (by an earlier event, or at renewal) and has not
//     signed in since, since a sign-in records "admitted". "already_out".
//     This is what makes the route idempotent: a ban arrives with a
//     member-remove, and a role removal may be followed by a leave, and
//     only the first of them revokes and reaches the mod log.
//   - Otherwise: revokeUser, the code path POST /admin/users/{id}/revoke-sessions
//     and /logout/everywhere share; the refusal is recorded as the gate's
//     answer (not_member for a ban or a leave, no_role for a role taken
//     away), so the admin view shows it; and a line goes to the mod log.
//     "revoked".

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/discord"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/users"
)

// MemberRemoval is why the C2 server no longer admits a person.
type MemberRemoval string

const (
	// MemberBanned is GUILD_BAN_ADD.
	MemberBanned MemberRemoval = "banned"
	// MemberRoleRemoved is GUILD_MEMBER_UPDATE without the player role.
	MemberRoleRemoved MemberRemoval = "role_removed"
	// MemberLeft is GUILD_MEMBER_REMOVE: they left, or were kicked.
	MemberLeft MemberRemoval = "left"
)

// access is the gate answer the removal amounts to.
func (m MemberRemoval) access() (discord.Access, bool) {
	switch m {
	case MemberBanned, MemberLeft:
		return discord.AccessNotMember, true
	case MemberRoleRemoved:
		return discord.AccessNoRole, true
	}
	return "", false
}

// What the ban sync route did.
const (
	BanSyncRevoked    = "revoked"
	BanSyncNoAccount  = "no_account"
	BanSyncExempt     = "exempt"
	BanSyncGateOff    = "gate_off"
	BanSyncAlreadyOut = "already_out"
)

type banSyncRequest struct {
	Reason MemberRemoval `json:"reason"`
}

// banSyncResponse is the route's answer. It names no account: the bot
// needs only the outcome, and its logs stay free of our ids.
type banSyncResponse struct {
	Outcome       string `json:"outcome"`
	SocketsClosed int    `json:"sockets_closed"`
}

// accessReader reads the last recorded gate answer. *users.SQLStore
// implements it; a store that does not skips the already-out check and
// always revokes, which is safe, only noisier.
type accessReader interface {
	AccessCheck(ctx context.Context, id uuid.UUID) (time.Time, discord.Access, error)
}

var _ accessReader = (*users.SQLStore)(nil)

// banSyncMu serializes the route. A ban arrives together with a
// member-remove, each event on its own goroutine in the bot, and the
// already-out check followed by the revoke must not interleave, or both
// would revoke and both would reach the mod log.
var banSyncMu sync.Mutex

// adminBanSync is POST /admin/discord-members/{discord_id}/revoke-sessions.
func adminBanSync(c Config, w http.ResponseWriter, r *http.Request) error {
	discordID := strings.TrimSpace(r.PathValue("discord_id"))
	if !isSnowflake(discordID) {
		return httpError(http.StatusBadRequest, "discord_id is not a Discord id")
	}
	var req banSyncRequest
	if err := decodeJSON(w, r, &req); err != nil {
		return err
	}
	result, ok := req.Reason.access()
	if !ok {
		return httpError(http.StatusBadRequest, `reason must be "banned", "role_removed" or "left"`)
	}
	if !c.Discord.Gate.Enabled() {
		return writeJSON(w, http.StatusOK, banSyncResponse{Outcome: BanSyncGateOff})
	}
	if !c.gateApplies(discordID) {
		return writeJSON(w, http.StatusOK, banSyncResponse{Outcome: BanSyncExempt})
	}

	banSyncMu.Lock()
	defer banSyncMu.Unlock()

	ctx := r.Context()
	userID, err := c.userStore().UserIDForDiscord(ctx, discordID)
	if errors.Is(err, users.ErrNotFound) {
		return writeJSON(w, http.StatusOK, banSyncResponse{Outcome: BanSyncNoAccount})
	}
	if err != nil {
		c.logger().Error("ban sync: looking up the account failed", "reason", req.Reason, "err", err)
		return httpError(http.StatusInternalServerError, "could not look up the account; try again")
	}
	if store, ok := c.userStore().(accessReader); ok {
		if _, last, err := store.AccessCheck(ctx, userID); err == nil && (last == discord.AccessNotMember || last == discord.AccessNoRole) {
			return writeJSON(w, http.StatusOK, banSyncResponse{Outcome: BanSyncAlreadyOut})
		}
	}

	out, err := revokeUser(ctx, c, userID)
	if err != nil {
		return err
	}
	c.recordAccess(ctx, userID, result)
	c.logger().Info("ban sync: sessions revoked", "user_id", userID, "reason", req.Reason, "sockets_closed", out.SocketsClosed)
	if c.ModLog != nil {
		var name string
		if u, uerr := c.userStore().Get(ctx, userID); uerr == nil {
			name = u.DisplayName
		}
		c.ModLog.DiscordRemoval(userID, name, req.Reason, out.SocketsClosed)
	}
	return writeJSON(w, http.StatusOK, banSyncResponse{Outcome: BanSyncRevoked, SocketsClosed: out.SocketsClosed})
}
