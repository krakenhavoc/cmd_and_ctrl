package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mark of Fury — {R} Enchantment — Aura (Urza's Destiny, S82 OLD batch 703, #5621):
//
//	"Enchant creature
//	 Enchanted creature has haste.
//	 At the beginning of the end step, return this Aura to its owner's hand."
//
// The end-step trigger fires in every turn's end step (Ball Lightning's
// "each end step"), and does nothing if the Aura has already left.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1c350891-8810-4c14-b8ea-84486864637b",
		Name:         "Mark of Fury",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.GrantToAttached("haste")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginEndStep, effects.AnyPlayer, "Mark of Fury — return this Aura to its owner's hand",
				effects.Do(effects.Instruction(effects.StdB9ReturnThisToHand))),
		},
	})
}
