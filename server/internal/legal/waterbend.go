package legal

import (
	"sort"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// waterbend.go — #1310 / #1311: the enumerator's answer to "which
// permanents does this waterbend payment tap" (CR 701.67a), for an
// activated ability's "Waterbend {N}:" and for a "Ward—Waterbend {N}"
// pay-unless.
//
// ONE payment, not every subset — the discipline crewPayment and
// the discard payment follow (#544: a cost expansion must not spend the
// per-source budget the target loop needs). The candidates come from
// game.WaterbendOptionsForEffect, the walk the engine validates
// against, and the tapped set is excluded from the affordability check
// exactly as the engine excludes it from the auto-tapper, so an offered
// payment is one the engine accepts.
//
// Which permanents, in which order:
//
//  1. FREE ones first — an artifact that is neither a creature nor a
//     mana source. Tapping it costs the seat nothing it could have
//     used this turn, so the payment always takes as many of these as
//     the budget allows.
//  2. Creatures that make no mana. Each one tapped is a blocker or an
//     attacker spent, so the payment takes only as many as it needs
//     to become affordable.
//  3. Mana sources last. Tapping a Birds of Paradise for {1} of
//     waterbend and tapping it for {G} towards the rest pay the same
//     single mana, so a mana source adds nothing a plan could not
//     already find — it is here only so a board of nothing but mana
//     creatures can still pay.
//
// Within each tier the seat's own fuel price orders the pool
// (cheapestFuelFirst), so a policy that values a creature keeps it.

// waterbendTier ranks a candidate for the order above.
func waterbendTier(g *game.Game, id uuid.UUID) int {
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.InstanceID != id {
			continue
		}
		manaSource := len(game.ManaAbilitiesForCard(*c)) > 0
		switch {
		case manaSource:
			return 2
		case c.IsCreature():
			return 1
		default:
			return 0
		}
	}
	return 2
}

// waterbendOrder is `options` in payment order, and how many of them
// are free (tier 0). The returned slice is fresh.
func (e *enumerator) waterbendOrder(options []uuid.UUID) ([]uuid.UUID, int) {
	ordered := e.cheapestFuelFirst(options)
	ordered = append([]uuid.UUID(nil), ordered...)
	tier := make(map[uuid.UUID]int, len(ordered))
	free := 0
	for _, id := range ordered {
		tier[id] = waterbendTier(e.g, id)
		if tier[id] == 0 {
			free++
		}
	}
	sort.SliceStable(ordered, func(i, j int) bool { return tier[ordered[i]] < tier[ordered[j]] })
	return ordered, free
}

// waterbendPayment is the smallest prefix of the payment order, taking
// every free permanent the budget allows, for which `payable` holds —
// or false when no prefix up to the budget is payable, and the payment
// is not offered at all (#544).
func (e *enumerator) waterbendPayment(options []uuid.UUID, budget int, payable func(taps []uuid.UUID) bool) ([]uuid.UUID, bool) {
	ordered, free := e.waterbendOrder(options)
	hi := budget
	if hi > len(ordered) {
		hi = len(ordered)
	}
	lo := free
	if lo > hi {
		lo = hi
	}
	for k := lo; k <= hi; k++ {
		taps := ordered[:k]
		if payable(taps) {
			return append([]uuid.UUID(nil), taps...), true
		}
	}
	return nil, false
}

// waterbendAbilityPayment solves an activated ability whose cost
// carries a waterbend clause (#1310): the X it announces (the LARGEST
// affordable at or above `floor`, as affordableXExcluding picks for a
// mana {X}) and the permanents it taps.
//
// `cost` is the PRICED mana component (AbilityManaCostForEffect),
// `excluded` the engine's base auto-tap exclusions. For each X the
// widest payment is tried first to learn whether X is reachable at
// all — tapping more never makes a cost harder to pay — and the scan
// stops at the first X that is not, because the cost is monotonic in
// X. The payment returned for the chosen X is then the smallest one
// waterbendPayment finds.
func (e *enumerator) waterbendAbilityPayment(
	source *game.Card,
	ab game.ActivatedAbilityShape,
	cost game.ParsedCost,
	floor int,
	excluded map[uuid.UUID]bool,
) (int, []uuid.UUID, bool) {
	wb := ab.Cost.Waterbend
	options := e.g.WaterbendOptionsForEffect(e.seat, game.AbilityWaterbendExclusion(source.InstanceID, ab.Cost), wb)
	spend := game.ManaSpendForAbility(*source)
	payableAt := func(x int) func(taps []uuid.UUID) bool {
		return func(taps []uuid.UUID) bool {
			return e.canPayExcluding(game.WaterbendReduced(cost, x, len(taps)), 0, spend,
				game.WithAutoTapExclusions(excluded, taps))
		}
	}
	solve := func(x int) ([]uuid.UUID, bool) {
		return e.waterbendPayment(options, game.WaterbendBudget(wb, cost, x), payableAt(x))
	}
	if cost.XSlots == 0 {
		taps, ok := solve(0)
		return 0, taps, ok
	}
	if floor < 0 {
		floor = 0
	}
	best, ok := -1, false
	for x := floor; x <= e.opts.MaxX; x++ {
		budget := game.WaterbendBudget(wb, cost, x)
		widest := options
		if len(widest) > budget {
			ordered, _ := e.waterbendOrder(options)
			widest = ordered[:budget]
		}
		if !payableAt(x)(widest) {
			break
		}
		best, ok = x, true
	}
	if !ok {
		return 0, nil, false
	}
	taps, found := solve(best)
	return best, taps, found
}

// waterbendPayUnlessPayment is the pay_unless half (#1311): whether the
// chooser can pay the prompt's cost with some taps plus mana, and with
// which taps. Zero spend context, matching payCostLocked — a
// pay-unless cost is neither a cast nor an activation (#352).
func (e *enumerator) waterbendPayUnlessPayment(c *game.PendingChoice, cost game.ParsedCost) ([]uuid.UUID, bool) {
	tc := c.PayTapCost()
	options := e.g.WaterbendOptionsForEffect(e.seat, uuid.Nil, tc)
	return e.waterbendPayment(options, game.WaterbendBudget(tc, cost, 0), func(taps []uuid.UUID) bool {
		return e.canPayExcluding(game.WaterbendReduced(cost, 0, len(taps)), 0, game.ManaSpendContext{},
			game.WithAutoTapExclusions(nil, taps))
	})
}
