package lobby

// practice_http.go is the HTTP half of the tutorial's practice table
// (practice.go, ADR 0076 §2.2):
//
//	POST /games/practice              — any session: open one, seated
//	POST /games/{id}/practice/leave   — the practice token: abandon it
//
// # The session swap, and putting it back
//
// Opening a practice table mints a player session for its human seat,
// exactly as joining any table does, and sets the session cookie to
// it. That cookie is the one thing the client cannot put back by
// itself: the server reads the cookie BEFORE the Authorization header
// (auth.CredentialFromRequest), so a client that restored its previous
// session from storage would still be talking as the practice seat.
// So leaving takes the token to restore and re-issues the cookie for
// it — the server-side half of "the player's own values are written
// back on exit" (§2.2), for the session rather than the settings.
//
// Leave is unauthenticated on purpose and carries both tokens in its
// body. It is called from the page's way out — a closed tab, a
// navigation — and again on the next page load if that call never
// landed, and by then the cookie may hold either token or a dead one.
// Reading the tokens from the body makes the outcome the same in every
// one of those states. It only accepts application/json, which a
// cross-site form cannot send, so another site cannot use it to plant
// a session cookie in this one.

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/decks"
)

// practiceBotTier is the practice bot's tier. ADR 0076 §2.2: `random`,
// with the slow decklist carrying the difficulty, rather than a
// dedicated tutorial tier.
const practiceBotTier = "random"

// practiceBotName is the bot seat's label.
const practiceBotName = "Practice Bot"

// practiceLeaveRequest is POST /games/{id}/practice/leave's body.
type practiceLeaveRequest struct {
	// PracticeToken is the practice seat's session: the credential
	// for leaving. Empty or no longer valid skips the leave (the
	// reaper will have it) and still restores the cookie.
	PracticeToken string `json:"practice_token"`
	// RestoreToken is the session to put back in the cookie. Empty or
	// no longer valid clears the cookie instead.
	RestoreToken string `json:"restore_token,omitempty"`
}

// createPractice handles POST /games/practice.
func createPractice(c Config, w http.ResponseWriter, r *http.Request) error {
	if c.Bots == nil {
		return httpError(http.StatusServiceUnavailable, "bot seats are not enabled on this server")
	}
	if c.Cards == nil || c.Cards.Count() == 0 {
		return httpError(http.StatusServiceUnavailable, "card index not loaded; run scripts/scryfall-refresh.sh")
	}
	offered := false
	for _, t := range c.Bots.Tiers() {
		if t == practiceBotTier {
			offered = true
			break
		}
	}
	if !offered {
		return httpError(http.StatusServiceUnavailable, "the practice bot's tier is not available on this server")
	}
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		return httpError(http.StatusInternalServerError, "missing principal")
	}

	// The decks run the same parse → resolve → validate pipeline as a
	// deck a player uploads (decks.Deck.Load). A failure here is the
	// server's card data, not the caller's request.
	playerDeck, botDeck := decks.TutorialPlayer(), decks.TutorialBot()
	humanList, err := playerDeck.Load(c.Cards)
	if err != nil {
		return httpError(http.StatusInternalServerError, fmt.Sprintf("practice deck: %v", err))
	}
	botList, err := botDeck.Load(c.Cards)
	if err != nil {
		return httpError(http.StatusInternalServerError, fmt.Sprintf("practice bot deck: %v", err))
	}

	identity := principalIdentity(p)
	name := strings.TrimSpace(p.Name)
	if p.Role != auth.RolePlayer && p.Role != auth.RoleIdentified {
		// "admin" and a spectator's watch label are not a seat name.
		name = ""
	}
	if name == "" {
		name = "Player"
	}
	meta, playerID, err := c.Lobby.CreatePractice(PracticeHuman{
		Name:     name,
		Identity: identity,
		UserID:   p.UserID,
		Owner:    practiceOwner(c, p),
		DeckName: playerDeck.Name,
		Cards:    humanList.ToGameCards(),
	}, PracticeBot{
		Name:     practiceBotName,
		Tier:     practiceBotTier,
		DeckID:   botDeck.ID,
		DeckName: botDeck.Name,
		Cards:    botList.ToGameCards(),
	})
	if errors.Is(err, ErrPracticeTablesFull) {
		return httpError(http.StatusServiceUnavailable, "every practice table is in use; try again in a few minutes")
	}
	if err != nil {
		return err
	}

	// The seat's session, minted like a join's: bound to this game and
	// seat, and carrying the caller's user and Discord identity so the
	// table shows their avatar and a revocation still reaches it. A
	// signed-in caller's seat expires with the session that opened it
	// (ADR 0110 §1); anyone else's gets SessionTTL.
	seat := withIdentity(auth.Principal{
		Role:     auth.RolePlayer,
		UserID:   p.UserID,
		GameID:   meta.ID,
		PlayerID: playerID,
		Name:     name,
	}, identity)
	tok, issued, err := issueFor(r.Context(), c, seat, &p)
	if err != nil {
		return err
	}
	setSessionCookie(c, w, tok, issued.ExpiresAt)
	return writeJSON(w, http.StatusCreated, sessionResponse{
		Token:     tok,
		ExpiresAt: issued.ExpiresAt,
		Principal: issued,
		Game:      &meta,
		PlayerID:  playerID,
	})
}

// leavePractice handles POST /games/{id}/practice/leave. Always 204
// for a well-formed request: leaving a table that is already gone, or
// with a practice token that no longer validates, is not an error —
// the page's way out and the next page load may both call it.
func leavePractice(c Config, w http.ResponseWriter, r *http.Request) error {
	if ct := r.Header.Get("Content-Type"); !strings.HasPrefix(strings.ToLower(strings.TrimSpace(ct)), "application/json") {
		return httpError(http.StatusUnsupportedMediaType, "send the body as application/json")
	}
	id, err := gameIDFromPath(r)
	if err != nil {
		return err
	}
	var body practiceLeaveRequest
	if err := decodeJSON(w, r, &body); err != nil {
		return err
	}

	if tok := strings.TrimSpace(body.PracticeToken); tok != "" {
		if p, verr := c.Auth.Validate(r.Context(), tok); verr == nil && p.Role == auth.RolePlayer && p.GameID == id {
			switch err := c.Lobby.LeavePractice(id, p.PlayerID); {
			case err == nil, errors.Is(err, ErrGameNotFound):
			case errors.Is(err, ErrNotPracticeTable):
				return httpError(http.StatusConflict, "not a practice table")
			default:
				return err
			}
		}
	}

	if tok := strings.TrimSpace(body.RestoreToken); tok != "" {
		if p, verr := c.Auth.Validate(r.Context(), tok); verr == nil {
			setSessionCookie(c, w, tok, p.ExpiresAt)
			w.WriteHeader(http.StatusNoContent)
			return nil
		}
	}
	clearSessionCookie(c, w)
	w.WriteHeader(http.StatusNoContent)
	return nil
}

// principalIdentity is the Discord identity a session carries, or the
// zero identity.
func principalIdentity(p auth.Principal) DiscordIdentity {
	if p.DiscordID == "" {
		return DiscordIdentity{}
	}
	return DiscordIdentity{
		ID:         p.DiscordID,
		Username:   p.DiscordUsername,
		GlobalName: p.DiscordGlobalName,
		AvatarHash: p.DiscordAvatarHash,
	}
}

// practiceOwner is who a practice table belongs to for the
// one-per-person rule: the user, else the Discord identity, else the
// admin credential, else the seat. A session that is itself a practice
// seat belongs to whoever opened that table. "" — a spectator with no
// identity — opts out of the rule; the table ceiling still applies.
func practiceOwner(c Config, p auth.Principal) string {
	if p.Role == auth.RolePlayer && p.GameID != uuid.Nil {
		if owner, ok := c.Lobby.PracticeOwner(p.GameID); ok {
			return owner
		}
	}
	switch {
	case p.UserID != uuid.Nil:
		return "user:" + p.UserID.String()
	case p.DiscordID != "":
		return "discord:" + p.DiscordID
	case p.Role == auth.RoleAdmin:
		return "admin"
	case p.Role == auth.RolePlayer && p.PlayerID != uuid.Nil:
		return "seat:" + p.PlayerID.String()
	default:
		return ""
	}
}
