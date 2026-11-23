package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thermal Navigator — Artifact Creature — Construct {3}, 2/2 (Fifth Dawn, #5604):
//
//	"Sacrifice an artifact: This creature gains flying until end of
//	 turn."
//
// Dross Hopper's flying grant paid with Atog's "sacrifice an artifact".
// The Navigator is an artifact, so it may sacrifice itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "355c2840-2cb8-431a-bcdc-f4b8824a1f5d",
		Name:         "Thermal Navigator",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice an artifact: This creature gains flying until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant | game.AnswerSacOutlet},
			Cost:    effects.SacrificeN(1, "an artifact", effects.Artifact()),
			Effect:  effects.ThisGainsKeywordUntilEndOfTurn("flying", "Thermal Navigator — flying until end of turn"),
		}},
	})
}
