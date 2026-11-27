package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Heap Doll — Artifact Creature — Scarecrow {1}, 1/1 (Shadowmoor, #5591):
//
//	"Sacrifice this creature: Exile target card from a graveyard."
//
// The target is chosen as the ability is put on the stack (CR 602.2b); if
// the card is gone by resolution nothing happens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "aa7f0f15-bcbb-4af5-ad31-ff215dd81e71",
		Name:         "Heap Doll",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this creature: Exile target card from a graveyard.",
			Purpose: game.Purpose{Answers: game.AnswerGraveyard},
			Cost:    effects.SacrificeThis(),
			Targets: effects.TargetCardInGraveyard("target card from a graveyard"),
			Effect:  effects.ExileFirstLegalCardTarget,
		}},
	})
}
