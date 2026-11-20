package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "f2b284cd-fa12-4ba1-b297-80105bb48fde",
		Name:            "Hobblefiend",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Activated: []ActivatedAbility{{
			Label:   "{1}, Sacrifice another creature: Put a +1/+1 counter on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet, SourceCounters: 1},
			Cost:    Plus(ManaCost("{1}"), SacrificeAnotherN(1, "another creature", Creature())),
			Effect:  putACounterOnThis(game.CounterPlusOne),
		}},
	})
}
