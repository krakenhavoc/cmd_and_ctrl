package legal

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// revealPayment is the one card a reveal / behold branch names (ADR 0100
// amendment 2026-10-07): from the engine's own RevealCostOptionsForEffect,
// so a payment this builds is one validateRevealLocked accepts. A beholding
// seat names a permanent it already shows the table when it has one — the
// options list ends with the permanents — and otherwise the first matching
// card in hand, because revealing keeps the card and the policy has
// nothing to choose between two of them. ok is false when the seat has no
// card to name, which drops the branch from the move list (it is also
// unpayable to the view).
func (e *enumerator) revealPayment(castID uuid.UUID, mandatory *game.AdditionalCost) (ids []uuid.UUID, ok bool) {
	if mandatory == nil || mandatory.Reveal == nil {
		return nil, true
	}
	options := e.g.RevealCostOptionsForEffect(e.seat, castID, mandatory.Reveal)
	if len(options) == 0 {
		return nil, false
	}
	if mandatory.Reveal.Behold {
		return []uuid.UUID{options[len(options)-1]}, true
	}
	return []uuid.UUID{options[0]}, true
}
