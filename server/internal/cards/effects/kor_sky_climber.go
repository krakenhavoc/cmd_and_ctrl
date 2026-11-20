package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kor Sky Climber — Creature — Kor Soldier Ally {2}{W}, 3/2 (Oath of the Gatewatch, #5545):
//
//	"{1}{W}: This creature gains flying until end of turn."
//
// The flying is a layer-6 grant to this creature until cleanup.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "298a71d4-6dc1-444e-9057-0d25aef7fb64",
		Name:         "Kor Sky Climber",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}{W}: This creature gains flying until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    ManaCost("{1}{W}"),
			Effect:  thisGainsKeywordUntilEndOfTurn("flying", "Kor Sky Climber — flying until end of turn"),
		}},
	})
}
