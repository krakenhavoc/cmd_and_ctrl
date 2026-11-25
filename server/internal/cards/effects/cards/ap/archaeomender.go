package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Archaeomender — Creature — Human Wizard {2}{U}, 2/3 (Jumpstart, #5510):
//
//	"When this creature enters, return target artifact card from your
//	 graveyard to your hand."
//
// Oltec Archaeologists' first mode as the whole ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5a4553fc-f945-4094-b231-2d378ed6609e",
		Name:         "Archaeomender",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Archaeomender — return target artifact card from your graveyard to your hand",
					func(g *game.Game, item *game.StackItem) error {
						return effects.ReturnTargetFromGraveyardToHand(item, effects.NewContext(g, item))
					}),
				effects.TargetCardInGraveyard("target artifact card from your graveyard", effects.Artifact(), effects.YouOwn())),
		},
	})
}
