package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hobblefiend — Creature — Devil {1}{R}, 2/1 (Core Set 2021, #5511):
//
//	"Trample
//	 {1}, Sacrifice another creature: Put a +1/+1 counter on this creature."
//
// Hungry Ghoul's sacrifice outlet (SacrificeAnotherN, so it can't
// sacrifice itself) with trample printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f2b284cd-fa12-4ba1-b297-80105bb48fde",
		Name:            "Hobblefiend",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, Sacrifice another creature: Put a +1/+1 counter on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet, SourceCounters: 1},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.SacrificeAnotherN(1, "another creature", effects.Creature())),
			Effect:  effects.PutACounterOnThis(game.CounterPlusOne),
		}},
	})
}
