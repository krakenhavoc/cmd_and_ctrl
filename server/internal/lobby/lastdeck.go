package lobby

// lastdeck.go is ADR 0110 section 5 item 5 (Delivery PR 7): the deck a
// signed-in person last seated, users.last_deck (migration 0008).
//
// It is written whenever a signed-in person seats a deck at their own
// seat: a library deck, an upload that was saved to the library, or a
// pre-built deck. The lobby's deck panel reads it from GET
// /me/last-deck and PRESELECTS it; nothing is ever seated
// automatically, because a seated deck is visible to the table and a
// stale choice should cost a click, not a mulligan. A deck that has
// since gone (deleted from the library, or a pre-built one retired) is
// simply not preselected; the server does not chase it.
//
// Guests keep the same preselection in their browser
// (localStorage["cmdctrl.lastDeck"], pre-built decks only).

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/users"
)

// lastDeckResponse is the body of GET /me/last-deck. LastDeck is null
// when nothing has been recorded.
type lastDeckResponse struct {
	LastDeck *users.LastDeck `json:"last_deck"`
}

// recordLastDeck stores d as p's last deck, when p is a signed-in
// person seating THEIR OWN seat (gameID, playerID). An admin installing
// somebody else's deck is not choosing a deck for themselves, so it is
// not recorded. A failure is logged and never fails the seating.
func recordLastDeck(ctx context.Context, c Config, p auth.Principal, gameID, playerID uuid.UUID, d users.LastDeck) {
	if p.UserID == uuid.Nil || p.GameID != gameID || p.PlayerID != playerID {
		return
	}
	if err := c.userStore().SetLastDeck(ctx, p.UserID, d); err != nil && !errors.Is(err, users.ErrNotFound) {
		c.logger().Warn("users: recording the last deck failed", "err", err, "user_id", p.UserID)
	}
}

// myLastDeck is GET /me/last-deck. Same caller rule as the rest of
// /me/*: a signed-in person, else 403 (never 401, #1154).
func myLastDeck(c Config, w http.ResponseWriter, r *http.Request) error {
	p, err := signedInUser(r)
	if err != nil {
		return err
	}
	d, err := c.userStore().LastDeck(r.Context(), p.UserID)
	if errors.Is(err, users.ErrNotFound) || (err == nil && d.IsZero()) {
		return writeJSON(w, http.StatusOK, lastDeckResponse{})
	}
	if err != nil {
		c.logger().Error("reading a user's last deck failed", "err", err)
		return httpError(http.StatusInternalServerError, "could not load your last deck; try again")
	}
	return writeJSON(w, http.StatusOK, lastDeckResponse{LastDeck: &d})
}
