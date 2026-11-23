package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Trusty Packbeast — Creature — Beast {2}{W}, 2/3 (Core Set 2019, #5527):
//
//	"When this creature enters, return target artifact card from your graveyard to your hand."
//
// Loran's graveyard target, as this creature's own enters trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "12c7289f-da53-403a-a607-227f43c7e171",
		Name:         "Trusty Packbeast",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Trusty Packbeast — return target artifact card from your graveyard to your hand",
				effects.ReturnTargetGraveyardCardToHand),
				effects.TargetCardInGraveyard("target artifact card from your graveyard", effects.YouOwn(), effects.Artifact())),
		},
	})
}
