package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aku Djinn — Creature — Djinn {3}{B}{B}, 5/6 (Visions, #5631):
//
//	"Trample
//	 At the beginning of your upkeep, put a +1/+1 counter on each creature
//	 each opponent controls."
//
// The set of creatures is read as the trigger resolves, before the first
// counter goes on (Toxrill's shape), every creature a player other than the
// Djinn's controller controls.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "233c6686-ead8-4569-941d-3bfa768ff270",
		Name:            "Aku Djinn",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Aku Djinn — put a +1/+1 counter on each creature each opponent controls",
				func(g *game.Game, item *game.StackItem) error {
					return effects.B22PutCounterOnEachCreatureYouDontControl(g, item, game.CounterPlusOne)
				}),
		},
	})
}
