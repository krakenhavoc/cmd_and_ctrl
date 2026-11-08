package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Training Grounds — {U} Enchantment (Avatar: The Last Airbender
// Eternal, #5075):
//
//	"Activated abilities of creatures you control cost {2} less to
//	 activate. This effect can't reduce the mana in that cost to less
//	 than one mana."
//
// An activation-scoped board reduction (ActivationCostsLess) with the
// one-mana floor (KeepsOneMana, #302). The rulings it follows, each
// the engine's and not this file's:
//
//   - generic mana only: {2}{G} costs {G}, and {R}{R} or a {T}-only
//     cost is untouched (CR 118.7a);
//   - at least one mana: {2} costs {1}, and a {1} or {0} cost is never
//     raised to {1};
//   - the TOTAL cost, increases included: Suppression Field's {2} on an
//     {R} ability is taken back off, and {X} at its announced value is
//     part of it, so Drana's {X}{B}{B} at X = 5 costs {3}{B}{B};
//   - creatures on the battlefield only: an ability that works from
//     another zone (cycling, unearth) is not one of a creature you
//     control. Mana abilities are activated abilities (CR 605.1a), so
//     theirs are included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8de3eabe-9a8f-4865-beef-a4888c081d39",
		Name:         "Training Grounds",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			KeepsOneMana(ActivationCostsLess(2,
				"Activated abilities of creatures you control cost {2} less to activate. This effect can't reduce the mana in that cost to less than one mana.",
				OfACreatureYouControl())),
		},
	})
}
