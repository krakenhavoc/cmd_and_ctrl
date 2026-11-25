package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// ACyclingAbilityCost matches a cycling ability (CR 702.29, typecycling
// included, CR 702.29e): Fluctuator's "Cycling abilities you activate cost {2}
// less to activate" (Commander 2020, S82 OLD, #5513). It reads the keyword bit
// both cycling constructors set (game.AbilityCostSubject.Cycling), never the
// label, so a cycling ability a static granted to a card in hand (Jo Grant) is
// matched too. Pair it with ActivatedByTheModifiersController for "you
// activate".
func ACyclingAbilityCost() CostPredicate {
	return func(q game.CostQuery) bool {
		return q.Ability != nil && q.Ability.Cycling
	}
}
