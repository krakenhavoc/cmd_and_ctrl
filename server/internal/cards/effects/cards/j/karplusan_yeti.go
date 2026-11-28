package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Karplusan Yeti — Creature — Yeti {3}{R}{R}, 3/3 (Ice Age, S82 OLD, #5636):
//
//	"{T}: This creature deals damage equal to its power to target creature. That creature
//	 deals damage equal to its power to this creature."
//
// Tahngarth's mutual blows (S82OldB432MutualDamage): not a fight; both amounts are read as the
// ability resolves and dealt as one damage event. The {T} keeps it from the turn it arrives.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d6d84884-4a58-4187-aa5c-20b39b379590",
		Name:         "Karplusan Yeti",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: This creature deals damage equal to its power to target creature. That creature deals damage equal to its power to this creature.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.S82OldB432MutualDamage,
		}},
	})
}
