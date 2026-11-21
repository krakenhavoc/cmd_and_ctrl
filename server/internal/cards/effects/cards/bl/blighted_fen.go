package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Blighted Fen — Land (Jumpstart 2022, #5948):
//
//	"{T}: Add {C}.
//	 {4}{B}, {T}, Sacrifice this land: Target opponent sacrifices a
//	 creature of their choice."
//
// Crude Bent Blade's edict body, as an activated ability of a land.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b8f3da11-7c8f-4846-98a6-204bfd8d572b",
		Name:         "Blighted Fen",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}{B}, {T}, Sacrifice this land: Target opponent sacrifices a creature of their choice.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.Plus(effects.ManaCost("{4}{B}"), effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetPlayer("target opponent", effects.Opponent()),
			Effect:  effects.TargetPlayerSacrificesACreature("Blighted Fen — sacrifice a creature"),
		}},
	})
}
