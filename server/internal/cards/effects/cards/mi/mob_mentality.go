package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mob Mentality — Enchantment — Aura {R} (Visions, S82 older sets batch 790,
// #5631):
//
//	"Enchant creature
//	 Enchanted creature has trample.
//	 Whenever all non-Wall creatures you control attack, enchanted creature
//	 gets +X/+0 until end of turn, where X is the number of attacking
//	 creatures."
//
// One trigger per declaration (Master's Guidance's batch): it fires when the
// Aura's controller attacks with every non-Wall creature they control, and at
// least one. X is counted as the trigger resolves, over every attacking
// creature; the enchanted creature need not be attacking itself.
//
// No simplification.
func init() {
	const label = "Mob Mentality — enchanted creature gets +X/+0, X the number of attacking creatures"
	effects.Register(effects.Spec{
		OracleID:     "90e8fb2c-6473-42c3-92f4-f46a744920fa",
		Name:         "Mob Mentality",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.GrantToAttached("trample")},
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.On(game.EventAttack, effects.Old790AllNonWallCreaturesYouControlAttacked, label,
				effects.Old790EnchantedGetsPowerPerAttacker("Mob Mentality — +X/+0"))),
		},
	})
}
