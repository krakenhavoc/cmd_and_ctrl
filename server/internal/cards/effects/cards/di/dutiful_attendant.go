package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dutiful Attendant — Creature — Human Warrior {2}{B}, 1/2 (Jumpstart,
// #5510):
//
//	"When this creature dies, return another target creature card from your
//	 graveyard to your hand."
//
// Deadwood Treefolk's return clause on a dies trigger. The dying creature is
// itself in the graveyard when the trigger is put on the stack, so "another"
// excludes it (CR 603.3d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e9195f10-8cc2-482a-9e28-0be9a37099b4",
		Name:         "Dutiful Attendant",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisDies("Dutiful Attendant — return another target creature card from your graveyard to your hand",
					func(g *game.Game, item *game.StackItem) error {
						return effects.ReturnTargetFromGraveyardToHand(item, effects.NewContext(g, item))
					}),
				effects.Another(effects.TargetCardInGraveyard("another target creature card from your graveyard", effects.Creature(), effects.YouOwn()))),
		},
	})
}
