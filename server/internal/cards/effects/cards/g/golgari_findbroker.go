package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Golgari Findbroker — Creature — Elf Shaman {B}{B}{G}{G}, 3/4 (Guilds of Ravnica, #5525):
//
//	"When this creature enters, return target permanent card from your
//	 graveyard to your hand."
//
// Elvish Regrower's enter trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e5456b35-d7a9-4717-90df-c6c46f1ec437",
		Name:         "Golgari Findbroker",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Golgari Findbroker — return target permanent card from your graveyard to your hand",
				effects.ReturnTargetGraveyardCardToHand),
				effects.TargetCardInGraveyard("target permanent card from your graveyard", effects.Permanent(), effects.YouOwn())),
		},
	})
}
