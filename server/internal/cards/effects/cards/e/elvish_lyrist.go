package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elvish Lyrist — Creature — Elf {G}, 1/1 (Eighth Edition, #5607):
//
//	"{G}, {T}, Sacrifice this creature: Destroy target enchantment."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e0de1c9b-f71f-4498-a968-d3c635cb4c5c",
		Name:         "Elvish Lyrist",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{G}, {T}, Sacrifice this creature: Destroy target enchantment.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.Plus(effects.ManaCost("{G}"), effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetPermanent("target enchantment", effects.Enchantment()),
			Effect:  effects.DestroyFirstLegalTarget,
		}},
	})
}
