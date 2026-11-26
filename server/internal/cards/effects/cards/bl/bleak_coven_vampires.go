package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bleak Coven Vampires — Creature — Vampire Warrior {3}{B}{B}, 4/3 (Scars of
// Mirrodin, S82 OLD batch 469, #5577):
//
//	"Metalcraft — When this creature enters, if you control three or more
//	 artifacts, target player loses 4 life and you gain 4 life."
//
// Syphon Life's drain (TargetPlayerLosesLifeYouGain) as a targeted enter
// trigger on the metalcraft gate. A target that has left the game gains the
// controller nothing either (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "698f3526-489e-47a1-ac10-70cca73dea10",
		Name:         "Bleak Coven Vampires",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.OldB469WhenThisEntersIfMetalcraft("Bleak Coven Vampires — target player loses 4 life and you gain 4 life",
					func(g *game.Game, item *game.StackItem) error {
						return effects.TargetPlayerLosesLifeYouGain(4)(item, effects.NewContext(g, item))
					}),
				effects.TargetPlayer("target player"),
			),
		},
	})
}
