package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mortal Wound — Enchantment — Aura {G} (Visions, S82 older sets batch 790,
// #5631):
//
//	"Enchant creature
//	 When enchanted creature is dealt damage, destroy it."
//
// Mire Blight's damage watcher on the enchanted creature (any positive damage,
// combat or not), through the shared helper: the creature is named, with its
// object epoch, as the trigger goes on the stack, and destroyed only if it is
// still that creature on the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f7b70077-5518-45f9-9833-01a4f8afac71",
		Name:         "Mortal Wound",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.EnchantCreature(),
		Triggered:    []game.TriggeredAbility{effects.DestroyItWhenAttachedCreatureIsDealtDamage("Mortal Wound — destroy enchanted creature")},
	})
}
