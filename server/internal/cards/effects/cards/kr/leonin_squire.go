package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Leonin Squire — Creature — Cat Soldier {1}{W}, 2/2 (Fifth Dawn, #5604):
//
//	"When this creature enters, return target artifact card with mana
//	 value 1 or less from your graveyard to your hand."
//
// Auriok Salvagers' target, as an enters trigger. With no such card in the
// graveyard the trigger is removed (CR 603.3d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fa19c9d7-fe93-4e93-a5ef-9af9c6cd1a5e",
		Name:         "Leonin Squire",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Leonin Squire — return an artifact card with mana value 1 or less to your hand",
					effects.ReturnFirstLegalGraveyardTargetToHand),
				effects.TargetCardInGraveyard("target artifact card with mana value 1 or less in your graveyard",
					effects.Artifact(), effects.ManaValueLE(1), effects.YouOwn()),
			),
		},
	})
}
