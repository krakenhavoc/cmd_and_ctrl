package lobby

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/users"
)

// SessionRevoker withdraws every session a user holds (ADR 0051
// decision 6). *users.Revocations implements it. It returns the
// watermark it stored; a session issued at or before it is refused.
type SessionRevoker interface {
	RevokeAll(ctx context.Context, userID uuid.UUID) (time.Time, error)
}

// UserSessionEvictor closes the WebSockets a revoked user still has
// open. *ws.Hub implements it.
type UserSessionEvictor interface {
	EvictUserSessions(userID uuid.UUID, before time.Time) int
}

// revokeSessionsResponse is the admin route's answer.
type revokeSessionsResponse struct {
	UserID                uuid.UUID `json:"user_id"`
	SessionsInvalidBefore time.Time `json:"sessions_invalid_before"`
	// SocketsClosed is how many open WebSockets were closed. Zero
	// when no evictor is wired.
	SocketsClosed int `json:"sockets_closed"`
}

// revokeUser runs one revocation: the watermark, then the sockets.
func revokeUser(ctx context.Context, c Config, userID uuid.UUID) (revokeSessionsResponse, error) {
	if c.Revocations == nil {
		return revokeSessionsResponse{}, httpError(http.StatusServiceUnavailable,
			"session revocation needs the user database, and this server has none")
	}
	at, err := c.Revocations.RevokeAll(ctx, userID)
	if errors.Is(err, users.ErrNotFound) {
		return revokeSessionsResponse{}, httpError(http.StatusNotFound, "user not found")
	}
	if err != nil {
		if c.Log != nil {
			c.Log.Error("revoke sessions failed", "user_id", userID, "err", err)
		}
		return revokeSessionsResponse{}, httpError(http.StatusInternalServerError, "could not revoke sessions; try again")
	}
	out := revokeSessionsResponse{UserID: userID, SessionsInvalidBefore: at}
	if c.SessionEvictor != nil {
		out.SocketsClosed = c.SessionEvictor.EvictUserSessions(userID, at)
	}
	if c.Log != nil {
		c.Log.Info("user sessions revoked", "user_id", userID, "invalid_before", at, "sockets_closed", out.SocketsClosed)
	}
	return out, nil
}

// logoutEverywhere is POST /logout/everywhere: the caller's user
// withdraws every session it holds, in every browser, including the
// one making this request, and that browser's cookie is cleared. 204.
//
// It needs a valid session, unlike POST /logout: it acts on the user
// the session names, and a session that no longer validates has
// nothing left to sign out of. A session with no user (admin, guest,
// spectator, or any session on a server with no database) is refused
// with 403. Those have no watermark to move, and POST /logout is their
// sign-out.
func logoutEverywhere(c Config, w http.ResponseWriter, r *http.Request) error {
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		return httpError(http.StatusInternalServerError, "missing principal")
	}
	if p.UserID == uuid.Nil {
		return httpError(http.StatusForbidden,
			"this session is not tied to a signed-in user; use POST /logout")
	}
	if _, err := revokeUser(r.Context(), c, p.UserID); err != nil {
		return err
	}
	clearSessionCookie(c, w)
	w.WriteHeader(http.StatusNoContent)
	return nil
}

// adminRevokeUserSessions is POST /admin/users/{id}/revoke-sessions,
// ADR 0051's "admin remove-user". The ADR scopes removal to revocation
// and nothing more, and the route is named for what it does: the
// user's sessions die, open sockets close, and every row (the user,
// identities, seats, games, decks) stays. The person can sign in again
// with Discord and gets a fresh session. Keeping someone out for good
// would need a flag checked at sign-in, and that is not part of
// decision 6.
func adminRevokeUserSessions(c Config, w http.ResponseWriter, r *http.Request) error {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil || id == uuid.Nil {
		return httpError(http.StatusBadRequest, "invalid user id")
	}
	out, err := revokeUser(r.Context(), c, id)
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, out)
}
