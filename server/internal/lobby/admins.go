package lobby

// admins.go is ADR 0110 §3 and ADR 0112 §2: who is an admin.
//
// Admin is a capability checked on every request, never a role baked
// into a token. Two predicates decide every admin question in this
// package, and the difference between them is the point:
//
//   - isAdmin(p): "is this an operator?" True for the shared-token
//     session (auth.RoleAdmin, from POST /admin/login), and for a
//     signed-in person whose Discord ID is on
//     CMDCTRL_DISCORD_ADMIN_USER_IDS AND who has switched admin mode on
//     in the last 12 hours (ADR 0112 §2). Almost every admin check asks
//     this. An allowlisted person in player mode, the default, gets
//     exactly what any other signed-in person gets.
//   - isServerCredential(p): "is this the shared server credential?"
//     True for the shared-token session only. Asked by the three paths
//     that exist for the Discord bot calling on other people's behalf:
//     its own rate buckets, filing a deck request for a named member,
//     and naming a raw Discord snowflake in a DM invite. A signed-in
//     admin is a person: they keep their own bucket, file as
//     themselves and invite by user ID.
//
// A third, isAllowlisted(p), is "on the list, whatever the mode". Only
// the switch (PUT /me/admin-mode) and /me's admin_allowed ask it: it
// decides who may turn admin mode on, never what an admin may do.
//
// TestRoleAdminIsComparedOnlyInTheAdminPredicates holds the line: a
// bare auth.RoleAdmin anywhere else in the server fails CI, so a new
// route has to choose one of the two. Its second rule,
// TestAllowlistAndModeAreAskedOnlyInAdminsGo, fails CI on a call to
// AdminList.Has or AdminModes.On outside this file, so a route cannot
// test the allowlist and skip the mode.

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/users"
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

// Admins is what decides whether a person is an admin (ADR 0112 §2
// item 2): the allowlist and each person's admin mode. lobby.Config and
// WSAuthorizer share one instance, so HTTP and WebSockets can never
// disagree. A nil *Admins, and either half nil, is "nobody but the
// shared token".
//
// Its fields are unexported on purpose: the only questions anything
// outside this file may ask are isAdmin, isAdminPrincipal and
// isAllowlisted.
type Admins struct {
	list  *AdminList
	modes *users.AdminModes
}

// NewAdmins bundles the allowlist and the modes. Either may be nil: a
// nil list allows nobody, and nil modes leave every allowlisted person
// in player mode (a server with no user database has no users, so no
// mode to keep).
func NewAdmins(list *AdminList, modes *users.AdminModes) *Admins {
	return &Admins{list: list, modes: modes}
}

// AllowlistLen is the number of IDs on the allowlist, for the boot log.
func (a *Admins) AllowlistLen() int {
	if a == nil {
		return 0
	}
	return a.list.Len()
}

// modeEndsAt is when userID's admin mode lapses, and whether it is on.
func (a *Admins) modeEndsAt(userID uuid.UUID) (time.Time, bool) {
	if a == nil || a.modes == nil {
		return time.Time{}, false
	}
	return a.modes.EndsAt(userID, a.modes.Now())
}

// isServerCredential reports whether p is the shared admin token's
// session (POST /admin/login). It is the only predicate that compares
// auth.RoleAdmin; see the file comment for which paths ask it.
func isServerCredential(p auth.Principal) bool {
	return p.Role == auth.RoleAdmin
}

// isAllowlisted reports whether p is a signed-in person on the
// allowlist, whatever their admin mode. Asked by the switch route and
// by /me's admin_allowed, and by nothing else: it says who MAY turn
// admin mode on, never that someone is an admin.
//
// The UserID requirement is load-bearing (ADR 0110 §3 item 1).
// redeemSeatReclaim copies a seat's DiscordID onto a session with no
// user, so without it a reclaim ticket for an admin's seat would
// confer admin. With it, only a session minted from a real Discord
// sign-in qualifies, whatever its role: identified, player or
// spectator.
func isAllowlisted(a *Admins, p auth.Principal) bool {
	if a == nil {
		return false
	}
	return p.UserID != uuid.Nil && p.DiscordID != "" && a.list.Has(p.DiscordID)
}

// isAdminPrincipal is isAdmin with the Admins passed in, for the
// callers that hold one but no Config (WSAuthorizer): the shared token,
// or an allowlisted person in admin mode. Any failure to know the mode
// (no user database, a failed boot load, a lapse) reads as player mode.
func isAdminPrincipal(a *Admins, p auth.Principal) bool {
	if isServerCredential(p) {
		return true
	}
	if !isAllowlisted(a, p) || a.modes == nil {
		return false
	}
	return a.modes.On(p.UserID, a.modes.Now())
}

// isAdmin reports whether p is an operator: the shared token, or a
// signed-in person on the allowlist who is in admin mode. Computed per
// request; nothing about it is stored in a token, so a switch reaches
// every session the person holds at its next request.
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

// --- admin mode (ADR 0112 §2) ----------------------------------------

// AdminModeRebinder closes the WebSockets whose admin bit no longer
// matches the person's answer, so they reconnect and are bound again.
// *ws.Hub implements it.
type AdminModeRebinder interface {
	RebindUserSessions(userID uuid.UUID, admin bool) int
}

// adminModeRequest is PUT /me/admin-mode's body. On is a pointer so a
// missing field is a 400, never a silent "off".
type adminModeRequest struct {
	On *bool `json:"on"`
}

// adminModeResponse is PUT /me/admin-mode's answer. Admin is the
// effective answer (isAdmin); AdminMode is whether this person has
// admin mode on; AdminModeEndsAt is when it lapses, in Unix
// milliseconds, present only while it is on.
type adminModeResponse struct {
	Admin           bool  `json:"admin"`
	AdminMode       bool  `json:"admin_mode"`
	AdminModeEndsAt int64 `json:"admin_mode_ends_at,omitempty"`
}

// adminModeOf is p's admin-mode state for a response.
func (c Config) adminModeOf(p auth.Principal) adminModeResponse {
	out := adminModeResponse{Admin: c.isAdmin(p)}
	if !isAllowlisted(c.Admins, p) {
		return out
	}
	if ends, on := c.Admins.modeEndsAt(p.UserID); on {
		out.AdminMode = true
		out.AdminModeEndsAt = ends.UnixMilli()
	}
	return out
}

// putAdminMode is PUT /me/admin-mode (ADR 0112 §2 item 3): an
// allowlisted person switches admin mode on (for 12 hours, restarting
// the 12 if it was already on) or off. Anyone else, the shared token
// included (it has no mode), is a 403 "not an admin". The switch is
// written to the database, then the cache, and then the person's
// sockets whose admin bit is now wrong are closed with 4001 so they
// reconnect with the new answer.
//
// A failed write is a 500, and leaves the person in the safer mode: a
// failed switch on stays off, and a failed switch off is off in this
// process anyway (users.AdminModes.Set).
func putAdminMode(c Config, w http.ResponseWriter, r *http.Request) error {
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		return httpError(http.StatusInternalServerError, "missing principal")
	}
	if !isSignedInPerson(p) || !isAllowlisted(c.Admins, p) {
		return httpError(http.StatusForbidden, "not an admin")
	}
	var body adminModeRequest
	if err := decodeJSON(w, r, &body); err != nil {
		return err
	}
	if body.On == nil {
		return httpError(http.StatusBadRequest, `body must be {"on": true} or {"on": false}`)
	}
	if c.Admins.modes == nil {
		return httpError(http.StatusServiceUnavailable, "admin mode needs the user database, and this server has none")
	}
	_, err := c.Admins.modes.Set(r.Context(), p.UserID, *body.On)
	out := c.adminModeOf(p)
	closed := 0
	if c.AdminSockets != nil {
		closed = c.AdminSockets.RebindUserSessions(p.UserID, out.Admin)
	}
	if err != nil {
		c.logger().Error("admin mode switch failed", "admin_user_id", p.UserID.String(), "on", *body.On, "err", err)
		if errors.Is(err, users.ErrNotFound) {
			return httpError(http.StatusNotFound, "user not found")
		}
		if *body.On {
			return httpError(http.StatusInternalServerError, "could not switch admin mode on; you are still in player mode")
		}
		return httpError(http.StatusInternalServerError, "could not save player mode; it holds until the server restarts, so switch it off again then")
	}
	action := "admin mode off"
	if out.AdminMode {
		action = "admin mode on"
	}
	if c.Log != nil {
		c.Log.Info(action, "admin_user_id", p.UserID.String(), "ends_at", out.AdminModeEndsAt, "sockets_closed", closed)
	}
	return writeJSON(w, http.StatusOK, out)
}

// AdminModeSweepInterval is how often RunAdminModeSweeper looks for
// lapsed admin modes (ADR 0112 §2 item 1, owner answer 1).
const AdminModeSweepInterval = time.Minute

// RunAdminModeSweeper ends lapsed admin modes until ctx is done: once
// at start (a lapse found while the server was down) and then every
// interval. HTTP already reads a lapsed mode as off at the next
// request; the sweep is what clears the row and closes the person's
// admin sockets with 4001, so the admin menu goes away at the table
// too. Each lapse is logged as "admin mode lapsed" with admin_user_id.
// interval <= 0 means AdminModeSweepInterval.
func RunAdminModeSweeper(ctx context.Context, a *Admins, sockets AdminModeRebinder, log *slog.Logger, interval time.Duration) {
	if a == nil || a.modes == nil {
		return
	}
	if interval <= 0 {
		interval = AdminModeSweepInterval
	}
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		sweepAdminModes(ctx, a, sockets, log)
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
	}
}

// sweepAdminModes is one tick of RunAdminModeSweeper.
func sweepAdminModes(ctx context.Context, a *Admins, sockets AdminModeRebinder, log *slog.Logger) {
	ended, err := a.modes.Sweep(ctx, a.modes.Now())
	if err != nil && log != nil {
		log.Error("admin mode sweep could not clear every lapsed row; they read as player mode anyway", "err", err)
	}
	for _, id := range ended {
		closed := 0
		if sockets != nil {
			// The person's answer now, not a bare false: a switch made
			// since the sweep read the row must not be undone here.
			_, on := a.modeEndsAt(id)
			closed = sockets.RebindUserSessions(id, on)
		}
		if log != nil {
			log.Info("admin mode lapsed", "admin_user_id", id.String(), "sockets_closed", closed)
		}
	}
}
