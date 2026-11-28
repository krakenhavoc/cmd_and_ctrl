package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sunstone — Artifact {3} (Ice Age, S82 OLD #5636):
//
//	"{2}, Sacrifice a snow land: Prevent all combat damage that would be dealt
//	 this turn."
//
// The Fog shape (CR 615.1) with a snow land sacrificed as a cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "300ad0f7-92a7-450d-a9de-b658836fe3cb",
		Name:         "Sunstone",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, Sacrifice a snow land: Prevent all combat damage that would be dealt this turn.",
			Purpose: game.Purpose{Answers: game.AnswerPrevent},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.SacrificeN(1, "a snow land", effects.Land(), effects.Snow())),
			Effect:  effects.PreventAllCombatDamageB831("Sunstone"),
		}},
	})
}
