package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mire Blight — Enchantment — Aura {B} (Jumpstart 2022):
//
//	"Enchant creature
//	 When enchanted creature is dealt damage, destroy it."
//
// Pain for All's damage watcher on the enchanted creature (any positive
// damage, combat or not). The creature is named as the trigger goes on the
// stack, so "it" still names the host if the Aura is gone by resolution;
// regeneration and indestructible apply as for any destroy.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fcfbe3b8-a955-45d8-9329-3af565a3b53c",
		Name:         "Mire Blight",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.EnchantCreature(),
		Triggered:    []game.TriggeredAbility{effects.DestroyItWhenAttachedCreatureIsDealtDamage("Mire Blight — destroy enchanted creature")},
	})
}
