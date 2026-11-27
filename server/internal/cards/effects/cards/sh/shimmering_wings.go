package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shimmering Wings — Enchantment — Aura {U} (Invasion, S82 older sets batch 661, #5616):
//
//	"Enchant creature
//	 Enchanted creature has flying.
//	 {U}: Return this Aura to its owner's hand."
//
// Conviction's return ability with a static on the enchanted creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7d892d61-5a69-4500-9fb2-7cf50728aedd",
		Name:         "Shimmering Wings",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.GrantToAttached("flying")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}: Return this Aura to its owner's hand.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.ManaCost("{U}"),
			Effect:  effects.Do(effects.Instruction(effects.StdB9ReturnThisToHand)),
		}},
	})
}
