package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Salvation Engine — Artifact — Vehicle {4}{W}, 6/10 (Aetherdrift, #4464):
//
//	"Other artifact creatures you control get +2/+2.
//	 Whenever this Vehicle attacks, return up to one target artifact card
//	 from your graveyard to the battlefield.
//	 Crew 6"
//
// Chief of the Foundry's anthem at +2/+2 (post-layer types, so a crewed
// Vehicle or an animated artifact counts; the Engine itself never does),
// and Dr. Madison Li's "target artifact card from your graveyard" made up
// to one, returned untapped by returnFirstLegalGraveyardTargetToBattlefield
// (nothing if the card has left by resolution, CR 608.2b). Crew is the
// engine's (CR 702.122a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7903666e-7803-4cfc-9ed9-482a6b577c96",
		Name:         "Salvation Engine",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.B16Anthem(effects.B27OtherArtifactCreaturesYouControl, 2, 2),
		},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WheneverThisAttacks("Salvation Engine — return up to one target artifact card from your graveyard to the battlefield",
					effects.ReturnFirstLegalGraveyardTargetToBattlefield),
				effects.TargetCardInGraveyard("up to one target artifact card from your graveyard", effects.Artifact(), effects.YouOwn()).WithCount(0, 1)),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 6",
			Cost:    effects.CrewCost(6),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Salvation Engine"),
		}},
	})
}
