package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "17739f55-486b-46ea-8785-70cc2bd2d872",
		Name:         "Hoofprints of the Stag",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.WheneverYouDraw("Hoofprints of the Stag — put a hoofprint counter",
				effects.Do(effects.CounterOnThis{Kind: "hoofprint", N: 1})),
				"Hoofprints of the Stag — put a hoofprint counter on it?"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:     "{2}{W}, Remove four hoofprint counters from this enchantment: Create a 4/4 white Elemental creature token with flying. Activate only during your turn.",
			Cost:      effects.Plus(effects.ManaCost("{2}{W}"), effects.RemoveCountersFromThis("hoofprint", 4)),
			Condition: effects.DuringYourTurn(),
			Purpose:   game.Purpose{Tokens: 1, Answers: game.AnswerValue},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.CreateToken{Controller: item.Controller, Template: effects.TokenCard("4/4 white Elemental with flying"), N: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
