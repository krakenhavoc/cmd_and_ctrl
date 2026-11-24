package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Workshop Assistant — Artifact Creature — Construct {3}, 1/2 (Commander
// Legends):
//
//	"When this creature dies, return another target artifact card from your
//	 graveyard to your hand."
//
// A dies trigger whose target is chosen as it goes on the stack from the
// controller's graveyard; "another" excludes the Assistant itself, which is
// in the graveyard by then, and the target is re-checked on resolution
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "319ee221-bace-41e9-b386-6b528bd780c1",
		Name:         "Workshop Assistant",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisDies("Workshop Assistant — return another target artifact card from your graveyard to your hand", effects.ReturnTargetCardToHand),
				effects.Another(effects.TargetCardInGraveyard("another target artifact card from your graveyard", effects.Artifact(), effects.YouOwn()))),
		},
	})
}
