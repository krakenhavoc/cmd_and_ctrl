package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Baloth Null — Creature — Zombie Beast {4}{B}{G}, 4/5 (Oath of the
// Gatewatch, S82 OLD #5545):
//
//	"When this creature enters, return up to two target creature cards from
//	 your graveyard to your hand."
//
// Soul Salvage's sentence on an enters trigger: each target creature card
// still in the graveyard as it resolves returns to hand (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "55a0a40c-cb40-469d-9bf3-9c33b378f4f7",
		Name:         "Baloth Null",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Baloth Null — return up to two target creature cards from your graveyard to your hand",
					effects.Do(effects.Instruction(effects.ReturnLegalGraveyardTargetsToHand))),
				effects.TargetCardInGraveyard("up to two target creature cards from your graveyard", effects.YouOwn(), effects.Creature()).WithCount(0, 2)),
		},
	})
}
