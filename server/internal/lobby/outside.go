package lobby

import (
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// outside.go — the search behind "a card you own from outside the game"
// when the chooser's pool is the card database (ADR 0198 §3).
//
// The prompt itself rides the WebSocket like every other choice and is
// answered there with resolve_choice. Only the SEARCH is HTTP, for the
// reason the spawner's is: it is a read of the card index, one request
// per debounced keystroke, and nothing about it mutates the game.
//
// It is not a general card search. It answers only the seat that owes the
// open collection pick it names, and only with cards that pick would
// accept, so it cannot be used to browse the database outside a wish or
// by anyone but the wisher.

// OutsideCardsRoute is the search route.
const OutsideCardsRoute = "GET /games/{id}/outside/cards"

// outsideCardResult is one row of the search: the dev spawner's row
// shape (devCardResult), so the client renders both with one component.
type outsideCardResult struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	TypeLine string `json:"type_line"`
	ManaCost string `json:"mana_cost"`
	SetCode  string `json:"set"`
}

// outsideCardSearch serves GET /games/{id}/outside/cards?choice=&q=.
func outsideCardSearch(c Config, w http.ResponseWriter, r *http.Request) error {
	id, err := gameIDFromPath(r)
	if err != nil {
		return err
	}
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		return httpError(http.StatusInternalServerError, "missing principal")
	}
	// The pick is a seat's own: no spectator, no admin bypass.
	if p.Role == auth.RoleSpectator || p.PlayerID == uuid.Nil || p.GameID != id {
		return httpError(http.StatusForbidden, "not a seat in this game")
	}
	choiceID, err := uuid.Parse(strings.TrimSpace(r.URL.Query().Get("choice")))
	if err != nil {
		return httpError(http.StatusBadRequest, "choice must be a uuid")
	}
	g, err := c.Lobby.LookupGame(id)
	if err != nil || g == nil {
		return httpError(http.StatusNotFound, "game not found")
	}
	found, err := g.SearchOutsideCollection(choiceID, p.PlayerID, r.URL.Query().Get("q"), 0)
	if errors.Is(err, game.ErrNotACollectionPick) {
		return httpError(http.StatusNotFound, "no open outside-the-game pick for this seat")
	}
	if err != nil {
		return err
	}
	out := make([]outsideCardResult, 0, len(found))
	for _, card := range found {
		row := outsideCardResult{
			ID:       card.ScryfallID,
			Name:     card.Name,
			TypeLine: card.TypeLine,
			ManaCost: card.ManaCost,
		}
		if pid, err := uuid.Parse(card.ScryfallID); err == nil && c.Cards != nil {
			if printing, ok := c.Cards.Get(pid); ok {
				row.SetCode = printing.SetCode
			}
		}
		out = append(out, row)
	}
	return writeJSON(w, http.StatusOK, map[string][]outsideCardResult{"cards": out})
}
