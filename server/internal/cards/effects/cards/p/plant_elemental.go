package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Plant Elemental — Creature — Plant Elemental {1}{G}, 3/4 (Portal, S82 OLD
// batch 783, #5630):
//
//	"When this creature enters, sacrifice it unless you sacrifice a Forest."
//
// Scythe Tiger's tax over Forests: the controller sacrifices a Forest of their
// choice, or the Elemental; with no Forest to sacrifice it goes.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e822bf3d-3a29-4a02-9ae8-e2830ce70f15",
		Name:         "Plant Elemental",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Plant Elemental — sacrifice it unless you sacrifice a Forest",
				effects.SacrificeThisUnlessYouSacrificeA("Plant Elemental", "a Forest", effects.Subtype("Forest"), effects.S82OldB783AnswerSacrificeAForestOrThis)),
		},
	})
}
