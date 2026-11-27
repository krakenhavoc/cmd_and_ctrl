package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Monk Realist — Creature — Human Monk Cleric {1}{W}, 1/1 (Commander 2011, S82
// older sets batch 453, #5574):
//
//	"When this creature enters, destroy target enchantment."
//
// Nikko-Onna's enters trigger, which names a target enchantment (CR 603.3d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8df64d89-64ee-4388-826d-dd091db4510a",
		Name:         "Monk Realist",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Monk Realist — destroy target enchantment", effects.DestroyFirstLegalCardTarget),
				effects.TargetPermanent("target enchantment", effects.Enchantment())),
		},
	})
}
