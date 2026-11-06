package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bushy Bodyguard — Creature — Squirrel Warrior {1}{G}, 2/1
// (Bloomburrow, #4467):
//
//	"Offspring {2} (You may pay an additional {2} as you cast this
//	 spell. If you do, when this creature enters, create a 1/1 token
//	 copy of it.)
//	 When this creature enters, you may forage. If you do, put two
//	 +1/+1 counters on it. (To forage, exile three cards from your
//	 graveyard or sacrifice a Food.)"
//
// Offspring is the shared keyword (Offspring, OffspringToken). The enters
// trigger is MayForage (#5184): asked only when the controller can
// forage, and the counters go on once the forage is paid. The token copy
// has the trigger too, so it may forage for its own counters. A
// Bodyguard that left and came back before the forage is a new object
// and gets nothing (AddCounter's source guard, CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "5436bb0d-b5c7-43a9-8b3f-67cdc0dd0381",
		Name:          "Bushy Bodyguard",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Offspring("{2}")},
		Triggered: []game.TriggeredAbility{
			OffspringToken("Bushy Bodyguard"),
			WhenThisEnters("Bushy Bodyguard — you may forage; if you do, put two +1/+1 counters on it",
				func(g *game.Game, item *game.StackItem) error {
					return MayForage(NewContext(g, item), "Bushy Bodyguard — forage to put two +1/+1 counters on it?", bushyBodyguardForaged)
				}),
		},
	})
}

// bushyBodyguardForaged is the paid half, keyed (#4264): two +1/+1
// counters on the creature whose trigger asked.
var bushyBodyguardForaged = PayUnlessBranches("pay-unless/bushy-bodyguard-counters",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return AddCounter{Target: ctx.Source(), Kind: game.CounterPlusOne, N: 2}.Apply(ctx)
	}, nil)
