package lobby

// admins.go is ADR 0110 §3: who is an admin.
//
// Admin is a capability checked on every request, never a role baked
// into a token. Two predicates decide every admin question in this
// package, and the difference between them is the point:
//
//   - isAdmin(p): "is this an operator?" True for the shared-token
//     session (auth.RoleAdmin, from POST /admin/login), and for a
//     signed-in person whose Discord ID is on
//     CMDCTRL_DISCORD_ADMIN_USER_IDS. Almost every admin check asks
//     this.
//   - isServerCredential(p): "is this the shared server credential?"
//     True for the shared-token session only. Asked by the three paths
//     that exist for the Discord bot calling on other people's behalf:
//     its own rate buckets, filing a deck request for a named member,
//     and naming a raw Discord snowflake in a DM invite. A signed-in
//     admin is a person: they keep their own bucket, file as
//     themselves and invite by user ID.
//
// TestRoleAdminIsComparedOnlyInTheAdminPredicates holds the line: a
// bare auth.RoleAdmin anywhere else in the server fails CI, so a new
// route has to choose one of the two.

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
)

// AdminUserIDsEnv names the allowlist. It is the same variable the
// Discord bot reads for /c2-end (owner answer 3): one list, so an admin
// on Discord is an admin on the site, and removing someone removes them
// from both.
const AdminUserIDsEnv = "CMDCTRL_DISCORD_ADMIN_USER_IDS"

// AdminList is the set of Discord user IDs whose signed-in sessions are
// admins. The zero value and a nil *AdminList are both the empty list.
//
// It is held behind an atomic pointer so Replace takes effect on the
// very next request, with no lock on the read path. Production builds
// it once at boot from the environment and never replaces it, because
// the environment only changes on a deploy, which restarts the server
// and drops every socket. Replace exists so tests can show that
// nothing about admin outlives the list: a token that was admin a
// moment ago is not admin once its ID is gone.
type AdminList struct {
	ids atomic.Pointer[map[string]struct{}]
}

// ParseAdminList parses the allowlist's raw value: comma-separated
// Discord snowflakes, whitespace around each ignored. Empty or blank
// means no allowlisted admins, which is a supported state. A malformed
// entry is an error naming its position but not its value, so the boot
// can fail on a typo (which would otherwise silently deny someone)
// without the log carrying the list.
func ParseAdminList(raw string) (*AdminList, error) {
	var ids []string
	for i, part := range strings.Split(raw, ",") {
		id := strings.TrimSpace(part)
		if id == "" {
			continue
		}
		if !isSnowflake(id) {
			return nil, fmt.Errorf("%s: entry %d is not a Discord user id (digits only, at most 20)", AdminUserIDsEnv, i+1)
		}
		ids = append(ids, id)
	}
	l := &AdminList{}
	l.Replace(ids)
	return l, nil
}

// NewAdminList is ParseAdminList for callers that already hold the IDs
// (tests). It does not validate them.
func NewAdminList(ids ...string) *AdminList {
	l := &AdminList{}
	l.Replace(ids)
	return l
}

// Replace swaps the whole list. The next Has sees the new one.
func (l *AdminList) Replace(ids []string) {
	set := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		if id = strings.TrimSpace(id); id != "" {
			set[id] = struct{}{}
		}
	}
	l.ids.Store(&set)
}

// Has reports whether id is on the list. Nil-safe; "" is never on it.
func (l *AdminList) Has(id string) bool {
	if l == nil || id == "" {
		return false
	}
	set := l.ids.Load()
	if set == nil {
		return false
	}
	_, ok := (*set)[id]
	return ok
}

// Len is the number of IDs on the list, for the boot log. The IDs
// themselves are never logged and never served.
func (l *AdminList) Len() int {
	if l == nil {
		return 0
	}
	set := l.ids.Load()
	if set == nil {
		return 0
	}
	return len(*set)
}

// isServerCredential reports whether p is the shared admin token's
// session (POST /admin/login). It is the only predicate that compares
// auth.RoleAdmin; see the file comment for which paths ask it.
func isServerCredential(p auth.Principal) bool {
	return p.Role == auth.RoleAdmin
}

// isAdminPrincipal is isAdmin with the list passed in, for the callers
// that hold an *AdminList but no Config (WSAuthorizer).
//
// The UserID requirement is load-bearing (ADR 0110 §3 item 1).
// redeemSeatReclaim copies a seat's DiscordID onto a session with no
// user, so without it a reclaim ticket for an admin's seat would
// confer admin. With it, only a session minted from a real Discord
// sign-in qualifies, whatever its role: identified, player or
// spectator.
func isAdminPrincipal(admins *AdminList, p auth.Principal) bool {
	if isServerCredential(p) {
		return true
	}
	return p.UserID != uuid.Nil && p.DiscordID != "" && admins.Has(p.DiscordID)
}

// isAdmin reports whether p is an operator: the shared token, or a
// signed-in person on the allowlist. Computed per request; nothing
// about it is stored in a token.
func (c Config) isAdmin(p auth.Principal) bool {
	return isAdminPrincipal(c.Admins, p)
}

// requireAdmin is the admin-only routes' gate (ADR 0110 §3 item 2),
// in place of auth.Middleware(c.Auth, auth.RoleAdmin): any valid
// session, then isAdmin, else 403. A missing or bad credential is
// still the auth middleware's 401.
//
// Every request it lets through is logged at Info with the action and
// who did it (§3 item 7): admin_user_id for an allowlisted person,
// admin_id for the shared token. Never the token itself.
func requireAdmin(c Config, h http.Handler) http.Handler {
	return auth.Middleware(c.Auth)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p, ok := auth.PrincipalFromContext(r.Context())
		if !ok || !c.isAdmin(p) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{"error":"admin only"}`))
			return
		}
		logAdminAction(c.Log, r.Method+" "+r.URL.Path, p)
		h.ServeHTTP(w, r)
	}))
}

// logAdminAction writes the admin audit line. Nil logger: nothing.
func logAdminAction(log *slog.Logger, action string, p auth.Principal, extra ...any) {
	if log == nil {
		return
	}
	log.Info("admin action", append(append([]any{"action", action}, adminWho(p)...), extra...)...)
}

// adminWho names the admin behind p for a log line: the user for an
// allowlisted person, the per-login admin id for the shared token.
func adminWho(p auth.Principal) []any {
	if p.UserID != uuid.Nil {
		return []any{"admin_user_id", p.UserID.String()}
	}
	return []any{"admin_id", p.AdminID.String()}
}

// actorIn is the seat p acts as at game id, for the engine's "who did
// this" record: the session's own seat when it is bound to this game,
// and uuid.Nil ("the server admin") otherwise. An allowlisted admin
// managing a table they are not seated at must not be logged as their
// seat at some other game.
func actorIn(p auth.Principal, id uuid.UUID) uuid.UUID {
	if p.GameID != id {
		return uuid.Nil
	}
	return p.PlayerID
}
