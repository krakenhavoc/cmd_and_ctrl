package lobby

// mygames_http.go is the HTTP half of mygames.go: GET /me/games, POST
// /me/games/{id}/session, and the helper the join routes use to seat a
// signed-in person as themselves. ADR 0051 sub-PR 4.

import (
	"net/http"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
)

// myGamesResponse is the body of GET /me/games.
type myGamesResponse struct {
	Games []MyGame `json:"games"`
}

// signedInUser returns the principal on the request if it is a person:
// an identity session or a player session that carries a UserID. Admin
// sessions are a server credential, spectator and guest sessions are
// nobody in particular, and a deployment with no database has no
// users, so every one of those is a 401 on the /me/games routes.
func signedInUser(r *http.Request) (auth.Principal, error) {
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		return auth.Principal{}, httpError(http.StatusInternalServerError, "missing principal")
	}
	if p.UserID == uuid.Nil || (p.Role != auth.RoleIdentified && p.Role != auth.RolePlayer) {
		return auth.Principal{}, httpError(http.StatusUnauthorized, "sign in with Discord to see your games")
	}
	return p, nil
}

// myGames is GET /me/games: every seat the caller's user holds, newest
// game first, ended and archived included. It never carries an invite
// token; the way back into an open table is Rejoin, which needs the
// caller's own session.
func myGames(c Config, w http.ResponseWriter, r *http.Request) error {
	p, err := signedInUser(r)
	if err != nil {
		return err
	}
	games, err := c.Lobby.MyGames(p.UserID)
	if err != nil {
		c.logger().Error("listing a user's games failed", "err", err)
		return httpError(http.StatusInternalServerError, "could not load your games; try again")
	}
	return writeJSON(w, http.StatusOK, myGamesResponse{Games: games})
}

// myGameSession is POST /me/games/{id}/session: seat reclaim by user
// (ADR 0051 decision 3). A signed-in person gets a fresh player session
// for the seat whose user_id is theirs, at any live table, without the
// invite link or an admin-minted ticket. The ticket route stays the
// backstop for guests, who have no identity to prove.
//
// No body. The response is the same sessionResponse /join returns,
// with both invite tokens stripped: the caller proved they hold one
// seat, not that they host the table.
func myGameSession(c Config, w http.ResponseWriter, r *http.Request) error {
	p, err := signedInUser(r)
	if err != nil {
		return err
	}
	id, err := gameIDFromPath(r)
	if err != nil {
		return err
	}
	meta, seat, err := c.Lobby.ReclaimByUser(id, p.UserID)
	if err != nil {
		return err
	}
	np := auth.Principal{
		Role:              auth.RolePlayer,
		UserID:            p.UserID,
		GameID:            meta.ID,
		PlayerID:          seat.PlayerID,
		Name:              seatLabel(seat),
		DiscordID:         seat.DiscordID,
		DiscordGlobalName: seat.DisplayName,
		DiscordAvatarHash: seat.DiscordAvatarHash,
	}
	// The username is not on the seat; carry it over when the caller's
	// session is the same Discord account.
	if p.DiscordID != "" && p.DiscordID == seat.DiscordID {
		np.DiscordUsername = p.DiscordUsername
	}
	tok, issued, err := c.Auth.Issue(r.Context(), np, c.SessionTTL)
	if err != nil {
		return err
	}
	setSessionCookie(c, w, tok, issued.ExpiresAt)
	meta.InviteToken = ""
	meta.SpectatorInvite = ""
	return writeJSON(w, http.StatusOK, sessionResponse{
		Token:     tok,
		ExpiresAt: issued.ExpiresAt,
		Principal: issued,
		Game:      &meta,
		PlayerID:  seat.PlayerID,
	})
}

// signedInIdentity reads an OPTIONAL session off a join request and
// returns the Discord identity and user to seat, if it belongs to a
// signed-in person: an identity session, or a player session from a
// Discord sign-in (the same person at their next table). Anything
// else, including a credential that does not validate, is the zero
// identity, and the join goes ahead by name exactly as before.
func signedInIdentity(c Config, r *http.Request) (DiscordIdentity, uuid.UUID) {
	cred := auth.CredentialFromRequest(r)
	if cred == "" {
		return DiscordIdentity{}, uuid.Nil
	}
	p, err := c.Auth.Validate(r.Context(), cred)
	if err != nil || p.DiscordID == "" {
		return DiscordIdentity{}, uuid.Nil
	}
	switch {
	case p.Role == auth.RoleIdentified:
	case p.Role == auth.RolePlayer && p.UserID != uuid.Nil:
	default:
		return DiscordIdentity{}, uuid.Nil
	}
	return DiscordIdentity{
		ID:         p.DiscordID,
		Username:   p.DiscordUsername,
		GlobalName: p.DiscordGlobalName,
		AvatarHash: p.DiscordAvatarHash,
	}, p.UserID
}
