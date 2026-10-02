package lobby

// session_ttl.go is ADR 0110 §1's one rule for how long a minted
// session lives. Every session the lobby mints goes through issueFor,
// and nothing in this package calls Authenticator.Issue directly
// (TestOnlyIssueForMintsSessions holds that line).

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
)

// issueFor mints p's session and returns the token and the issued
// principal, as Authenticator.Issue does. The lifetime is chosen here,
// and only here:
//
//   - p carries no UserID: SessionTTL (12 hours by default). Guest
//     seats, guest spectators, admin-token sessions and reclaim-ticket
//     sessions. They cannot be revoked (ADR 0051 decision 6), so they
//     stay short.
//   - p carries a UserID and source is nil: IdentityTTL (30 days by
//     default). Those are the Discord callback's mints, the sign-in
//     itself, and POST /me/session's renewal of a session more than
//     half spent (session_renew.go, ADR 0110 §1 item 5), which is the
//     only other way a sign-in's lifetime is extended.
//   - p carries a UserID and source is the signed-in session the caller
//     presented: source's own ExpiresAt, to the millisecond. A seat
//     claimed from a sign-in with three days left also has three days
//     left, so joining a table never extends a sign-in. The new token
//     still has a fresh IssuedAt, so sessions_invalid_before treats it
//     like any other.
//
// The identity-only session (RoleIdentified) with no source keeps
// IdentityTTL even with no UserID: on a deployment with no database
// that is the session ADR 0051 decision 3 already gave the long life,
// and ADR 0110 does not shorten it.
//
// A source whose user is not p's is a caller bug and is refused.
func issueFor(ctx context.Context, c Config, p auth.Principal, source *auth.Principal) (string, auth.Principal, error) {
	if p.UserID == uuid.Nil {
		if source == nil && p.Role == auth.RoleIdentified {
			return c.Auth.Issue(ctx, p, c.IdentityTTL)
		}
		return c.Auth.Issue(ctx, p, c.SessionTTL)
	}
	if source == nil {
		return c.Auth.Issue(ctx, p, c.IdentityTTL)
	}
	if source.UserID != p.UserID {
		return "", auth.Principal{}, fmt.Errorf("lobby: minting a session for user %s from a session of user %s", p.UserID, source.UserID)
	}
	tok, issued, err := c.Auth.IssueUntil(ctx, p, source.ExpiresAt)
	if errors.Is(err, auth.ErrExpiredCredential) {
		// The source validated a moment ago and has run out since.
		return "", auth.Principal{}, httpError(http.StatusUnauthorized, "session expired")
	}
	return tok, issued, err
}

// isSignedInPerson reports whether p is a person signed in with
// Discord: a session with a UserID whose role is a person's. Identity,
// seat and spectator sessions qualify (ADR 0110 §1 item 2 added the
// spectator). An admin-token session is a server credential, not a
// person, whatever it carries.
func isSignedInPerson(p auth.Principal) bool {
	if p.UserID == uuid.Nil {
		return false
	}
	switch p.Role {
	case auth.RoleIdentified, auth.RolePlayer, auth.RoleSpectator:
		return true
	}
	return false
}

// optionalSession validates the credential on a request that does not
// require one (the join, spectate and reclaim routes). A missing,
// expired, revoked or bogus credential is (zero, false): those routes
// then go ahead exactly as they would for a stranger, rather than
// 401-ing someone whose old session lapsed while they found the link.
func optionalSession(c Config, r *http.Request) (auth.Principal, bool) {
	cred := auth.CredentialFromRequest(r)
	if cred == "" {
		return auth.Principal{}, false
	}
	p, err := c.Auth.Validate(r.Context(), cred)
	if err != nil {
		return auth.Principal{}, false
	}
	return p, true
}

// withIdentity copies a Discord identity onto a seat or spectator
// principal and names it after the Discord display name. A zero
// identity leaves p alone.
func withIdentity(p auth.Principal, identity DiscordIdentity) auth.Principal {
	if !identity.Populated() {
		return p
	}
	p.Name = identity.DisplayName()
	p.DiscordID = identity.ID
	p.DiscordUsername = identity.Username
	p.DiscordGlobalName = identity.GlobalName
	p.DiscordAvatarHash = identity.AvatarHash
	return p
}
