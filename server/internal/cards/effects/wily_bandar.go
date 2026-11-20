package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wily Bandar — Creature — Cat Monkey {G}, 1/1 (Kaladesh, #5539):
//
//	"{2}{G}: This creature gains indestructible until end of turn.
//	 (Damage and effects that say "destroy" don't destroy it.)"
//
// Femeref Knight's self grant: a layer-6 indestructible pinned to this object
// until the turn ends (CR 613.1f, CR 514.2) through the shared
// thisGainsKeywordUntilEndOfTurn body.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2b0ba5a5-649a-4d7a-8136-6bcf692337f5",
		Name:         "Wily Bandar",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}{G}: This creature gains indestructible until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    ManaCost("{2}{G}"),
			Effect:  thisGainsKeywordUntilEndOfTurn("indestructible", "Wily Bandar — indestructible until end of turn"),
		}},
	})
}
