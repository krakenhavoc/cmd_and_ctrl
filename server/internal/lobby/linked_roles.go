package lobby

// linked_roles.go is the lobby's half of ADR 0163, Discord Linked Roles:
// the verification URL (GET /auth/discord/linked-role), the callback's
// role-connection branch, and the push the end of a game asks for. The
// grant, the numbers and the Discord calls are internal/linkedroles.
//
// Off (Config.LinkedRoles nil, every environment but production), the
// route sends the person to the client's page saying linking is
// unavailable, no state is parked, and a game's end queues nothing.

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/discord"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/users"
)

// LinkedRoles is what the routes need from internal/linkedroles.
// *linkedroles.Service implements it.
type LinkedRoles interface {
	// Link stores the role-connection grant for the account and pushes
	// its numbers. On an error nothing is kept.
	Link(ctx context.Context, userID uuid.UUID, subject string, tok discord.TokenResponse) error
	// Forget empties a deleted account's connection on Discord, in the
	// background and best effort.
	Forget(subject string, sealed []byte)
}

// RoleConnectionQueue is what the end of a game needs: a push for each
// account that sat in it. *linkedroles.Service implements it. Enqueue
// must not block, because it is called under l.mu.
type RoleConnectionQueue interface {
	Enqueue(ids ...uuid.UUID)
}

// linkedRoleCallTimeout bounds the callback's Discord calls and the
// first push, so a hung Discord holds the redirect for seconds.
const linkedRoleCallTimeout = 15 * time.Second

// The client's page for each way a link can end (ADR 0163 §3).
const (
	linkedRoleLinked      = "/#/linked-role?status=linked"
	linkedRoleSignIn      = "/#/linked-role?status=signin"
	linkedRoleUnavailable = "/#/linked-role?status=unavailable"
	linkedRoleFailed      = "/#/linked-role?status=failed"
)

// SetLinkedRoles attaches the push queue, which turns on a push when a
// game ends. Called once at boot.
func (l *Lobby) SetLinkedRoles(q RoleConnectionQueue) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.linkedRoles = q
}

// linkedRolesEndedLocked asks for a push for every account that sat in
// entry's game, once, when it ends (syncStateLocked). Only a game the
// numbers count: one that ended with a win or a draw, not a practice
// table, and a seat that holds an account (a bot or agent seat holds
// none). Whether each account is linked is the worker's question, off
// l.mu. Callers hold l.mu.
func (l *Lobby) linkedRolesEndedLocked(entry *gameEntry) {
	if l.linkedRoles == nil || entry.practice {
		return
	}
	if entry.outcome != "win" && entry.outcome != "draw" {
		return
	}
	var ids []uuid.UUID
	for _, s := range entry.meta.Players {
		if s.IsBot || s.IsAgent || s.UserID == "" {
			continue
		}
		if id, err := uuid.Parse(s.UserID); err == nil {
			ids = append(ids, id)
		}
	}
	if len(ids) > 0 {
		l.linkedRoles.Enqueue(ids...)
	}
}

// discordLinkedRole is GET /auth/discord/linked-role, the Linked Roles
// Verification URL set in the Developer Portal (ADR 0163 §3). It starts
// the role-connection grant: scopes `identify role_connections.write`,
// the consent screen, and the same callback and redirect URI as sign-in.
// No session is needed or minted: the callback finds the account by the
// Discord id the grant names.
func discordLinkedRole(c Config, w http.ResponseWriter, r *http.Request) error {
	if c.LinkedRoles == nil || !c.Discord.Enabled() {
		http.Redirect(w, r, linkedRoleUnavailable, http.StatusFound)
		return nil
	}
	state, challenge, err := c.discordStore().StartRoleConnection()
	if err != nil {
		return err
	}
	http.Redirect(w, r, c.Discord.RoleConnectionAuthorizeURL(state, challenge), http.StatusFound)
	return nil
}

// finishRoleConnection is the callback's branch for a role-connection
// round-trip (ADR 0163 §3). It writes nothing about a person who has no
// account, and mints no session for one who does: linking is not
// signing in. Every outcome is a redirect to the client's page; the
// reason for a failure goes to the log, never to the browser.
func finishRoleConnection(c Config, w http.ResponseWriter, r *http.Request, entry discord.StateEntry, code string) {
	if c.LinkedRoles == nil {
		http.Redirect(w, r, linkedRoleUnavailable, http.StatusFound)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), linkedRoleCallTimeout)
	defer cancel()
	log := c.logger()
	failed := func(msg string, err error) {
		log.Warn(msg, "err", err)
		http.Redirect(w, r, linkedRoleFailed, http.StatusFound)
	}

	client := c.discordClient()
	token, err := c.Discord.ExchangeCode(ctx, client, code, entry.CodeVerifier)
	if err != nil {
		failed("linked roles: code exchange failed", err)
		return
	}
	user, err := c.Discord.FetchUser(ctx, client, token.AccessToken)
	if err != nil {
		failed("linked roles: reading the Discord user failed", err)
		return
	}
	userID, err := c.userStore().UserIDForDiscord(ctx, user.ID)
	if errors.Is(err, users.ErrNotFound) {
		// No account: nothing is stored, and the page offers sign-in.
		http.Redirect(w, r, linkedRoleSignIn, http.StatusFound)
		return
	}
	if err != nil {
		failed("linked roles: looking up the account failed", err)
		return
	}
	if err := c.LinkedRoles.Link(ctx, userID, user.ID, token); err != nil {
		failed("linked roles: linking failed", err)
		return
	}
	http.Redirect(w, r, linkedRoleLinked, http.StatusFound)
}
