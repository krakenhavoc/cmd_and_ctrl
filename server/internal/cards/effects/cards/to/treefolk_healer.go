package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Treefolk Healer (Invasion, S82 OLD batch 647, #5616):
//
//	Creature — Treefolk Cleric {4}{G}, 2/3:
//	"{2}{W}, {T}: Prevent the next 2 damage that would be dealt to any target this turn."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a5ef15b5-83d2-4b23-9403-a9a52be46154",
		Name:         "Treefolk Healer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{W}, {T}: Prevent the next 2 damage that would be dealt to any target this turn.",
			Purpose: game.Purpose{Answers: game.AnswerPrevent},
			Cost:    effects.Plus(effects.ManaCost("{2}{W}"), effects.TapCost()),
			Targets: effects.TargetAny(),
			Effect:  effects.PreventNextDamageToTheTarget("Treefolk Healer", 2),
		}},
	})
}
