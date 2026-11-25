package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Salvager of Ruin — Artifact Creature — Construct {3}, 2/1 (Core Set 2020):
//
//	"Sacrifice this creature: Choose target permanent card in your graveyard
//	 that was put there from the battlefield this turn. Return it to your
//	 hand."
//
// Brought Back's target clause: a permanent card in your graveyard that this
// turn's record says went there from the battlefield. Targets are chosen
// before the sacrifice is paid (CR 602.2b), so the Salvager itself, still on
// the battlefield, is not a legal choice. The card returns to its owner's
// hand; a target that left the graveyard in response does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c6dc70de-4fa5-466b-af28-0ced885d4e12",
		Name:         "Salvager of Ruin",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label: "Sacrifice this creature: Choose target permanent card in your graveyard that was put there from the battlefield this turn. Return it to your hand.",
			Cost:  effects.SacrificeThis(),
			Targets: effects.TargetCardInGraveyard("target permanent card in your graveyard that was put there from the battlefield this turn",
				effects.YouOwn(), effects.Permanent(), effects.PutThereFromTheBattlefieldThisTurn),
			Effect: effects.ReturnFirstLegalGraveyardTargetToHand,
		}},
	})
}
