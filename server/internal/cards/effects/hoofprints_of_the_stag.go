package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hoofprints of the Stag — Kindred Enchantment — Elemental {1}{W}
// (Bloomburrow Commander, #5368):
//
//	"Whenever you draw a card, you may put a hoofprint counter on this
//	 enchantment.
//	 {2}{W}, Remove four hoofprint counters from this enchantment: Create
//	 a 4/4 white Elemental creature token with flying. Activate only
//	 during your turn."
//
// Quest for Pure Flame's shape: a "may" counter trigger (one per card
// drawn, each its own choice) and an activated ability that pays its
// counters as a cost (CR 602.2). The window is DuringYourTurn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "17739f55-486b-46ea-8785-70cc2bd2d872",
		Name:         "Hoofprints of the Stag",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(WheneverYouDraw("Hoofprints of the Stag — put a hoofprint counter",
				Do(CounterOnThis{Kind: "hoofprint", N: 1})),
				"Hoofprints of the Stag — put a hoofprint counter on it?"),
		},
		Activated: []ActivatedAbility{{
			Label:     "{2}{W}, Remove four hoofprint counters from this enchantment: Create a 4/4 white Elemental creature token with flying. Activate only during your turn.",
			Cost:      Plus(ManaCost("{2}{W}"), RemoveCountersFromThis("hoofprint", 4)),
			Condition: DuringYourTurn(),
			Purpose:   game.Purpose{Tokens: 1},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return CreateToken{Controller: item.Controller, Template: TokenCard("4/4 white Elemental with flying"), N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}
