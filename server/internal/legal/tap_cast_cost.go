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
//
// An "any number" cost (#3961) is sized outside, by tapAnyPayments, so the
// count is priced with the cast: `tapAny` is that expansion's payment,
// returned as it is (none is a legal payment).
func (e *enumerator) tapCastPayment(castID uuid.UUID, mandatory *game.AdditionalCost, tapAny []uuid.UUID) (ids []uuid.UUID, ok bool) {
	if mandatory == nil || mandatory.TapOthers.Empty() {
		return nil, true
	}
	tc := mandatory.TapOthers
	if game.TapOthersAnyNumber(tc) {
		return tapAny, true
	}
	pool := e.cheapestTapFirst(e.g.TapOthersOptionsForEffect(e.seat, castID, tc))
	if len(pool) < tc.Count {
		return nil, false
	}
	return append([]uuid.UUID(nil), pool[:tc.Count]...), true
}

// tapAnyPayments is the payments a cast offers a mandatory "tap any number
// of untapped creatures you control" additional cost (#3961): Burn at the
// Stake's damage and Explosive Singularity's discount are both the count.
// [nil] for every card without one, so the caller's loop runs once and
// the enumeration is unchanged.
//
// The count is part of the announcement and changes the spell (its price,
// or what it does), so each count is its own move, priced on its own:
// none, then one, two, … creatures off the front of the engine's own
// candidate walk (game.TapOthersOptionsForEffect) in the policy's
// cheapest-to-tap order, up to the cost-payment cap, and then every
// option, the most the board can pay. The small counts are the cheapest
// way to make a discounted spell affordable; the full one is the most a
// "three times the number" spell can do. Which creatures of a given count
// is not worth a cross product: the policy has already said which it
// would rather tap (handExilePayments' discipline). The policy prices
// each tap and what the count buys, and picks.
func (e *enumerator) tapAnyPayments(card game.Card) [][]uuid.UUID {
	ac := game.AdditionalCostFor(game.CatalogKey(card))
	if ac == nil || !game.TapOthersAnyNumber(ac.TapOthers) {
		return [][]uuid.UUID{nil}
	}
	pool := e.cheapestTapFirst(e.g.TapOthersOptionsForEffect(e.seat, card.InstanceID, ac.TapOthers))
	n := min(len(pool), e.capOr(maxEnumeratedCostPayments))
	out := make([][]uuid.UUID, 0, n+2)
	out = append(out, nil)
	for k := 1; k <= n; k++ {
		out = append(out, append([]uuid.UUID(nil), pool[:k]...))
	}
	if len(pool) > n {
		out = append(out, append([]uuid.UUID(nil), pool...))
		if left := len(pool) - n - 1; left > 0 {
			e.noteCut(CapCostPayments, left, false)
		}
	}
	return out
}
