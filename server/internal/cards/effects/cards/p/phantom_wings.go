package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Phantom Wings — Enchantment — Aura {1}{U} (Betrayers of Kamigawa, S82 OLD b594):
//
//	"Enchant creature
//	 Enchanted creature has flying.
//	 Sacrifice this Aura: Return enchanted creature to its owner's hand."
//
// A flying grant and Spiral into Solitude's Aura ability: the cost sacrifices the Aura and
// "enchanted creature" is the creature it was attached to as it left (CR 608.2h), so the
// bounce still finds it. It does not target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9da74f0f-143d-4db2-81eb-c7873744a2d0",
		Name:         "Phantom Wings",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Static:       []game.StaticAbility{effects.GrantToAttached("flying")},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this Aura: Return enchanted creature to its owner's hand.",
			Cost:    effects.SacrificeThis(),
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Effect:  effects.ReturnEnchantedCreatureToHand,
		}},
	})
}
