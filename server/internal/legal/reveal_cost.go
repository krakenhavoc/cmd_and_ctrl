package legal

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// revealPayment is the one card a reveal / behold cost names (ADR 0100
// amendment 2026-10-07): from the engine's own RevealCostOptionsForEffect,
// so a payment this builds is one validateRevealLocked accepts. The cost
// is the mandatory branch's, or — #4642 — an announced optional behold's
// ("you may behold a Hero"). A beholding seat names a permanent it already
// shows the table when it has one — the options list ends with the
// permanents — and otherwise the first matching card in hand, because
// revealing keeps the card and the policy has nothing to choose between
// two of them. ok is false when the seat has no card to name, which drops
// the branch, or the announced set that includes the optional behold, from
// the move list (it is also unpayable to the view); the unbeholding set is
// walked on its own and stays.
func (e *enumerator) revealPayment(castID uuid.UUID, mandatory *game.AdditionalCost, optional []game.AdditionalCost, chosen []int) (ids []uuid.UUID, ok bool) {
	rc := revealCostOf(mandatory, optional, chosen)
	if rc == nil {
		return nil, true
	}
	options := e.g.RevealCostOptionsForEffect(e.seat, castID, rc)
	if len(options) == 0 {
		return nil, false
	}
	if rc.Behold {
		return []uuid.UUID{options[len(options)-1]}, true
	}
	return []uuid.UUID{options[0]}, true
}

// revealCostOf is the reveal component this announcement pays: the
// mandatory cost's, else the first announced optional cost's. Register
// keeps a card from declaring both (checkOptionalBehold).
func revealCostOf(mandatory *game.AdditionalCost, optional []game.AdditionalCost, chosen []int) *game.RevealCost {
	if mandatory != nil && mandatory.Reveal != nil {
		return mandatory.Reveal
	}
	for _, i := range chosen {
		if i >= 0 && i < len(optional) && optional[i].Reveal != nil {
			return optional[i].Reveal
		}
	}
	return nil
}
