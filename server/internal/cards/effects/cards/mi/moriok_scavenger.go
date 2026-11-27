package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Moriok Scavenger — Creature — Human Rogue {3}{B}, 2/3 (Mirrodin, #5606):
//
//	"When this creature enters, you may return target artifact creature card
//	 from your graveyard to your hand."
//
// Midnight Scavengers' optional enters trigger with an artifact-creature
// clause.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fdccdd9f-6bf9-4339-ad7a-e9dd8f771b22",
		Name:         "Moriok Scavenger",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(effects.WhenThisEnters("Moriok Scavenger — you may return target artifact creature card from your graveyard to your hand", effects.ReturnTargetedCardToHand),
				effects.TargetCardInGraveyard("target artifact creature card from your graveyard", effects.Artifact(), effects.Creature(), effects.YouOwn())),
				"Moriok Scavenger — return the artifact creature card to your hand?"),
		},
	})
}
