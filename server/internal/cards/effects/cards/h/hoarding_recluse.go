package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hoarding Recluse — Creature — Spider {3}{G}, 2/3 (The Brothers' War,
// #5482):
//
//	"Reach, deathtouch
//	 When this creature dies, put up to one other target card from a
//	 graveyard on the bottom of its owner's library."
//
// "Other" is object identity (Another): the Recluse itself, now in the
// graveyard, is not a legal pick. Up to one, so the trigger needs no
// target and still resolves as a no-op.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5fc09704-56cf-4e47-bd96-27ac34434941",
		Name:            "Hoarding Recluse",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach", "deathtouch"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisDies("Hoarding Recluse — put up to one other target card from a graveyard on the bottom of its owner's library",
					effects.SocB37PutTargetGraveyardCardOnBottom),
				effects.Another(effects.TargetCardInGraveyard("up to one other target card from a graveyard").WithCount(0, 1)),
			),
		},
	})
}
