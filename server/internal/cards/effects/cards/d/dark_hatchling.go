package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dark Hatchling — Creature — Horror {4}{B}{B}, 3/3 (Urza's Saga):
//
//	"Flying
//	 When this creature enters, destroy target nonblack creature. It
//	 can't be regenerated."
//
// Big Game Hunter's enters trigger with Doom Blade's target. The target
// is judged as the trigger goes on the stack and again as it resolves
// (CR 608.2b), so a creature that turned black in response is spared, and
// a regeneration shield on the target is ignored (CR 701.19c). With no
// nonblack creature on the battlefield the trigger is removed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "36633b43-855b-4620-8afb-70c39fc07280",
		Name:            "Dark Hatchling",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Dark Hatchling — destroy target nonblack creature",
					func(g *game.Game, item *game.StackItem) error {
						return effects.DestroyTheTargetPermanentNoRegen(item, effects.NewContext(g, item))
					}),
				effects.TargetCreature("target nonblack creature", effects.NonBlack())),
		},
	})
}
