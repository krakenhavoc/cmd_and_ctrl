package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elvish Healer — Creature — Elf Cleric {2}{W}, 1/2 (Ice Age, S82 OLD, #5636):
//
//	"{T}: Prevent the next 1 damage that would be dealt to any target this turn.
//	 If it's a green creature, prevent the next 2 damage instead."
//
// Benevolent Ancestor's shield, sized as it resolves by the target's colour
// (CR 615.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ea46835c-9dac-4e1e-8338-ad99136f511a",
		Name:         "Elvish Healer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Prevent the next 1 damage that would be dealt to any target this turn. If it's a green creature, prevent the next 2 damage instead.",
			Purpose: game.Purpose{Answers: game.AnswerPrevent},
			Cost:    effects.TapCost(),
			Targets: effects.TargetAny(),
			Effect:  effects.OldB833PreventNextToTheTarget("Elvish Healer"),
		}},
	})
}
