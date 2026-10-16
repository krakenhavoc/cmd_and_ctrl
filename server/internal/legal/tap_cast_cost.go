package legal

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tapCastPayment is the permanents a mandatory "tap an untapped Vampire
// you control" additional cost names (#3868, New Blood): the engine's own
// candidate walk (TapOthersOptionsForEffect), in the policy's
// cheapest-to-tap order (TargetCandidate.Tap, the order Orim's Cure's
// creature is picked in), so a payment this builds is one
// validateTapOthersCostLocked accepts. ONE payment, the first Count of
// that order: the cast has nothing else to choose between two Vampires,
// as teamworkBlightPayment's discipline has it. They ride the move's
// teamwork_ids, so the affordability check keeps them out of the mana
// (CastAutoTapExclusions). ok is false when the seat cannot pay, which
// drops the cast (CR 118.3); the view ships the same picker short.
func (e *enumerator) tapCastPayment(castID uuid.UUID, mandatory *game.AdditionalCost) (ids []uuid.UUID, ok bool) {
	if mandatory == nil || mandatory.TapOthers.Empty() {
		return nil, true
	}
	tc := mandatory.TapOthers
	pool := e.cheapestTapFirst(e.g.TapOthersOptionsForEffect(e.seat, castID, tc))
	if len(pool) < tc.Count {
		return nil, false
	}
	return append([]uuid.UUID(nil), pool[:tc.Count]...), true
}
