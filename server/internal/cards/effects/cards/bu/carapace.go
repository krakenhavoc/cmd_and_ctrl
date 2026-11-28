package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Carapace — Enchantment — Aura {G} (Homelands, S82 OLD b825, #5635):
//
//	"Enchant creature
//	 Enchanted creature gets +0/+2.
//	 Sacrifice this Aura: Regenerate enchanted creature."
//
// +0/+2 is a layer 7c modify. The Aura is sacrificed to pay for the ability,
// so "enchanted creature" is the creature it was attached to as it last
// existed (CR 608.2h); one that has left the battlefield by resolution gets no
// shield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b893b554-9156-4d79-aca8-6e09c523760b",
		Name:         "Carapace",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.PumpAttached(0, 2)},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this Aura: Regenerate enchanted creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.SacrificeThis(),
			Effect:  effects.RegenerateTheEnchantedCreatureAsItLastExisted,
		}},
	})
}
