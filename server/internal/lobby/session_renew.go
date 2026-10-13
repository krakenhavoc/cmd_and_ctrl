package lobby

// session_renew.go is POST /me/session, ADR 0110 §1 items 5 and 6:
// renewal on use (owner answer 1) and the reinstall of a saved
// identity.
//
// The route does two things with one credential:
//
//   - It sets the session cookie to the credential it was handed, and
//     returns that session. This is the reinstall: a client that kept a
//     signed-in session aside while it held one with no user (the admin
//     token, a reclaim ticket for someone else's seat) sends the saved
//     token as its bearer once the other session has expired, and the
//     browser is back on the person's own session.
//   - When that session is more than half spent, it is re-issued first,
//     with a fresh IdentityTTL. A person who uses the site at least once
//     in every half-lifetime is never asked to sign in again.
//
// With the sign-up gate on (ADR 0156 §1 item 5), a renewal first asks
// Discord again whether the person holds the role (renewalGate): a
// refusal withdraws every session they hold and answers 401; no refresh
// token to ask with answers the session as it is, not renewed; an
// outage renews anyway, logged at WARN. An allowlisted admin is exempt.
//
// What a renewal keeps and what it changes: the principal is copied
// whole, so the role, the user, the game and seat (or the watched
// game), the seat label and the cached Discord fields are the same.
// Only IssuedAt and ExpiresAt are new. The old token is not revoked:
// under HMAC sessions Revoke is advisory, and another tab may still be
// holding it, so it simply runs out at its own expiry.
//
// Who may call it: a signed-in person only (isSignedInPerson), so a
// session that cannot be revoked is never renewed or reinstalled. A
// revoked token fails Validate before the handler runs, and a
// revocation that lands while a renewal is being minted is caught by
// validating the presented token again after the mint.

import (
	"errors"
	"net/http"
	"time"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
)

// sessionRenewRoute is the route's auth wrapper: auth.Middleware, but
// reading only the cookie and the Authorization header. The route sets
// the cookie from the credential it reads, and a ?token= is something
// a cross-site form can supply (auth.CredentialFromCookieOrHeader).
func sessionRenewRoute(c Config, next http.Handler) http.Handler {
	return auth.MiddlewareWith(c.Auth, auth.CredentialFromCookieOrHeader)(next)
}

// halfSpent reports whether more than half of p's own lifetime, from
// its IssuedAt to its ExpiresAt, has passed at now. The client makes
// the same test on the same two fields to decide when to call
// (client/src/lib/identity.ts), so the two never disagree about a
// session's half-life.
func halfSpent(p auth.Principal, now time.Time) bool {
	life := p.ExpiresAt.Sub(p.IssuedAt)
	return now.Sub(p.IssuedAt) > life/2
}

// renewSession is POST /me/session. No body. The response is the
// sessionResponse the other mints return (no embedded game), with the
// cookie set to the token it carries.
func renewSession(c Config, w http.ResponseWriter, r *http.Request) error {
	p, err := signedInUser(r)
	if err != nil {
		return err
	}
	cred := auth.CredentialFromCookieOrHeader(r)
	tok, out := cred, p
	spent := halfSpent(p, c.now())
	verdict := renewalGo
	if spent {
		// The sign-up gate's re-check (ADR 0156 §1 item 5). Off, or for
		// an exempt admin, it is renewalGo without a call.
		verdict = c.renewalGate(r.Context(), p)
	}
	if verdict == renewalRefused {
		// Every session this person holds ends, this one included: the
		// role is gone, so the account is closed to them until a mod
		// gives it back and they sign in again.
		if _, err := revokeUser(r.Context(), c, p.UserID); err != nil {
			return err
		}
		clearSessionCookie(c, w)
		return httpError(http.StatusUnauthorized, "session revoked: your Discord account no longer has the role this site needs")
	}
	if spent && verdict == renewalGo {
		renewed := p
		renewed.IssuedAt, renewed.ExpiresAt = time.Time{}, time.Time{}
		// A UserID and no source session: IdentityTTL (issueFor).
		tok, out, err = issueFor(r.Context(), c, renewed, nil)
		if err != nil {
			return err
		}
		// A revocation between the Validate in front of this handler
		// and the mint above would not reach the new token, whose
		// IssuedAt is later than the revocation. Asking about the
		// presented token again closes that window: if it is revoked
		// now, the new one is never handed out.
		if _, err := c.Auth.Validate(r.Context(), cred); err != nil {
			if errors.Is(err, auth.ErrRevokedCredential) {
				return httpError(http.StatusUnauthorized, "session revoked")
			}
			return httpError(http.StatusUnauthorized, "session expired")
		}
		c.logger().Info("session renewed", "user_id", out.UserID, "role", out.Role, "expires_at", out.ExpiresAt)
	}
	setSessionCookie(c, w, tok, out.ExpiresAt)
	return writeJSON(w, http.StatusOK, sessionResponse{
		Token:     tok,
		ExpiresAt: out.ExpiresAt,
		Principal: out,
		PlayerID:  out.PlayerID,
	})
}

// now is the lobby's clock: Config.Now, or the wall clock. The
// renewal's half-life test reads it, and so does the time an access
// check is recorded at (ADR 0156 §1 item 7); the authenticator keeps
// its own.
func (c Config) now() time.Time {
	if c.Now != nil {
		return c.Now()
	}
	return time.Now()
}
