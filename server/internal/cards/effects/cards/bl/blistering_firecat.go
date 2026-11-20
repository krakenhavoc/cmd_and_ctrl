package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blistering Firecat — Creature — Elemental Cat {1}{R}{R}{R}, 7/1
// (Onslaught):
//
//	"Trample, haste
//	 At the beginning of the end step, sacrifice this creature.
//	 Morph {R}{R}"
//
// Frostbite Pyromental's end-step sacrifice: the END STEP of every
// turn, not only its controller's. Face down it is a 2/2 with no text
// (CR 708.2a), so it does not sacrifice itself while it hides; turned
// face up for {R}{R} it is the 7/1 with trample and haste, and the next
// end step takes it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "b3595359-0d40-495e-85da-6a56fbce7357",
		Name:             "Blistering Firecat",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"trample", "haste"},
		AlternativeCosts: []game.AlternativeCost{effects.Morph("{R}{R}")},
		Triggered: []game.TriggeredAbility{
			effects.AtEachStep(game.StepEnd, "Blistering Firecat — sacrifice it",
				func(g *game.Game, item *game.StackItem) error {
					return effects.SacrificePermanent{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
