package lobby

// PUT /admin/decks/{id}/power (#3671): rate a saved deck's power level.
//
// The ratings are made by hand, on demand (an owner reads the saved
// decks and rates them on the Commander Bracket scale), not by the
// server, so the only write is this admin route. The seat line at a
// table shows the rating of every seated library deck.

import (
	"errors"
	"net/http"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/decklibrary"
)

// deckPowerRequest is PUT /admin/decks/{id}/power's body. PowerLevel is
// 1 to 5 in half steps; 0 or null clears the rating.
type deckPowerRequest struct {
	PowerLevel *float64 `json:"power_level"`
}

// deckPowerResponse is the rated deck. PowerLevel 0 is unrated.
type deckPowerResponse struct {
	DeckID     string  `json:"deck_id"`
	Name       string  `json:"name"`
	PowerLevel float64 `json:"power_level"`
}

func adminSetDeckPower(c Config, w http.ResponseWriter, r *http.Request) error {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil || id == uuid.Nil {
		return httpError(http.StatusBadRequest, "invalid deck id")
	}
	var body deckPowerRequest
	if err := decodeJSON(w, r, &body); err != nil {
		return err
	}
	var level float64
	if body.PowerLevel != nil {
		level = *body.PowerLevel
	}
	if c.DeckLibrary == nil {
		return httpError(http.StatusServiceUnavailable, "deck library not configured")
	}
	d, err := c.DeckLibrary.SetPowerLevel(r.Context(), id, level)
	switch {
	case errors.Is(err, decklibrary.ErrInvalidPowerLevel):
		return httpError(http.StatusBadRequest, "power_level must be 1 to 5 in half steps, or 0 or null to clear it")
	case errors.Is(err, decklibrary.ErrNotFound):
		return httpError(http.StatusNotFound, "deck not found")
	case err != nil:
		return err
	}
	if c.Lobby != nil {
		c.Lobby.SetLibraryDeckPower(d.ID.String(), d.PowerLevel)
	}
	return writeJSON(w, http.StatusOK, deckPowerResponse{DeckID: d.ID.String(), Name: d.Name, PowerLevel: d.PowerLevel})
}
